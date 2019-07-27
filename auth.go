package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"os"
	"os/exec"
	"runtime"

	"github.com/gorilla/websocket"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var auth = getSpotifyAuthenticator()

type spotifyAuthenticatorInterface interface {
	AuthURL(string) string
	Token(string, *http.Request) (*oauth2.Token, error)
	NewClient(*oauth2.Token) spotify.Client
}

func getSpotifyAuthenticator() spotifyAuthenticatorInterface {
	envKeys := []string{"SPOTIFY_CLIENT_ID", "SPOTIFY_SECRET"}
	envVars := map[string]string{}
	for _, key := range envKeys {
		v := os.Getenv(key)
		if v == "" {
			log.Fatalf("Quiting, there is no %s environment variable.", key)
		}
		envVars[key] = v
	}

	redirectURI := url.URL{Scheme: "http", Host: "localhost:8888", Path: "/spotify-cli"}

	auth := spotify.NewAuthenticator(
		redirectURI.String(),
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserLibraryRead,
		// Used for Web Playback SDK
		"streaming",
		spotify.ScopeUserReadBirthdate,
		spotify.ScopeUserReadEmail,
	)
	auth.SetAuthInfo(envVars["SPOTIFY_CLIENT_ID"], envVars["SPOTIFY_SECRET"])
	return auth
}

type appState struct {
	// client is created as the result of Spotify backend calling auth callback,
	// as soon as spotify client is created, the rest of the application can use
	// is to call any spotify apis.
	client            chan *spotify.Client

	// signal sent by the application to the websocket handler that triggers tear down
	// of websocket.
	playerShutdown    chan bool

	// as soon as web player is ready it sends player device id as the very first
	// message on the websocket connection. Application may want to use this ID in
	// order to control on which player to play music (the one that it created, or
	// maybe on the other device that is also working)
	playerDeviceID    chan spotify.ID

	// each time web player changes it's state it sens information about what is
	// currently played on the websocket
	playerStateChange chan *WebPlaybackState

	// State is random string used in order to authenticate client. It is used to
	// generate spotify backend URL to which user will be redirected. After user
	// comes back to the application (to the auth callback) it is used to verify
	// message received from the spotify backend.
	state             string
}

// startRemoteAuthentication redirects to spotify's API in order to authenticate user
func startRemoteAuthentication(state string) error {
	authUrl := auth.AuthURL(state)
	err := openBrowserWith(authUrl)
	if err != nil {
		return fmt.Errorf("could not open browser with url: %s, err: %v", authUrl, err)
	}
	return nil
}

// authCallback is a function to by Spotify upon successful
// user login at their site
func (s *appState) authCallback(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(s.state, r)
	if err != nil {
		errMsg := fmt.Sprintf("Could not get token, error: %v", err)
		http.Error(w, errMsg, http.StatusNotFound)
		log.Print(errMsg)
		return
	}

	client := auth.NewClient(token)
	s.client <- &client

	t, _ := template.ParseFiles("index_tmpl.html")

	playbackPage, err := insertTokenToTemplate(token.AccessToken, t)

	if err != nil {
		errMsg := fmt.Sprintf("Could not get token, error: %v", err)
		http.Error(w, errMsg, http.StatusNotFound)
		log.Print(errMsg)
		return
	}

	fmt.Fprint(w, playbackPage)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebPlaybackReadyDevice struct {
	DeviceId string
}

type WebPlaybackState struct {
	CurrentTrackName  string
	CurrentAlbumName  string
	CurrentArtistName string
}

func (s *appState) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var v WebPlaybackReadyDevice
	_, message, err := conn.ReadMessage()
	err = json.Unmarshal(message, &v)
	s.playerDeviceID <- spotify.ID(v.DeviceId)

	go func() {
		for range time.Tick(500 * time.Millisecond) {
			var state WebPlaybackState
			_, message, err = conn.ReadMessage()
			err = json.Unmarshal(message, &state)
			if err != nil {
				log.Printf("could not Unmarshall message: %s, err: %s", message, err)
			}
			s.playerStateChange <- &state
		}
	}()

	<-s.playerShutdown
	err = conn.WriteJSON("{\"close\": true}")
	if err != nil {
		log.Printf("could not close connection, err %v", err)
	}

}

var runtimeGOOS = runtime.GOOS
var execCommand = exec.Command

var openBrowserWith = openBrowserWithImpl

// openBrowserWith open browsers with given url
func openBrowserWithImpl(url string) error {
	switch runtimeGOOS {
	case "darwin":
		return startCommand(execCommand("open", "-a", "/Applications/Google Chrome.app", url))
	case "linux":
		return startCommand(execCommand("xdg-open", url))
	default:
		return fmt.Errorf("Sorry, %v OS is not supported", runtimeGOOS)
	}
}

var startCommand = startCommandImpl

func startCommandImpl(command *exec.Cmd) error {
	return command.Start()
}

type templateInterface interface {
	Execute(io.Writer, interface{}) error
}

type tokenToInsert struct {
	Token string
}

var osCreate = os.Create
var insertTokenToTemplate = insertTokenToTemplateImpl

func insertTokenToTemplateImpl(token string, template templateInterface) (string, error) {
	buf := new(bytes.Buffer)
	err := template.Execute(buf, tokenToInsert{token})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
