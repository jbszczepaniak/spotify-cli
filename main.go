package main

import (
	"flag"
	"fmt"
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"log"
	"strings"
	"time"
)

type Album struct {
	artist string
	title  string
}

var albums []Album

type DevicesTable struct {
	table *tui.Table
	box   *tui.Box
}

type CurrentlyPlaying struct {
	box      tui.Box
	song     string
	devices  DevicesTable
	playback Playback
}

type Playback struct {
	previous *tui.Button
	next     *tui.Button
	stop     *tui.Button
	play     *tui.Button
	box      *tui.Box
}

type AlbumsList struct {
	table tui.Table
	box   tui.Box
}

type Layout struct {
	currently CurrentlyPlaying
}

var debugMode bool

func checkMode() {
	debugModeFlag := flag.Bool("debug", false, "When set to true, app is populated with faked data and is not connecting with Spotify Web API.")
	flag.Parse()
	debugMode = *debugModeFlag
}

type SpotifyClient interface {
	CurrentUsersAlbums() (*spotify.SavedAlbumPage, error)
	Play() error
	PlayOpt(opt *spotify.PlayOptions) error
	Pause() error
	Previous() error
	Next() error
	PlayerCurrentlyPlaying() (*spotify.CurrentlyPlaying, error)
	PlayerDevices() ([]spotify.PlayerDevice, error)
	TransferPlayback(spotify.ID, bool) error
	CurrentUser() (*spotify.PrivateUser, error)
	Token() (*oauth2.Token, error)
	Search(query string, t spotify.SearchType) (*spotify.SearchResult, error)
}

func main() {
	checkMode()
	var client SpotifyClient
	if debugMode {
		client = FakedClient{}
	} else {
		client = authenticate()
	}

	spotifyAlbums, err := client.CurrentUsersAlbums()
	if err != nil {
		log.Fatal(err)
	}
	currentlyPlayingLabel := tui.NewLabel("")
	updateCurrentlyPlayingLabel(client, currentlyPlayingLabel)

	availableDevicesTable := createAvailableDevicesTable(client)
	albumsList := renderAlbumsTable(spotifyAlbums)
	albumsList.SetSizePolicy(tui.Minimum, tui.Preferred)
	sidebar := tui.NewHBox(tui.NewVBox(albumsList, tui.NewSpacer()), tui.NewSpacer())

	playbackButtons := createPlaybackButtons(client, currentlyPlayingLabel)

	currentlyPlayingBox := tui.NewHBox(currentlyPlayingLabel, availableDevicesTable.box, playbackButtons.box)
	currentlyPlayingBox.SetBorder(true)
	currentlyPlayingBox.SetTitle("Currently playing")

	searchedSongsTable := tui.NewTable(0, 0)
	searchedSongsSlice := make([]spotify.FullTrack, 0)
	searchedSongs := tui.NewVBox(searchedSongsTable, tui.NewSpacer())
	searchedSongs.SetTitle("Songs")
	searchedSongs.SetBorder(true)

	searchedAlbumsTable := tui.NewTable(0, 0)
	searchedAlbumsSlice := make([]spotify.SimpleAlbum, 0)
	searchedAlbums := tui.NewVBox(searchedAlbumsTable, tui.NewSpacer())
	searchedAlbums.SetTitle("Albums")
	searchedAlbums.SetBorder(true)

	searchedArtistsTable := tui.NewTable(0, 0)
	searchedArtistsSlice := make([]spotify.FullArtist, 0)
	searchedArtists := tui.NewVBox(searchedArtistsTable, tui.NewSpacer())
	searchedArtists.SetTitle("Artists")
	searchedArtists.SetBorder(true)

	searchedAlbumsTable.OnItemActivated(func(t *tui.Table) {
		selectedRow := t.Selected()
		albumURI := &searchedAlbumsSlice[selectedRow].URI
		client.PlayOpt(&spotify.PlayOptions{PlaybackContext: albumURI})
	})

	searchedSongsTable.OnItemActivated(func(t *tui.Table) {
		selectedRow := t.Selected()
		trackURI := &searchedSongsSlice[selectedRow].URI
		client.PlayOpt(&spotify.PlayOptions{URIs: []spotify.URI{*trackURI}}) // Playing single songs is different
	})

	searchedArtistsTable.OnItemActivated(func(t *tui.Table) {
		selectedRow := t.Selected()
		artistURI := &searchedArtistsSlice[selectedRow].URI
		client.PlayOpt(&spotify.PlayOptions{PlaybackContext: artistURI})
	})

	search := tui.NewEntry()
	search.OnSubmit(func(entry *tui.Entry) {
		result, _ := client.Search(
			entry.Text(),
			spotify.SearchTypeAlbum|spotify.SearchTypeTrack|spotify.SearchTypeArtist,
		)
		searchedAlbumsTable.RemoveRows()
		searchedAlbumsSlice = searchedAlbumsSlice[:0]

		searchedSongsTable.RemoveRows()
		searchedSongsSlice = searchedSongsSlice[:0]

		searchedArtistsTable.RemoveRows()
		searchedArtistsSlice = searchedArtistsSlice[:0]

		for _, i := range result.Albums.Albums {
			searchedAlbumsTable.AppendRow(tui.NewLabel(i.Name))
			searchedAlbumsSlice = append(searchedAlbumsSlice, i)
		}
		for _, i := range result.Tracks.Tracks {
			searchedSongsTable.AppendRow(tui.NewLabel(i.Name))
			searchedSongsSlice = append(searchedSongsSlice, i)
		}
		for _, i := range result.Artists.Artists {
			searchedArtistsTable.AppendRow(tui.NewLabel(i.Name))
			searchedArtistsSlice = append(searchedArtistsSlice, i)
		}

	})
	search.SetSizePolicy(tui.Preferred, tui.Minimum)
	searchBox := tui.NewHBox(search, tui.NewSpacer())
	searchBox.SetTitle("Search")
	searchBox.SetBorder(true)

	searchResults := tui.NewVBox(searchedSongs, searchedAlbums, searchedArtists)
	searchResults.SetTitle("Search Results")

	mainFrame := tui.NewVBox(
		searchBox,
		searchResults,
		tui.NewSpacer(),
		currentlyPlayingBox,
	)
	mainFrame.SetSizePolicy(tui.Expanding, tui.Expanding)

	box := tui.NewHBox(sidebar, mainFrame)
	box.SetTitle("SPOTIFY CLI")

	playBackButtons := []tui.Widget{playbackButtons.previous, playbackButtons.play, playbackButtons.stop, playbackButtons.next}
	focusables := append(playBackButtons, search)
	focusables = append(focusables, searchedSongsTable)
	focusables = append(focusables, searchedAlbumsTable)
	focusables = append(focusables, searchedArtistsTable)
	focusables = append(focusables, availableDevicesTable.table)

	tui.DefaultFocusChain.Set(focusables...)

	theme := tui.DefaultTheme
	theme.SetStyle("box.focused.border", tui.Style{Fg: tui.ColorYellow, Bg: tui.ColorDefault})
	theme.SetStyle("table.focused.border", tui.Style{Fg: tui.ColorYellow, Bg: tui.ColorDefault})

	// tui.DefaultTheme.SetStyle("table.focused.border", tui.Style{Fg: tui.ColorYellow, Bg: tui.ColorDefault})

	ui, err := tui.New(box)
	if err != nil {
		panic(err)
	}

	ui.SetKeybinding("Esc", func() {
		ui.Quit()
		return
	})

	if err := ui.Run(); err != nil {
		panic(err)
	}
}

