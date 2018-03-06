package main

import (
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

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

func (fc FakedClient) Play() error {
	return nil
}

func (fc FakedClient) Previous() error {
	return nil
}

func (fc FakedClient) Pause() error {
	return nil
}

func (fc FakedClient) Next() error {
	return nil
}

func (fc FakedClient) PlayerCurrentlyPlaying() (*spotify.CurrentlyPlaying, error) {
	return &spotify.CurrentlyPlaying{Item: &spotify.FullTrack{SimpleTrack: spotify.SimpleTrack{Name: "Currently Playing Song"}}}, nil
}

func (fc FakedClient) PlayerDevices() ([]spotify.PlayerDevice, error) {

	return []spotify.PlayerDevice{
		{Name: "iPad", Type: "Tablet"},
		{Name: "iPhone", Type: "Smarthphone"},
		{Name: "Mac", Type: "App Player"},
	}, nil
}

func (fc FakedClient) TransferPlayback(id spotify.ID, play bool) error {
	return nil
}

func (fc FakedClient) CurrentUser() (*spotify.PrivateUser, error) {
	return &spotify.PrivateUser{}, nil
}

func (fc FakedClient) Token() (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}
