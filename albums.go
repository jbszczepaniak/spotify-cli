package main

import (
	"fmt"

	tui "github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
)

type albumDescription struct {
	artist string
	title  string
	uri    spotify.URI
}

type sideBar struct {
	albumList *AlbumList
	box       *tui.Box
}

type AlbumList struct {
	client             SpotifyClient
	albumsDescriptions []albumDescription
	currDataIdx        int
	lastTwoSelected    []int
	table              *tui.Table
	box                *tui.Box
}

var (
	visibleAlbums        = 45
	spotifyAPIPageSize   = 25
	spotifyAPIPageOffset = 25
	uiColumnWidth        = 20
)

func NewSideBar(client SpotifyClient) (*sideBar, error) {
	al := NewAlbumList(client)
	err := al.render()
	if err != nil {
		return nil, err
	}
	box := tui.NewHBox(al.box, tui.NewSpacer())
	return &sideBar{albumList: al, box: box}, nil
}

func NewAlbumList(client SpotifyClient) *AlbumList {
	table := tui.NewTable(0, 0)
	table.SetColumnStretch(0, 1)
	table.SetColumnStretch(1, 1)
	table.SetColumnStretch(2, 4)

	albumListBox := tui.NewVBox(table, tui.NewSpacer())
	albumListBox.SetBorder(true)
	albumListBox.SetTitle("User albums")
	albumListBox.SetSizePolicy(tui.Preferred, tui.Expanding)

	return &AlbumList{
		client:          client,
		currDataIdx:     0,
		lastTwoSelected: []int{-1, -1},
		table:           table,
		box:             albumListBox,
	}
}

func (albumList *AlbumList) render() error {
	savedAlbums, err := albumList.getUserAlbums()
	if err != nil {
		return err
	}
	for _, album := range savedAlbums {
		albumList.albumsDescriptions = append(albumList.albumsDescriptions, albumDescription{album.Name, album.Artists[0].Name, album.URI})
	}

	albumList.renderAlbumListPage(0, visibleAlbums)

	albumList.table.OnSelectionChanged(func(t *tui.Table) {
		if albumList.nextPage() {
			albumList.renderAlbumListPage(
				(albumList.currDataIdx/visibleAlbums)*visibleAlbums,
				(albumList.currDataIdx/visibleAlbums)*visibleAlbums+visibleAlbums,
			)
			albumList.lastTwoSelected = []int{-1, -1}
			t.Select(1)
			return
		}
		if albumList.previousPage(t.Selected()) {
			albumList.renderAlbumListPage(
				(albumList.currDataIdx/visibleAlbums)*visibleAlbums-visibleAlbums,
				(albumList.currDataIdx/visibleAlbums)*visibleAlbums,
			)
			albumList.lastTwoSelected = []int{visibleAlbums + 2, visibleAlbums + 1}
			t.Select(visibleAlbums)
			return
		}
		albumList.updateIndexes(t.Selected())
	})

	albumList.table.OnItemActivated(func(t *tui.Table) {
		albumList.client.PlayOpt(&spotify.PlayOptions{PlaybackContext: &albumList.albumsDescriptions[albumList.currDataIdx-1].uri})
	})

	return nil
}

func (albumList *AlbumList) getUserAlbums() ([]spotify.SavedAlbum, error) {
	initialPage, err := albumList.client.CurrentUsersAlbumsOpt(&spotify.Options{Limit: &spotifyAPIPageSize})
	if err != nil {
		return nil, fmt.Errorf("could not fetch current user albums: %v", err)
	}
	userAlbums := make([]spotify.SavedAlbum, 0)
	userAlbums = append(userAlbums, initialPage.Albums...)

	page := initialPage
	for page.Offset < initialPage.Total {
		page, err = albumList.client.CurrentUsersAlbumsOpt(&spotify.Options{
			Limit:  &initialPage.Limit,
			Offset: &spotifyAPIPageOffset,
		})
		if err != nil {
			return nil, fmt.Errorf("could not fetch page current user albums: %v", err)
		}
		spotifyAPIPageOffset += spotifyAPIPageSize
		userAlbums = append(userAlbums, page.Albums...)
	}
	return userAlbums, nil
}

func (albumList *AlbumList) nextPage() bool {
	return albumList.lastTwoSelected[0] == visibleAlbums-1 && albumList.lastTwoSelected[1] == visibleAlbums
}

func (albumList *AlbumList) previousPage(selected int) bool {
	return albumList.lastTwoSelected[1] == 1 && selected == 0 && albumList.currDataIdx >= visibleAlbums
}

func (albumList *AlbumList) updateIndexes(selected int) {
	albumList.lastTwoSelected[0] = albumList.lastTwoSelected[1]
	albumList.lastTwoSelected[1] = selected

	if albumList.lastTwoSelected[0] > albumList.lastTwoSelected[1] {
		albumList.currDataIdx--
	}
	if albumList.lastTwoSelected[0] < albumList.lastTwoSelected[1] {
		albumList.currDataIdx++
	}
}

func (albumList *AlbumList) renderAlbumListPage(start, end int) {
	albumList.table.RemoveRows()
	albumList.table.AppendRow(
		tui.NewLabel("Title"),
		tui.NewLabel("Artist"),
	)
	for _, album := range albumList.albumsDescriptions[start:end] {
		albumList.table.AppendRow(
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
