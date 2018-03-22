package main

import (
	"fmt"

	tui "github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
)

type albumDescription struct {
	artist string
	title  string
}

func NewSideBar(client SpotifyClient) (tui.Widget, error) {
	spotifyAlbums, err := client.CurrentUsersAlbums()
	if err != nil {
		return nil, fmt.Errorf("could not fetch current user albums: %v", err)
	}
	albumsList := renderAlbumsTable(spotifyAlbums)
	albumsList.SetSizePolicy(tui.Minimum, tui.Preferred)
	return tui.NewHBox(tui.NewVBox(albumsList, tui.NewSpacer()), tui.NewSpacer()), nil
}

func renderAlbumsTable(albumsPage *spotify.SavedAlbumPage) *tui.Box {
	albumsDescriptions := make([]albumDescription, 0)
	for _, album := range albumsPage.Albums {
		albumsDescriptions = append(albumsDescriptions, albumDescription{album.Name, album.Artists[0].Name})
	}
	albumsList := tui.NewTable(0, 0)
	albumsList.SetColumnStretch(0, 1)
	albumsList.SetColumnStretch(1, 1)
	albumsList.SetColumnStretch(2, 4)

	albumsList.AppendRow(
		tui.NewLabel("Title"),
		tui.NewLabel("Artist"),
	)
	colLength := 20
	for _, album := range albumsDescriptions {
		var artistRow, albumRow *tui.Label
		if len(album.artist) > colLength {
			artistRow = tui.NewLabel(album.artist[:colLength] + "...")
		} else {
			artistRow = tui.NewLabel(album.artist)
		}

		if len(album.title) > colLength {
			albumRow = tui.NewLabel(album.title[:colLength] + "...")
		} else {
			albumRow = tui.NewLabel(album.title)
		}

		albumsList.AppendRow(artistRow, albumRow)
	}
	albumListBox := tui.NewVBox(albumsList)
	albumListBox.SetBorder(true)
	albumListBox.SetTitle("User albums")
	return albumListBox
}
