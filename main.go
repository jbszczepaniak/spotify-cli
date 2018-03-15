package main

import (
	"flag"
	"fmt"
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"log"
	"os"
	"strings"
	"time"
)

type albumDescription struct {
	artist string
	title  string
}

var albumsDescriptions []albumDescription

type devicesTable struct {
	table *tui.Table
	box   *tui.Box
}

type currentlyPlaying struct {
	box      tui.Box
	song     string
	devices  devicesTable
	playback playback
}

type playback struct {
	previous *tui.Button
	next     *tui.Button
	stop     *tui.Button
	play     *tui.Button
	box      *tui.Box
}

type albumsList struct {
	table tui.Table
	box   tui.Box
}

type layout struct {
	currently currentlyPlaying
}

var debugMode bool

func checkMode() {
	debugModeFlag := flag.Bool("debug", false, "When set to true, app is populated with faked data and is not connecting with Spotify Web API.")
	flag.Parse()
	debugMode = *debugModeFlag
}

// SpotifyClient is a wrapper interface around spotify.client
// used in order to improve testability of the code.
type SpotifyClient interface {
	CurrentUsersAlbums() (*spotify.SavedAlbumPage, error)
	Player
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

type Player interface {
	Play() error
	PlayOpt(opt *spotify.PlayOptions) error
}

func main() {
	f, _ := os.Create("log.txt")
	defer f.Close()
	log.SetOutput(f)

	checkMode()
	var client SpotifyClient
	if debugMode {
		client = NewDebugClient()
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

	searchedSongs := NewSearchResults(client)
	searchedAlbums := NewSearchResults(client)
	searchedArtists := NewSearchResults(client)

	search := tui.NewEntry()
	search.OnSubmit(func(entry *tui.Entry) {
		result, _ := client.Search(
			entry.Text(),
			spotify.SearchTypeAlbum|spotify.SearchTypeTrack|spotify.SearchTypeArtist,
		)

		searchedAlbums.resetSearchResults()
		for _, i := range result.Albums.Albums {
			searchedAlbums.appendSearchResult(URIName{Name: i.Name, URI: i.URI})
		}

		searchedSongs.resetSearchResults()
		for _, i := range result.Tracks.Tracks {
			searchedSongs.appendSearchResult(URIName{Name: i.Name, URI: i.URI})
		}

		searchedArtists.resetSearchResults()
		for _, i := range result.Artists.Artists {
			searchedArtists.appendSearchResult(URIName{Name: i.Name, URI: i.URI})
		}

	})
	search.SetSizePolicy(tui.Preferred, tui.Minimum)
	searchBox := tui.NewHBox(search, tui.NewSpacer())
	searchBox.SetTitle("Search")
	searchBox.SetBorder(true)

	searchResults := tui.NewVBox(searchedSongs.box, searchedAlbums.box, searchedArtists.box)
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
	focusables = append(focusables, searchedSongs.table)
	focusables = append(focusables, searchedAlbums.table)
	focusables = append(focusables, searchedArtists.table)
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
		closeBrowser <- true
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
		currentSongName = getTrackRepr(currentlyPlaying.Item)
	}
	label.SetText(currentSongName)
}

func createPlaybackButtons(client SpotifyClient, currentlyPlayingLabel *tui.Label) playback {
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

	return playback{
		play:     playButton,
		stop:     stopButton,
		previous: previousButton,
		next:     nextButton,
		box:      buttons,
	}
}

func createAvailableDevicesTable(client SpotifyClient) devicesTable {
	time.Sleep(time.Second * 2) // Workaround, will be done smarter
	table := tui.NewTable(0, 0)
	tableBox := tui.NewHBox(table)
	tableBox.SetTitle("Devices")
	tableBox.SetBorder(true)

	avalaibleDevices, err := client.PlayerDevices()
	if err != nil {
		return devicesTable{box: tableBox, table: table}
	}
	table.AppendRow(
		tui.NewLabel("Name"),
		tui.NewLabel("Type"),
	)
	for i, device := range avalaibleDevices {
		log.Println(device)
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

	return devicesTable{box: tableBox, table: table}
}

func transferPlaybackToDevice(client SpotifyClient, pd *spotify.PlayerDevice) {
	client.TransferPlayback(pd.ID, true)
}

func renderAlbumsTable(albumsPage *spotify.SavedAlbumPage) *tui.Box {
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

func getTrackRepr(track *spotify.FullTrack) string {
	var artistsNames []string
	for _, artist := range track.Artists {
		artistsNames = append(artistsNames, artist.Name)
	}
	return fmt.Sprintf("%v (%v)", track.Name, strings.Join(artistsNames, ", "))
}