func updateCurrentlyPlayingLabel(client SpotifyClient, label *tui.Label) {
	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	var currentSongName string
	if err != nil {
		currentSongName = "None"
	} else {
		currentSongName = GetTrackRepr(currentlyPlaying.Item)
	}
	label.SetText(currentSongName)
}

func createPlaybackButtons(client SpotifyClient, currentlyPlayingLabel *tui.Label) Playback {
	playButton := tui.NewButton("[ ▷ Play]")
	stopButton := tui.NewButton("[ ■ Stop]")
	previousButton := tui.NewButton("[ |◄ Previous ]")
	nextButton := tui.NewButton("[ ►| Next ]")

	playButton.OnActivated(func(btn *tui.Button) {
		updateCurrentlyPlayingLabel(client, currentlyPlayingLabel)
		client.Play()
	})

	stopButton.OnActivated(func(*tui.Button) {
		client.Pause()
	})

	previousButton.OnActivated(func(*tui.Button) {
		updateCurrentlyPlayingLabel(client, currentlyPlayingLabel)
		client.Previous()
	})

	nextButton.OnActivated(func(*tui.Button) {
		updateCurrentlyPlayingLabel(client, currentlyPlayingLabel)
		client.Next()
	})

	buttons := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(1, 0, previousButton),
		tui.NewPadder(1, 0, playButton),
		tui.NewPadder(1, 0, stopButton),
		tui.NewPadder(1, 0, nextButton),
	)
	buttons.SetBorder(true)

	return Playback{
		play:     playButton,
		stop:     stopButton,
		previous: previousButton,
		next:     nextButton,
		box:      buttons,
	}
}

func createAvailableDevicesTable(client SpotifyClient) DevicesTable {
	time.Sleep(time.Second * 2) // Workaround, will be done smarter
	table := tui.NewTable(0, 0)
	tableBox := tui.NewHBox(table)
	tableBox.SetTitle("Devices")
	tableBox.SetBorder(true)

	avalaibleDevices, err := client.PlayerDevices()
	if err != nil {
		return DevicesTable{box: tableBox, table: table}
	}
	table.AppendRow(
		tui.NewLabel("Name"),
		tui.NewLabel("Type"),
	)
	for i, device := range avalaibleDevices {
		table.AppendRow(
			tui.NewLabel(device.Name),
			tui.NewLabel(device.Type),
		)
		if device.Active {
			table.SetSelected(i)
		}
	}

	table.OnItemActivated(func(t *tui.Table) {
		selctedRow := t.Selected()
		if selctedRow == 0 {
			return // Selecting table header
		}
		transferPlaybackToDevice(client, &avalaibleDevices[selctedRow-1])
	})

	return DevicesTable{box: tableBox, table: table}
}

func transferPlaybackToDevice(client SpotifyClient, pd *spotify.PlayerDevice) {
	client.TransferPlayback(pd.ID, true)
}

func renderAlbumsTable(albumsPage *spotify.SavedAlbumPage) *tui.Box {
	for _, album := range albumsPage.Albums {
		albums = append(albums, Album{album.Name, album.Artists[0].Name})
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
	for _, album := range albums {
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

func GetTrackRepr(track *spotify.FullTrack) string {
	var artistsNames []string
	for _, artist := range track.Artists {
		artistsNames = append(artistsNames, artist.Name)
	}
	return fmt.Sprintf("%v (%v)", track.Name, strings.Join(artistsNames, ", "))
}
