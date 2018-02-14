package spotifycli

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zmb3/spotify"
	"net/http"
	"os"
)

var redirectURI = "http://localhost:8888/spotify-cli"

var (
	auth = spotify.NewAuthenticator(
		redirectURI,
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserLibraryRead,
	)
	ch    = make(chan *spotify.Client)
	state = uuid.New().String()
)

var clientId = os.Getenv("SPOTIFY_CLIENT_ID")
var clientSecret = os.Getenv("SPOTIFY_SECRET")

func authenticate() *spotify.Client {
	auth.SetAuthInfo(clientId, clientSecret)
	url := auth.AuthURL(state)

	http.HandleFunc("/spotify-cli", func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.Token(state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusNotFound)
			return
		}
		client := auth.NewClient(token)
		ch <- &client
	})

	go http.ListenAndServe(":8888", nil)

	fmt.Println("Please log in to Spotify by visiting the following page in your browser:")
	fmt.Println(url)

	client := <-ch
	return client
}
