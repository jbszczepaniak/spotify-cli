package main

import (
	"flag"
	"fmt"
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"log"
	"strings"
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
	Pause() error
	Previous() error
	Next() error
	PlayerCurrentlyPlaying() (*spotify.CurrentlyPlaying, error)
	PlayerDevices() ([]spotify.PlayerDevice, error)
	TransferPlayback(spotify.ID, bool) error
	CurrentUser() (*spotify.PrivateUser, error)
	Token() (*oauth2.Token, error)
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
	playbackButtons := createPlaybackButtons(client, currentlyPlayingLabel)

	currentlyPlayingBox := tui.NewHBox(currentlyPlayingLabel, availableDevicesTable.box, playbackButtons.box)
	currentlyPlayingBox.SetBorder(true)
	currentlyPlayingBox.SetTitle("Currently playing")

	search := tui.NewEntry()
	searchBox := tui.NewHBox(search)
	searchBox.SetTitle("Search")
	searchBox.SetBorder(true)

	box := tui.NewVBox(
		searchBox,
		albumsList,
		currentlyPlayingBox,
	)
	box.SetTitle("SPOTIFY CLI")

	playBackButtons := []tui.Widget{playbackButtons.next, playbackButtons.stop, playbackButtons.previous, playbackButtons.play}
	focusables := append(playBackButtons, search)
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
		tui.NewLabel("Artist"),
		tui.NewLabel("Title"),
	)

	for _, album := range albums {
		albumsList.AppendRow(
			tui.NewLabel(album.artist),
			tui.NewLabel(album.title),
		)
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
