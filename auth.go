package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

var (
	auth         = getSpotifyAuthenticator()
	state        = uuid.New().String()
	redirectURI  = "http://localhost:8888/spotify-cli"
	clientID     = os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret = os.Getenv("SPOTIFY_SECRET")
)

type spotifyAuthenticatorInterface interface {
	AuthURL(string) string
	Token(string, *http.Request) (*oauth2.Token, error)
	NewClient(*oauth2.Token) spotify.Client
}

func getSpotifyAuthenticator() spotifyAuthenticatorInterface {
	auth := spotify.NewAuthenticator(
		redirectURI,
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
	auth.SetAuthInfo(clientID, clientSecret)
	return auth
}

type appState struct {
	client         chan *spotify.Client
	playerShutdown chan bool
	playerDeviceId chan spotify.ID
}

// authenticate authenticate user with Sotify API
func authenticate(as appState) (SpotifyClient, error) {
	h := http.NewServeMux()
	h.HandleFunc("/ws", as.handleWebSocket)
	h.HandleFunc("/spotify-cli", as.authCallback)
	go http.ListenAndServe(":8888", h)

	url := auth.AuthURL(state)
	err := openBrowserWith(url)
	if err != nil {
		return nil, fmt.Errorf("Could not open browser")
	}
	return <-as.client, err
}

// authCallback is a function to by Spotify upon successful
// user login at their site
func (s *appState) authCallback(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusNotFound)
		return
	}

	client := auth.NewClient(token)
	s.client <- &client

	t, _ := template.ParseFiles("index_tmpl.html")

	playbackPage, err := insertTokenToTemplate(token.AccessToken, t)

	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusNotFound)
		return
	}

	fmt.Fprint(w, playbackPage)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *appState) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	type WebPlayBackState struct {
		DeviceReady string
	}

	var v WebPlayBackState
	_, message, err := conn.ReadMessage()
	json.Unmarshal(message, &v)
	s.playerDeviceId <- spotify.ID(v.DeviceReady)

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
		return execCommand("open", "-a", "/Applications/Google Chrome.app", url).Start()
	case "linux":
		return execCommand("xdg-open", url).Start()
	default:
		return fmt.Errorf("Sorry, %v OS is not supported", runtimeGOOS)
	}
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
