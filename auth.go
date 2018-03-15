package main

import (
	"bytes"
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
	closeBrowser = make(chan bool)
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

type server struct {
	client        chan *spotify.Client
	playerChanged chan string
}

// authenticate authenticate user with Sotify API
func authenticate() SpotifyClient {
	s := server{
		client:        make(chan *spotify.Client),
		playerChanged: make(chan string),
	}

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/spotify-cli", s.authCallback)
	http.HandleFunc("/player-is-up", s.stateChangedCallback)
	go http.ListenAndServe(":8888", nil)

	url := auth.AuthURL(state)
	err := openBroswerWith(url)
	if err != nil {
		log.Fatal(err)
	}

	return <-s.client
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	<-closeBrowser
	conn.WriteJSON("{\"close\": true}")
}

func (s *server) stateChangedCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	s.playerChanged <- r.FormValue("deviceId")
}

var runtimeGOOS = runtime.GOOS
var execCommand = exec.Command

// openBrowserWith open browsers with given url
func openBroswerWith(url string) error {
	switch runtimeGOOS {
	case "darwin":
		return execCommand("open", "-a", "/Applications/Google Chrome.app", url).Start()
	case "linux":
		return execCommand("xdg-open", url).Start()
	default:
		return fmt.Errorf("Sorry, %v OS is not supported", runtimeGOOS)
	}
}

// authCallback is a function to by Spotify upon successful
// user login at their site
func (s *server) authCallback(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Couldn't get token", http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, playbackPage)
}

type templateInterface interface {
	Execute(io.Writer, interface{}) error
}

type tokenToInsert struct {
	Token string
}

var osCreate = os.Create

func insertTokenToTemplate(token string, template templateInterface) (string, error) {
	buf := new(bytes.Buffer)
	err := template.Execute(buf, tokenToInsert{token})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
