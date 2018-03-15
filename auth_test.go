package main

import (
	"errors"
	"fmt"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TemplateMock struct {
	ExecuteCount int
	ExecuteData  interface{}
	ExecuteError bool
}

func (tm *TemplateMock) Execute(wr io.Writer, data interface{}) error {
	tm.ExecuteCount++
	tm.ExecuteData = data
	if tm.ExecuteError {
		return fmt.Errorf("")
	}
	return nil
}

func TestInsertTokenToTemplateShouldExecuteOnTemplate(t *testing.T) {
	mock := TemplateMock{}

	insertTokenToTemplate("test token", &mock)

	if mock.ExecuteCount != 1 {
		t.Error("template.Execute should be called once. It was called", mock.ExecuteCount, "times.")
	}

	if mock.ExecuteData.(tokenToInsert).Token != "test token" {
		t.Error("template.Execute should be called with test token. It was called with: ", mock.ExecuteData)
	}
}

func TestInsertTokenbToTemplateReturnsErrorWhenTemplateExecuteReturnsError(t *testing.T) {
	mock := TemplateMock{}
	mock.ExecuteError = true

	_, err := insertTokenToTemplate("test token", &mock)

	if err == nil {
		t.Error("Should return error")
	}
}

type FakeAuthenticator struct {
	Err error
}

func (am FakeAuthenticator) Token(state string, r *http.Request) (*oauth2.Token, error) {
	if am.Err != nil {
		return nil, errors.New("")
	}
	return &oauth2.Token{}, nil
}

func (am FakeAuthenticator) NewClient(token *oauth2.Token) spotify.Client {
	return spotify.Client{}
}

func (am FakeAuthenticator) AuthURL(state string) string {
	return ""
}

func TestAuthCallback(t *testing.T) {
	cases := []struct {
		f                  FakeAuthenticator
		expectedStatusCode int
		expectedJSsnippet  string
	}{
		{
			f: FakeAuthenticator{
				Err: nil,
			},
			expectedStatusCode: http.StatusOK,
			expectedJSsnippet:  "<script src=\"https://sdk.scdn.co/spotify-player.js\"></script>",
		},
		{
			f: FakeAuthenticator{
				Err: errors.New(""),
			},
			expectedStatusCode: http.StatusNotFound,
			expectedJSsnippet:  "",
		},
	}

	for _, c := range cases {
		auth = c.f
		r := httptest.NewRecorder()
		server := server{client: make(chan *spotify.Client)}
		go func() {
			<-server.client
		}()
		server.authCallback(r, httptest.NewRequest("GET", "/", nil))
		if c.expectedStatusCode != r.Result().StatusCode {
			t.Errorf("Expected status to be %d but it was %d", c.expectedStatusCode, r.Result().StatusCode)
		}
		if actualBody := r.Body.String(); strings.Contains(actualBody, c.expectedJSsnippet) != true {
			t.Errorf("Expected body to contain %s", c.expectedJSsnippet)
		}
	}
}

type FakeCmd struct {
}

func (fc *FakeCmd) Start() error {
	return nil
}

func TestNotSupportedOS(t *testing.T) {
	runtimeGOOS = "Windows 10"
	err := openBroswerWith("http://golang.org")
	if expectedMsg := "Sorry, Windows 10 OS is not supported"; err.Error() != expectedMsg {
		t.Fatal("Expected error to be: %s, have %s", expectedMsg, err)
	}
}
