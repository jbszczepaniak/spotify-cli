package main

import (
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

type AuthenticatorMock struct {
}

var returnErrToken bool

func (am AuthenticatorMock) Token(state string, r *http.Request) (*oauth2.Token, error) {
	if returnErrToken {
		return nil, fmt.Errorf("could not return token")
	}
	return &oauth2.Token{}, nil

}
func (am AuthenticatorMock) NewClient(token *oauth2.Token) spotify.Client {
	return spotify.Client{}
}
func (am AuthenticatorMock) AuthURL(state string) string {
	return ""
}

func TestAuthCallBackReturnsPageWithWebPlaybackSDK(t *testing.T) {
	auth = &AuthenticatorMock{} // Especially mock out Token method which must return valid token.

	ch = make(chan *spotify.Client)
	go func() {
		<-ch
	}()

	r := httptest.NewRecorder()
	authCallback(r, httptest.NewRequest("GET", "/", nil))
	spotifyPlaybackScript := "<script src=\"https://sdk.scdn.co/spotify-player.js\"></script>"

	if actualBody := r.Body.String(); strings.Contains(actualBody, spotifyPlaybackScript) != true {
		t.Log(actualBody)
		t.Error("Body does not contain spotify playback script")
	}
}

func TestAuthCallBackReturnsErrorIfTokenNotCreated(t *testing.T) {
	returnErrToken = true
	auth = &AuthenticatorMock{}
	r := httptest.NewRecorder()
	authCallback(r, httptest.NewRequest("GET", "/", nil))
	if statusCode := r.Result().StatusCode; statusCode != http.StatusNotFound {
		t.Errorf("Should return status code %d, returned %d", http.StatusNotFound, statusCode)
	}
}
