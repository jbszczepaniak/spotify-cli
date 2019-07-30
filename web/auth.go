package web

import (
	"fmt"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"net/url"
	"os"
)

type AuthHandler struct {
	// Client is created as the result of Spotify backend calling auth callback,
	// as soon as spotify Client is created, the rest of the application can use
	// is to call any spotify apis.
	Client chan *spotify.Client

	// State is random string used in order to authenticate Client. It is used to
	// generate spotify backend URL to which user will be redirected. After user
	// comes back to the application (to the auth callback) it is used to verify
	// message received from the spotify backend.
	State string

	// used by auth callback to verify message from spotify backend
	// and to create a spotify Client.
	Authenticator SpotifyAuthenticatorInterface
}


// authCallback is a function to by Spotify upon successful
// user login at their site
func (s *AuthHandler) AuthCallback(w http.ResponseWriter, r *http.Request) {
	token, err := s.Authenticator.Token(s.State, r)
	if err != nil {
		errMsg := fmt.Sprintf("Could not get token, error: %v", err)
		http.Error(w, errMsg, http.StatusNotFound)
		log.Print(errMsg)
		return
	}

	client := s.Authenticator.NewClient(token)
	s.Client <- &client

	http.Redirect(w, r, fmt.Sprintf("http://localhost:8888/player?token=%s", token.AccessToken), 301)
}



type SpotifyAuthenticatorInterface interface {
	AuthURL(string) string
	Token(string, *http.Request) (*oauth2.Token, error)
	NewClient(*oauth2.Token) spotify.Client
}

func NewSpotifyAuthenticator() SpotifyAuthenticatorInterface {
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