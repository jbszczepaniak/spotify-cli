package main

import (
	"flag"
	"github.com/jedruniu/spotify-cli/web"
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
	log.SetFlags(log.Llongfile)
	f, _ := os.Create("log.txt")
	defer f.Close()
	log.SetOutput(f)

	checkMode()

	var client SpotifyClient
	var spotifyAuthenticator = web.NewSpotifyAuthenticator()

	authHandler := web.AuthHandler{
		Client:            make(chan *spotify.Client),
		State:             uuid.New().String(),
		Authenticator: spotifyAuthenticator,
	}

	webSocketHandler := web.WebsocketHandler{
		PlayerShutdown:    make(chan bool),
		PlayerDeviceID:    make(chan spotify.ID),
		PlayerStateChange: make(chan *web.WebPlaybackState),
	}

	if debugMode {
		client = NewDebugClient()
		go func() {
			webSocketHandler.PlayerDeviceID <- "debug"
		}()
	} else {
		var err error

		h := http.NewServeMux()
		h.HandleFunc("/ws", webSocketHandler.Handle)
		h.HandleFunc("/spotify-cli", authHandler.AuthCallback) // TODO: How to do Handler with pointer receiver?
		h.HandleFunc("/player", web.PlayerHandle)


		go func() {
			log.Fatal(http.ListenAndServe(":8888", h))
		}()

		err = startRemoteAuthentication(spotifyAuthenticator, authHandler.State)
		if err != nil {
			log.Printf("could not get client, shutting down, err: %v", err)
		}
	}

	// wait for authentication to complete
	client = <- authHandler.Client

	// wait for device to be ready
	webPlayerID := <- webSocketHandler.PlayerDeviceID

	sidebar, _ := NewSideBar(client)
	search := NewSearch(client)
	playback := NewPlayback(client, webSocketHandler.PlayerStateChange, webPlayerID)

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
		webSocketHandler.PlayerShutdown <- true
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
