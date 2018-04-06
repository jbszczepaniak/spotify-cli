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

type albums struct {
	table *tui.Table
	box   *tui.Box
	data  []spotify.URI
}

var (
	visibleAlbums      = 45
	spotifyAPIPageSize = 25
	uiColumnWidth      = 20
)

func NewSideBar(client SpotifyClient) (*sideBar, error) {
	userAlbums, err := getUserAlbums(client)
	if err != nil {
		return nil, err
	}
	albumsList := renderAlbumsTable(userAlbums, client)
	box := tui.NewHBox(albumsList.box, tui.NewSpacer())
	return &sideBar{albums: albumsList, box: box}, nil
}

func getUserAlbums(client SpotifyClient) ([]spotify.SavedAlbum, error) {
	initialPage, err := client.CurrentUsersAlbumsOpt(&spotify.Options{Limit: &spotifyAPIPageSize})
	if err != nil {
		return nil, fmt.Errorf("could not fetch current user albums: %v", err)
	}
	userAlbums := make([]spotify.SavedAlbum, 0)
	userAlbums = append(userAlbums, initialPage.Albums...)

	page := initialPage
	for page.Offset < initialPage.Total {
		page, err = client.CurrentUsersAlbumsOpt(&spotify.Options{
			Limit:  &initialPage.Limit,
			Offset: &spotifyAPIPageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("could not fetch page current user albums: %v", err)
		}
		spotifyAPIPageSize += 25
		userAlbums = append(userAlbums, page.Albums...)
	}
	return userAlbums, nil
}

func renderAlbumListPage(albumsList *tui.Table, albumsDescriptions []albumDescription, start, end int) {
	albumsList.AppendRow(
		tui.NewLabel("Title"),
		tui.NewLabel("Artist"),
	)
	for _, album := range albumsDescriptions[start:end] {
		albumsList.AppendRow(
			tui.NewLabel(trimWithCommasIfTooLong(album.artist, uiColumnWidth)),
			tui.NewLabel(trimWithCommasIfTooLong(album.title, uiColumnWidth)),
		)
	}
}

func trimWithCommasIfTooLong(text string, maxLength int) string {
	if len(text) > maxLength {
		text = text[:maxLength] + "..."
	}
	return text
}

func renderAlbumsTable(savedAlbums []spotify.SavedAlbum, client SpotifyClient) *albums {
	lastTwoSelected := []int{-1, -1}
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

	renderAlbumListPage(albumsList, albumsDescriptions, 0, visibleAlbums)

	albumsList.OnSelectionChanged(func(t *tui.Table) {
		if lastTwoSelected[0] == visibleAlbums-1 && lastTwoSelected[1] == visibleAlbums {
			t.RemoveRows()
			renderAlbumListPage(t, albumsDescriptions, (currDataIdx/visibleAlbums)*visibleAlbums, (currDataIdx/visibleAlbums)*visibleAlbums+visibleAlbums)
			lastTwoSelected = []int{-1, -1}
			t.Select(1)
			return
		}
		if lastTwoSelected[1] == 1 && t.Selected() == 0 && currDataIdx >= visibleAlbums {
			t.RemoveRows()

			renderAlbumListPage(t, albumsDescriptions, (currDataIdx/visibleAlbums)*visibleAlbums-visibleAlbums, (currDataIdx/visibleAlbums)*visibleAlbums)
			lastTwoSelected = []int{visibleAlbums + 2, visibleAlbums + 1}
			t.Select(visibleAlbums)
			return
		}

		lastTwoSelected[0] = lastTwoSelected[1]
		lastTwoSelected[1] = t.Selected()

		if lastTwoSelected[0] > lastTwoSelected[1] {
			currDataIdx--
		}
		if lastTwoSelected[0] < lastTwoSelected[1] {
			currDataIdx++
		}
	})

	albumsList.OnItemActivated(func(t *tui.Table) {
		client.PlayOpt(&spotify.PlayOptions{PlaybackContext: &data[currDataIdx-1]})
	})
	albumListBox := tui.NewVBox(albumsList, tui.NewSpacer())
	albumListBox.SetBorder(true)
	albumListBox.SetTitle("User albums")
	albumListBox.SetSizePolicy(tui.Preferred, tui.Expanding)
	return &albums{box: albumListBox, table: albumsList}
}
