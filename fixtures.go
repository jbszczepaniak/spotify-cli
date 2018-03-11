package main

import (
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// FakedClient is a dummy struct used when running in debug mode
type FakedClient struct {
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
func (fc FakedClient) CurrentUsersAlbums() (*spotify.SavedAlbumPage, error) {
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

// Play is a dummy implementation used when running in debug mode
func (fc FakedClient) Play() error {
	return nil
}

// Previous is a dummy implementation used when running in debug mode
func (fc FakedClient) Previous() error {
	return nil
}

// Pause is a dummy implementation used when running in debug mode
func (fc FakedClient) Pause() error {
	return nil
}

// Next is a dummy implementation used when running in debug mode
func (fc FakedClient) Next() error {
	return nil
}

// PlayOpt is a dummy implementation used when running in debug mode
func (fc FakedClient) PlayOpt(opt *spotify.PlayOptions) error {
	return nil
}

// PlayerCurrentlyPlaying is a dummy implementation used when running in debug mode
func (fc FakedClient) PlayerCurrentlyPlaying() (*spotify.CurrentlyPlaying, error) {
	return &spotify.CurrentlyPlaying{Item: &spotify.FullTrack{SimpleTrack: spotify.SimpleTrack{Name: "Currently Playing Song"}}}, nil
}

// PlayerDevices is a dummy implementation used when running in debug mode
func (fc FakedClient) PlayerDevices() ([]spotify.PlayerDevice, error) {

	return []spotify.PlayerDevice{
		{Name: "iPad", Type: "Tablet"},
		{Name: "iPhone", Type: "Smarthphone"},
		{Name: "Mac", Type: "App Player"},
	}, nil
}

// TransferPlayback is a dummy implementation used when running in debug mode
func (fc FakedClient) TransferPlayback(id spotify.ID, play bool) error {
	return nil
}

// CurrentUser is a dummy implementation used when running in debug mode
func (fc FakedClient) CurrentUser() (*spotify.PrivateUser, error) {
	return &spotify.PrivateUser{}, nil
}

// Token is a dummy implementation used when running in debug mode
func (fc FakedClient) Token() (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

// Search is a dummy implementation used when running in debug mode
func (fc FakedClient) Search(query string, t spotify.SearchType) (*spotify.SearchResult, error) {
	return nil, nil
}
