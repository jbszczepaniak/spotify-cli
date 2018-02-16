package main

import (
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
	"log"
)

type Album struct {
	artist string
	title  string
}

var albums []Album

func main() {
	client := authenticate()

	spotifyAlbums, err := client.CurrentUsersAlbums()
	if err != nil {
		log.Fatal(err)
	}
	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	currentlyPlayingLabel := tui.NewLabel(currentlyPlaying.Item.Name)

	currentlyPlayingBox := tui.NewHBox(currentlyPlayingLabel)
	currentlyPlayingBox.SetBorder(true)
	currentlyPlayingBox.SetTitle("Currently playing")

	albumsList := renderAlbumsTable(spotifyAlbums)

	playButton := tui.NewButton("[ ▷ Play]")
	stopButton := tui.NewButton("[ ■ Stop]")
	previousButton := tui.NewButton("[ |◄ Previous ]")
	nextButton := tui.NewButton("[ ►| Next ]")

	playButton.OnActivated(func(*tui.Button) {
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

	playBackButtons := []tui.Widget{previousButton, playButton, stopButton, nextButton}

	buttons := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(1, 0, previousButton),
		tui.NewPadder(1, 0, playButton),
		tui.NewPadder(1, 0, stopButton),
		tui.NewPadder(1, 0, nextButton),
	)

	box := tui.NewVBox(
		albumsList,
		currentlyPlayingBox,
		tui.NewSpacer(),
		buttons,
	)
	box.SetBorder(true)
	box.SetTitle("SPOTIFY CLI")

	tui.DefaultFocusChain.Set(playBackButtons...)

	ui, err := tui.New(box)
	if err != nil {
		panic(err)
	}

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		panic(err)
	}
}

func updateCurrentlyPlayingLabel(client *spotify.Client, label *tui.Label) {
	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		log.Fatal(err)
	}
	label.SetText(currentlyPlaying.Item.Name)
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
