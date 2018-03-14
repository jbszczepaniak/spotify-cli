package main

import (
	// "fmt"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

func NewDebugClient() SpotifyClient {
	return DebugClient{Player: &DebugPlayer{}}
}

// DebugClient is a dummy struct used when running in debug mode
type DebugClient struct {
	Player
}

type DebugPlayer struct{}

// Play is a dummy implementation used when running in debug mode
func (fc DebugPlayer) Play() error {
	return nil
}

// PlayOpt is a dummy implementation used when running in debug mode
func (fc DebugPlayer) PlayOpt(opt *spotify.PlayOptions) error {
	return nil
}

var albumsArtistsData = []struct {
	AlbumName  string
	ArtistName string
}{
	{"Interstellar", "Hans Zimmer"},
	{"Tubular Bells", "Mike Oldfield"},
	{"A Humdrum Star(Deluxe)", "GoGo Penguin"},
	{"Timeline", "Yellowjackets"},
	{"Floa", "Mammal Hands"},
	{"groundUP", "Snarky Puppy"},
}

// CurrentUsersAlbums is a dummy implementation used when running in debug mode
func (fc DebugClient) CurrentUsersAlbums() (*spotify.SavedAlbumPage, error) {
	fakedAlbums := make([]spotify.SavedAlbum, 0)
	for _, albumArtist := range albumsArtistsData {
		album := spotify.SavedAlbum{}
		album.Name = albumArtist.AlbumName
		album.Artists = []spotify.SimpleArtist{spotify.SimpleArtist{Name: albumArtist.ArtistName}}
		fakedAlbums = append(fakedAlbums, album)
	}
	return &spotify.SavedAlbumPage{
		Albums: fakedAlbums,
	}, nil
}

// Previous is a dummy implementation used when running in debug mode
func (fc DebugClient) Previous() error {
	return nil
}

// Pause is a dummy implementation used when running in debug mode
func (fc DebugClient) Pause() error {
	return nil
}

// Next is a dummy implementation used when running in debug mode
func (fc DebugClient) Next() error {
	return nil
}

// PlayerCurrentlyPlaying is a dummy implementation used when running in debug mode
func (fc DebugClient) PlayerCurrentlyPlaying() (*spotify.CurrentlyPlaying, error) {
	return &spotify.CurrentlyPlaying{Item: &spotify.FullTrack{SimpleTrack: spotify.SimpleTrack{Name: "Currently Playing Song"}}}, nil
}

// PlayerDevices is a dummy implementation used when running in debug mode
func (fc DebugClient) PlayerDevices() ([]spotify.PlayerDevice, error) {

	return []spotify.PlayerDevice{
		{Name: "iPad", Type: "Tablet"},
		{Name: "iPhone", Type: "Smarthphone"},
		{Name: "Mac", Type: "App Player"},
	}, nil
}

// TransferPlayback is a dummy implementation used when running in debug mode
func (fc DebugClient) TransferPlayback(id spotify.ID, play bool) error {
	return nil
}

// CurrentUser is a dummy implementation used when running in debug mode
func (fc DebugClient) CurrentUser() (*spotify.PrivateUser, error) {
	return &spotify.PrivateUser{}, nil
}

// Token is a dummy implementation used when running in debug mode
func (fc DebugClient) Token() (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

// Search is a dummy implementation used when running in debug mode
func (fc DebugClient) Search(query string, t spotify.SearchType) (*spotify.SearchResult, error) {
	return nil, nil
}
