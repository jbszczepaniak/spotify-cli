package player

import (
	"fmt"
	"log"

	tui "github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
)

// SideBar represents box with album list inside this box.
type SideBar struct {
	AlbumList *AlbumList
	Box       *tui.Box
}

type renderer interface {
	render() error
}

type pageRenderer interface {
	renderPage([]albumDescription, int, int) error
}

type dataFetcher interface {
	fetchUserAlbums() ([]albumDescription, error)
}

type pagination interface {
	nextPage() bool
	previousPage() bool
	updateIndexes()

	getCurrDataIdx() int
	setLastTwoSelected([]int)
}

// AlbumList represents list of albums with underlying data,
// table to display, box in which table is places, indexes
// pointing to currently playing item, and last chosen items.
type AlbumList struct {
	client             SpotifyClient
	albumsDescriptions []albumDescription
	Table              *tui.Table
	box                *tui.Box

	renderer
	pageRenderer
	dataFetcher
	pagination
}

type albumDescription struct {
	artist string
	title  string
	uri    spotify.URI
}

var (
	visibleAlbums        = 45
	spotifyAPIPageSize   = 25
	spotifyAPIPageOffset = 25
	uiColumnWidth        = 20
)

// NewSideBar creates struct which holds references to
// SideBar Box and AlbumList placed inside SideBar
func NewSideBar(client SpotifyClient) (*SideBar, error) {
	al := newEmptyAlbumList(client)
	err := al.render()
	if err != nil {
		return nil, err
	}
	box := tui.NewHBox(al.box, tui.NewSpacer())
	return &SideBar{AlbumList: al, Box: box}, nil
}

func newEmptyAlbumList(client SpotifyClient) *AlbumList {
	table := tui.NewTable(0, 0)
	table.SetColumnStretch(0, 1)
	table.SetColumnStretch(1, 1)
	table.SetColumnStretch(2, 4)

	albumListBox := tui.NewVBox(table, tui.NewSpacer())
	albumListBox.SetBorder(true)
	albumListBox.SetTitle("User albums")
	albumListBox.SetSizePolicy(tui.Preferred, tui.Expanding)

	return &AlbumList{
		client:             client,
		Table:              table,
		box:                albumListBox,
		albumsDescriptions: []albumDescription{},

		dataFetcher:  &fetchUserAlbumsStruct{client: client},
		pageRenderer: &renderPageStruct{table: table},
		pagination:   &paginatorStruct{table: table, lastTwoSelected: []int{-1, -1}, currDataIdx: 0},
	}
}

func (albumList *AlbumList) render() error {
	albumsDescriptions, err := albumList.dataFetcher.fetchUserAlbums()
	if err != nil {
		return err
	}
	albumList.albumsDescriptions = albumsDescriptions
	err = albumList.pageRenderer.renderPage(albumList.albumsDescriptions, 0, visibleAlbums)
	if err != nil {
		return err
	}
	albumList.Table.OnSelectionChanged(albumList.onSelectedChanged())
	albumList.Table.OnItemActivated(albumList.onItemActivaed())
	return nil
}

type fetchUserAlbumsStruct struct {
	client SpotifyClient
}

func (fetchUserAlbumsStruct *fetchUserAlbumsStruct) fetchUserAlbums() ([]albumDescription, error) {
	initialPage, err := fetchUserAlbumsStruct.client.CurrentUsersAlbumsOpt(&spotify.Options{Limit: &spotifyAPIPageSize})
	if err != nil {
		return nil, fmt.Errorf("could not fetch current user albums: %v", err)
	}
	userAlbums := make([]spotify.SavedAlbum, 0)
	userAlbums = append(userAlbums, initialPage.Albums...)

	page := initialPage
	for spotifyAPIPageOffset < initialPage.Total {
		page, err = fetchUserAlbumsStruct.client.CurrentUsersAlbumsOpt(&spotify.Options{
			Limit:  &initialPage.Limit,
			Offset: &spotifyAPIPageOffset,
		})
		if err != nil {
			return nil, fmt.Errorf("could not fetch page current user albums: %v", err)
		}
		spotifyAPIPageOffset += spotifyAPIPageSize
		userAlbums = append(userAlbums, page.Albums...)
	}

	albumsDescriptions := make([]albumDescription, 0)
	for _, album := range userAlbums {
		albumsDescriptions = append(albumsDescriptions, albumDescription{album.Name, album.Artists[0].Name, album.URI})
	}
	return albumsDescriptions, nil
}

