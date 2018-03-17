package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

func TestAuthenticateClient(t *testing.T) {
	cases := []struct {
		openBrowserWithErr bool
		expectedErr        bool
	}{
		{
			openBrowserWithErr: true,
			expectedErr:        true,
		},
		{
			openBrowserWithErr: false,
			expectedErr:        false,
		},
	}

	for _, c := range cases {
		openBrowserWith = func(s string) error {
			if c.expectedErr {
				return fmt.Errorf("Could not open browser")
			}
			return nil
		}
		defer func() { openBrowserWith = openBrowserWithImpl }()
		testState := appState{
			client:         make(chan *spotify.Client),
			playerShutdown: make(chan bool),
			playerDeviceId: make(chan spotify.ID),
		}
		go func() {
			testState.client <- &spotify.Client{}
		}()

		_, err := authenticate(testState)
		if err == nil && c.expectedErr {
			t.Error("Error expected, but there was not one")
		}
		if err != nil && !c.expectedErr {
			t.Error("Error was not expected, but there was not")
		}
	}
}

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
		f                        FakeAuthenticator
		expectedStatusCode       int
		expectedJSsnippet        string
		insertTokenToTemplateErr bool
	}{
		{
			f: FakeAuthenticator{
				Err: nil,
			},
			expectedStatusCode:       http.StatusOK,
			expectedJSsnippet:        "<script src=\"https://sdk.scdn.co/spotify-player.js\"></script>",
			insertTokenToTemplateErr: false,
		},
		{
			f: FakeAuthenticator{
				Err: nil,
			},
			expectedStatusCode:       http.StatusNotFound,
			expectedJSsnippet:        "",
			insertTokenToTemplateErr: true,
		},
		{
			f: FakeAuthenticator{
				Err: errors.New(""),
			},
			expectedStatusCode:       http.StatusNotFound,
			expectedJSsnippet:        "",
			insertTokenToTemplateErr: false,
		},
	}

	for _, c := range cases {
		if c.insertTokenToTemplateErr {
			insertTokenToTemplate = func(token string, template templateInterface) (string, error) {
				return "", errors.New("")
			}
		}
		auth = c.f
		r := httptest.NewRecorder()
		as := appState{client: make(chan *spotify.Client)}
		go func() {
			<-as.client
		}()
		as.authCallback(r, httptest.NewRequest("GET", "/", nil))
		if c.expectedStatusCode != r.Result().StatusCode {
			t.Errorf("Expected status to be %d but it was %d", c.expectedStatusCode, r.Result().StatusCode)
		}
		if actualBody := r.Body.String(); strings.Contains(actualBody, c.expectedJSsnippet) != true {
			t.Errorf("Expected body to contain %s", c.expectedJSsnippet)
		}
	}
}

func TestOpenBrowser(t *testing.T) {
	cases := []struct {
		GOOS        string
		expectedApp string
		expectedErr bool
	}{
		{
			GOOS:        "Windows 10",
			expectedApp: "",
			expectedErr: true,
		},
		{
			GOOS:        "linux",
			expectedApp: "xdg-open",
			expectedErr: false,
		},
		{
			GOOS:        "darwin",
			expectedApp: "open",
			expectedErr: false,
		},
	}
	url := "http://golang.org"
	for _, c := range cases {
		runtimeGOOS = c.GOOS
		var app string
		startCommand = func(command *exec.Cmd) error {
			app = command.Path
			return nil
		}
		err := openBrowserWith(url)
		if c.expectedErr && err == nil {
			t.Fatalf("Expected fail for OS %s, but it did not fail", c.GOOS)
		}
		// Check only if real app name contains substring becasuse
		// OS on which test is run will append /usr/bin
		if err == nil && !strings.Contains(app, c.expectedApp) {
			t.Fatalf("Expected to run with app %s on OS %s, but it run on %s", c.expectedApp, c.GOOS, app)
		}
	}
}

func TestWebSocketHandler(t *testing.T) {
	as := appState{
		client:         make(chan *spotify.Client),
		playerShutdown: make(chan bool),
		playerDeviceId: make(chan spotify.ID),
	}

	h := http.NewServeMux()
	h.HandleFunc("/ws", as.handleWebSocket)
	go http.ListenAndServe(":8005", h)

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial("ws://0.0.0.0:8005/ws", nil)
	if err != nil {
		t.Fatal("Unable to dial websocket")
	}
	expectedReadyDeviceId := "1jdjd8dj38djd09dfhjk"
	msg := WebPlayBackState{DeviceReady: expectedReadyDeviceId}
	b, _ := json.Marshal(msg)
	ws.WriteMessage(websocket.TextMessage, b)

	readyDeviceId := string(<-as.playerDeviceId)

	if readyDeviceId != expectedReadyDeviceId {
		t.Fatalf("Expected id of ready device to be %s, have %s", expectedReadyDeviceId, readyDeviceId)
	}

	defer ws.Close()
}
