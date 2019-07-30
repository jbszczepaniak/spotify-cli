package web

import (
	"fmt"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"log"
	"net/http"
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

type SpotifyAuthenticatorInterface interface {
	AuthURL(string) string
	Token(string, *http.Request) (*oauth2.Token, error)
	NewClient(*oauth2.Token) spotify.Client
}



// authCallback is a function to by Spotify upon successful
// user login at their site
func (s *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
