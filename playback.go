package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
)

type devicesTable struct {
	table *tui.Table
	box   *tui.Box
}

type currentlyPlaying struct {
	box      tui.Widget
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

func NewPlayback(client SpotifyClient, as appState) currentlyPlaying {
	currentlyPlayingLabel := tui.NewLabel("")

	updateCurrentlyPlayingLabel(client, currentlyPlayingLabel)
	availableDevicesTable, _ := createAvailableDevicesTable(as, client)

	playbackButtons := createPlaybackButtons(client, currentlyPlayingLabel)

	currentlyPlayingBox := tui.NewHBox(currentlyPlayingLabel, availableDevicesTable.box, playbackButtons.box)
	currentlyPlayingBox.SetBorder(true)
	currentlyPlayingBox.SetTitle("Currently playing")
	return currentlyPlaying{
		box:      currentlyPlayingBox,
		devices:  *availableDevicesTable,
		playback: playbackButtons,
	}
}

func updateCurrentlyPlayingLabel(client SpotifyClient, label *tui.Label) {
	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	var currentSongName string
	if err != nil {
		log.Printf("could not currently playing track - fallback to None, %s", err)
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

func createAvailableDevicesTable(state appState, client SpotifyClient) (*devicesTable, error) {
	SDKplayerID := <-state.playerDeviceId
	err := transferPlaybackToDevice(client, SDKplayerID)
	if err != nil {
		return nil, err
	}

	table := tui.NewTable(0, 0)
	tableBox := tui.NewHBox(table)
	tableBox.SetTitle("Devices")
	tableBox.SetBorder(true)

	avalaibleDevices, err := client.PlayerDevices()
	if err != nil {
		return nil, err
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
		if device.ID == SDKplayerID {
			table.SetSelected(i + 1)
		}
	}

	table.OnItemActivated(func(t *tui.Table) {
		selctedRow := t.Selected()
		if selctedRow == 0 {
			return // Selecting table header
		}
		transferPlaybackToDevice(client, avalaibleDevices[selctedRow-1].ID)
	})

	return &devicesTable{box: tableBox, table: table}, nil
}

func transferPlaybackToDevice(client SpotifyClient, id spotify.ID) error {
	return client.TransferPlayback(id, true)
}

func getTrackRepr(track *spotify.FullTrack) string {
	var artistsNames []string
	for _, artist := range track.Artists {
		artistsNames = append(artistsNames, artist.Name)
	}
	return fmt.Sprintf("%v (%v)", track.Name, strings.Join(artistsNames, ", "))
}
