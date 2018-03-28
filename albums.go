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

type sideBar struct {
	albums *albums
	box    *tui.Box
}

func NewSideBar(client SpotifyClient) (*sideBar, error) {
	initialPage, err := client.CurrentUsersAlbumsOpt(&spotify.Options{Limit: &visibleUserAlbumsCount})
	if err != nil {
		return nil, fmt.Errorf("could not fetch current user albums: %v", err)
	}
	userAlbums := make([]spotify.SavedAlbum, 0)
	userAlbums = append(userAlbums, initialPage.Albums...)

	offset := 25
	page := initialPage
	for page.Offset < initialPage.Total {
		page, err = client.CurrentUsersAlbumsOpt(&spotify.Options{
			Limit:  &initialPage.Limit,
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("could not fetch page current user albums: %v", err)
		}
		offset += 25
		userAlbums = append(userAlbums, page.Albums...)
	}
	albumsList := renderAlbumsTable(userAlbums, client)
	box := tui.NewHBox(albumsList.box, tui.NewVBox(tui.NewSpacer()), tui.NewSpacer())
	return &sideBar{albums: albumsList, box: box}, nil
}

type albums struct {
	table *tui.Table
	box   *tui.Box
	data  []spotify.URI
}

var visibleUserAlbumsCount = 25

func renderAlbumListPage(albumsList *tui.Table, albumsDescriptions []albumDescription, start, end int) {
	albumsList.AppendRow(
		tui.NewLabel("Title"),
		tui.NewLabel("Artist"),
	)
	colLength := 20
	for _, album := range albumsDescriptions[start:end] {
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
}

func renderAlbumsTable(savedAlbums []spotify.SavedAlbum, client SpotifyClient) *albums {

	currDataIdx := 0
	albumsDescriptions := make([]albumDescription, 0)
	data := make([]spotify.URI, 0)
	for _, album := range savedAlbums {
		albumsDescriptions = append(albumsDescriptions, albumDescription{album.Name, album.Artists[0].Name})
		data = append(data, album.URI)
	}

	albumsList := tui.NewTable(0, 0)
	albumsList.SetColumnStretch(0, 1)
	albumsList.SetColumnStretch(1, 1)
	albumsList.SetColumnStretch(2, 4)

	renderAlbumListPage(albumsList, albumsDescriptions, 0, 25)

	albumsList.OnSelectionChanged(func(t *tui.Table) {
		currDataIdx = t.Selected() - 1
	})

	albumsList.OnItemActivated(func(t *tui.Table) {
		client.PlayOpt(&spotify.PlayOptions{PlaybackContext: &data[currDataIdx]})
	})
	albumListBox := tui.NewVBox(albumsList)
	albumListBox.SetBorder(true)
	albumListBox.SetTitle("User albums")
	return &albums{box: albumListBox, table: albumsList}
}
