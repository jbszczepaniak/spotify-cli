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

var albumsPageSize = 45

var NewSideBar = func(client SpotifyClient) (*sideBar, error) {
	initialPage, err := client.CurrentUsersAlbumsOpt(&spotify.Options{Limit: &albumsPageSize})
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
	box := tui.NewHBox(albumsList.box, tui.NewSpacer())
	return &sideBar{albums: albumsList, box: box}, nil
}

func renderAlbumListPage(albumsList *tui.Table, albumsDescriptions []albumDescription, start, end int) {
	albumsList.AppendRow(
		tui.NewLabel("Title"),
		tui.NewLabel("Artist"),
	)
	colLength := 20
	for _, album := range albumsDescriptions[start:end] {
		albumsList.AppendRow(
			tui.NewLabel(trimWithCommasIfTooLong(album.artist, colLength)),
			tui.NewLabel(trimWithCommasIfTooLong(album.title, colLength)),
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

	renderAlbumListPage(albumsList, albumsDescriptions, 0, albumsPageSize)

	albumsList.OnSelectionChanged(func(t *tui.Table) {
		if lastTwoSelected[0] == albumsPageSize-1 && lastTwoSelected[1] == albumsPageSize {
			t.RemoveRows()
			renderAlbumListPage(t, albumsDescriptions, (currDataIdx/albumsPageSize)*albumsPageSize, (currDataIdx/albumsPageSize)*albumsPageSize+albumsPageSize)
			lastTwoSelected = []int{-1, -1}
			t.Select(1)
			return
		}
		if lastTwoSelected[1] == 1 && t.Selected() == 0 && currDataIdx >= albumsPageSize {
			t.RemoveRows()

			renderAlbumListPage(t, albumsDescriptions, (currDataIdx/albumsPageSize)*albumsPageSize-albumsPageSize, (currDataIdx/albumsPageSize)*albumsPageSize)
			lastTwoSelected = []int{albumsPageSize + 2, albumsPageSize + 1}
			t.Select(albumsPageSize)
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
