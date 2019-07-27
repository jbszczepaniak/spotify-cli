package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"time"

	"github.com/google/uuid"
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
)

type albumsList struct {
	table tui.Table
	box   tui.Box
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
	userAlbumFetcher
	player
	searcher
	Pause() error
	Previous() error
	Next() error
	PlayerCurrentlyPlaying() (*spotify.CurrentlyPlaying, error)
	PlayerDevices() ([]spotify.PlayerDevice, error)
	TransferPlayback(spotify.ID, bool) error
}

type player interface {
	Play() error
	PlayOpt(opt *spotify.PlayOptions) error
}

type searcher interface {
	Search(string, spotify.SearchType) (*spotify.SearchResult, error)
}

type userAlbumFetcher interface {
	CurrentUsersAlbumsOpt(opt *spotify.Options) (*spotify.SavedAlbumPage, error)
}

func main() {
	log.SetFlags(log.Lshortfile)
	f, _ := os.Create("log.txt")
	defer f.Close()
	log.SetOutput(f)

	checkMode()

	var client SpotifyClient
	var spotifyAuthenticator = NewSpotifyAuthenticator()

	server := server{
		client:            make(chan *spotify.Client),
		playerShutdown:    make(chan bool),
		playerDeviceID:    make(chan spotify.ID),
		state:             uuid.New().String(),
		playerStateChange: make(chan *WebPlaybackState),
		authenticator: spotifyAuthenticator,
	}

	if debugMode {
		client = NewDebugClient()
		go func() {
			server.playerDeviceID <- "debug"
		}()
	} else {
		var err error

		h := http.NewServeMux()
		h.HandleFunc("/ws", server.handleWebSocket)
		h.HandleFunc("/spotify-cli", server.authCallback)

		go func() {
			log.Fatal(http.ListenAndServe(":8888", h))
		}()

		err = startRemoteAuthentication(spotifyAuthenticator, server.state)
		if err != nil {
			log.Printf("could not get client, shutting down, err: %v", err)
		}
	}

	// wait for authentication to complete
	client = <- server.client

	// wait for device to be ready
	webPlayerID := <- server.playerDeviceID

	sidebar, _ := NewSideBar(client)
	search := NewSearch(client)
	playback := NewPlayback(client, server.playerStateChange, webPlayerID)

	mainFrame := tui.NewVBox(
		search.box,
		tui.NewSpacer(),
		playback.box,
	)
	mainFrame.SetSizePolicy(tui.Expanding, tui.Expanding)

	window := tui.NewHBox(
		sidebar.box,
		mainFrame,
	)
	window.SetTitle("SPOTIFY CLI")

	playBackButtons := []tui.Widget{playback.playback.previous, playback.playback.play, playback.playback.stop, playback.playback.next}
	focusables := append(playBackButtons, sidebar.albumList.table)
	focusables = append(focusables, search.focusables...)
	focusables = append(focusables, playback.devices.table)

	tui.DefaultFocusChain.Set(focusables...)

	theme := tui.DefaultTheme
	theme.SetStyle("box.focused.border", tui.Style{Fg: tui.ColorYellow, Bg: tui.ColorDefault})
	theme.SetStyle("table.focused.border", tui.Style{Fg: tui.ColorYellow, Bg: tui.ColorDefault})

	ui, err := tui.New(window)
	if err != nil {
		panic(err)
	}

	ui.SetKeybinding("Esc", func() {
		ui.Quit()
		server.playerShutdown <- true
		return
	})

	go func() {
		for range time.Tick(500 * time.Millisecond) {
			ui.Update(func() {})
		}
	}()

	if err := ui.Run(); err != nil {
		panic(err)
	}

}
