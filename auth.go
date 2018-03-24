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
	client            chan *spotify.Client
	playerShutdown    chan bool
	playerDeviceId    chan spotify.ID
	playerStateChange chan *WebPlaybackState // currently playing
	state             string
}

// authenticate authenticate user with Sotify API
func authenticate(as appState) (SpotifyClient, error) {
	h := http.NewServeMux()
	h.HandleFunc("/ws", as.handleWebSocket)
	h.HandleFunc("/spotify-cli", as.authCallback)
	go http.ListenAndServe(":8888", h)

	url := auth.AuthURL(as.state)
	err := openBrowserWith(url)
	if err != nil {
		return nil, fmt.Errorf("Could not open browser")
	}
	return <-as.client, err
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
	s.playerDeviceId <- spotify.ID(v.DeviceId)

	go func() {
		for {
			var y WebPlaybackState
			_, message, err = conn.ReadMessage()
			err = json.Unmarshal(message, &y)
			if err != nil {
				log.Printf("Could not Unmarshall message: %s, err: %s", message, err)
			}
			s.playerStateChange <- &y
		}
	}()

	<-s.playerShutdown
	conn.WriteJSON("{\"close\": true}")

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