func (albumList *AlbumList) onSelectedChanged() func(*tui.Table) {
	return func(t *tui.Table) {
		if albumList.nextPage() {
			err := albumList.renderPage(
				albumList.albumsDescriptions,
				(albumList.getCurrDataIdx()/visibleAlbums)*visibleAlbums,
				(albumList.getCurrDataIdx()/visibleAlbums)*visibleAlbums+visibleAlbums,
			)
			if err != nil {
				log.Printf("Could not render next page of albums with %s", err)
				return
			}
			albumList.setLastTwoSelected([]int{-1, -1})
			t.Select(1)
			return
		}
		if albumList.previousPage() {
			err := albumList.renderPage(
				albumList.albumsDescriptions,
				(albumList.getCurrDataIdx()/visibleAlbums)*visibleAlbums-visibleAlbums,
				(albumList.getCurrDataIdx()/visibleAlbums)*visibleAlbums,
			)
			if err != nil {
				log.Printf("Could not render previous page of albums with %s", err)
				return
			}
			albumList.setLastTwoSelected([]int{visibleAlbums + 2, visibleAlbums + 1})
			t.Select(visibleAlbums)
			return
		}
		albumList.updateIndexes()
	}
}

type paginatorStruct struct {
	currDataIdx     int
	lastTwoSelected []int
	table           *tui.Table
}

func (paginator *paginatorStruct) setLastTwoSelected(lastTwoSelected []int) {
	paginator.lastTwoSelected = lastTwoSelected
}

func (paginator *paginatorStruct) getCurrDataIdx() int {
	return paginator.currDataIdx
}

func (paginator *paginatorStruct) nextPage() bool {
	return paginator.lastTwoSelected[0] == visibleAlbums-1 && paginator.lastTwoSelected[1] == visibleAlbums
}

func (paginator *paginatorStruct) previousPage() bool {
	return paginator.lastTwoSelected[1] == 1 && paginator.table.Selected() == 0 && paginator.currDataIdx >= visibleAlbums
}

func (paginator *paginatorStruct) updateIndexes() {
	paginator.lastTwoSelected[0] = paginator.lastTwoSelected[1]
	paginator.lastTwoSelected[1] = paginator.table.Selected()

	if paginator.lastTwoSelected[0] > paginator.lastTwoSelected[1] {
		paginator.currDataIdx--
	}
	if paginator.lastTwoSelected[0] < paginator.lastTwoSelected[1] {
		paginator.currDataIdx++
	}
}

func (albumList *AlbumList) onItemActivaed() func(*tui.Table) {
	return func(t *tui.Table) {
		// -2 because tui.Table starts counting at 1, and additional 1 is added because first row is a header
		uri := &albumList.albumsDescriptions[albumList.pagination.getCurrDataIdx()-2].uri
		err := albumList.client.PlayOpt(&spotify.PlayOptions{PlaybackContext: uri})
		if err != nil {
			log.Printf("Error occured while trying to play track with uri: %s", *uri)
		}
	}
}

type renderPageStruct struct {
	table *tui.Table
}

func (renderPageStruct *renderPageStruct) renderPage(albumsDescriptions []albumDescription, start, end int) error {
	renderPageStruct.table.RemoveRows()
	renderPageStruct.table.AppendRow(
		tui.NewLabel("Title"),
		tui.NewLabel("Artist"),
	)
	if len(albumsDescriptions) == 0 {
		return fmt.Errorf("could not iterate over empty slice")
	}
	if len(albumsDescriptions) < end {
		end = len(albumsDescriptions) // This means that there is less user albums than there is displayed at once on the page.
	}
	for _, album := range albumsDescriptions[start:end] {
		renderPageStruct.table.AppendRow(
			tui.NewLabel(trimWithCommasIfTooLong(album.artist, uiColumnWidth)),
			tui.NewLabel(trimWithCommasIfTooLong(album.title, uiColumnWidth)),
		)
	}
	return nil
}

func trimWithCommasIfTooLong(text string, maxLength int) string {
	if len(text) > maxLength {
		text = text[:maxLength] + "..."
	}
	return text
}
