package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
	"log"
	"net/http"
	"os"
)

type Album struct {
	artist string
	title  string
}

var albums []Album

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

func main() {

	clientId := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_SECRET")

	if clientId == "" {
		fmt.Println("SPOTIFY_CLIENT_ID not provided")
		return
	}

	if clientSecret == "" {
		fmt.Println("SPOTIFY_SECRET not provided")
		return
	}

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
	salbums, err := client.CurrentUsersAlbums()

	if err != nil {
		log.Fatal(err)
	}
	for _, album := range salbums.Albums {
		albums = append(albums, Album{album.Name, album.Artists[0].Name})
	}

	albumsList := tui.NewTable(0, 0)
	albumsList.SetColumnStretch(0, 1)
	albumsList.SetColumnStretch(1, 1)
	albumsList.SetColumnStretch(2, 4)

	albumsList.AppendRow(
		tui.NewLabel("Artist"),
		tui.NewLabel("Title"),
	)

	for _, album := range albums {
		albumsList.AppendRow(
			tui.NewLabel(album.artist),
			tui.NewLabel(album.title),
		)
	}
	albumsList.Select(3)

	box := tui.NewVBox(
		albumsList,
		tui.NewSpacer(),
	)

	box.SetTitle("SPOTIFY CLI")

	ui, err := tui.New(box)
	if err != nil {
		panic(err)
	}

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		panic(err)
	}
}
