package main

import "testing"
import "os"
import "io"
import "net/http/httptest"
import "github.com/zmb3/spotify"
import "golang.org/x/oauth2"
import "net/http"
import "fmt"

type TemplateMock struct {
	ExecuteCount int
	ExecuteData  interface{}
}

func (tm *TemplateMock) Execute(wr io.Writer, data interface{}) error {
	tm.ExecuteCount++
	tm.ExecuteData = data
	return nil
}

func TestInsertTokenToTemplate(t *testing.T) {
	mock := TemplateMock{}

	insertTokenToTemplate("test token", &mock)

	if mock.ExecuteCount != 1 {
		t.Error("template.Execute should be called once. It was called", mock.ExecuteCount, "times.")
	}

	if mock.ExecuteData.(tokenToInsert).Token != "test token" {
		t.Error("template.Execute should be called with test token. It was called with: ", mock.ExecuteData)
	}

	os.Remove("index.html")
}

type AuthenticatorMock struct {
}

func (am AuthenticatorMock) Token(state string, r *http.Request) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}
func (am AuthenticatorMock) NewClient(token *oauth2.Token) spotify.Client {
	return spotify.Client{}
}
func (am AuthenticatorMock) AuthURL(state string) string {
	return ""
}

func TestAuthCallBackReturnsHTMLWithNameFromSpotifyUser(t *testing.T) {
	auth = &AuthenticatorMock{}
	displayName := "George"
	getCurrentUser = func(spotify.Client) (*spotify.PrivateUser, error) {
		return &spotify.PrivateUser{User: spotify.User{DisplayName: displayName}}, nil
	}

	ch = make(chan *spotify.Client)
	go func() {
		<-ch
	}()

	r := httptest.NewRecorder()
	authCallback(r, httptest.NewRequest("GET", "/", nil))
	expectedBody := fmt.Sprintf("<h1>Logged into spotify cli as:</h1>\n<p>%v</p>", displayName)

	if r.Body.String() != expectedBody {
		t.Error("server returned wrong HTML, expected: ", expectedBody)
	}
}
