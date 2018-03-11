package main

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
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
	auth          = getSpotifyAuthenticator()
	state         = uuid.New().String()
	redirectURI   = "http://localhost:8888/spotify-cli"
	clientId      = os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret  = os.Getenv("SPOTIFY_SECRET")
)

type SpotifyAuthenticatorInterface interface {
	AuthURL(string) string
	Token(string, *http.Request) (*oauth2.Token, error)
	NewClient(*oauth2.Token) spotify.Client
}

func getSpotifyAuthenticator() SpotifyAuthenticatorInterface {
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
	auth.SetAuthInfo(clientId, clientSecret)
	return auth
}

type Server struct {
	client chan *spotify.Client
	playerChanged chan string
}


// authenticate authenticate user with Sotify API
func authenticate() SpotifyClient {
	s := Server{
		client: make(chan *spotify.Client),
		playerChanged: make(chan string),
	}

	http.HandleFunc("/spotify-cli", s.authCallback)
	http.HandleFunc("/player-is-up", s.stateChangedCallback)
	go http.ListenAndServe(":8888", nil)

	url := auth.AuthURL(state)

	_, err := openBroswerWith(url)
	if err != nil {
		log.Fatal(err)
	}

	return <-s.client
}

func (s *Server) stateChangedCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	s.playerChanged <- r.FormValue("deviceId")
}

// openBrowserWith open browsers with given url and returns process id of opened browser
func openBroswerWith(url string) (int, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-a", "/Applications/Google Chrome.app", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return 0, fmt.Errorf("Sorry, %v OS is not supported", runtime.GOOS)
	}

	err := cmd.Start()
	if err != nil {
		return 0, err
	}
	process := cmd.Process
	return process.Pid, nil
}

// authCallback is a function to by Spotify upon successful
// user login at their site
func (s *Server) authCallback(w http.ResponseWriter, r *http.Request) {
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

type TemplateInterface interface {
	Execute(io.Writer, interface{}) error
}

type tokenToInsert struct {
	Token string
}

var osCreate = os.Create

func insertTokenToTemplate(token string, template TemplateInterface) (string, error) {
	buf := new(bytes.Buffer)
	err := template.Execute(buf, tokenToInsert{token})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
