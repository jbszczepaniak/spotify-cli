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
	"golang.org/x/oauth2"
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
	CurrentUser() (*spotify.PrivateUser, error)
	Token() (*oauth2.Token, error)
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

	as := appState{
		client:            make(chan *spotify.Client),
		playerShutdown:    make(chan bool),
		playerDeviceID:    make(chan spotify.ID),
		state:             uuid.New().String(),
		playerStateChange: make(chan *WebPlaybackState),
	}

	if debugMode {
		client = NewDebugClient()
		go func() {
			as.playerDeviceID <- "debug"
		}()
	} else {
		var err error

		h := http.NewServeMux()
		h.HandleFunc("/ws", as.handleWebSocket)
		h.HandleFunc("/spotify-cli", as.authCallback)

		go func() {
			log.Fatal(http.ListenAndServe(":8888", h))
		}()

		err = startRemoteAuthentication(as.state)
		if err != nil {
			log.Printf("could not get client, shutting down, err: %v", err)
		}
	}

	// wait for authentication to complete
	client = <- as.client


	sidebar, _ := NewSideBar(client)
	search := NewSearch(client)
	playback := NewPlayback(client, as)

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
		as.playerShutdown <- true
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
