package player

import (
	"fmt"
	"log"
	"time"

	"github.com/jedruniu/spotify-cli/pkg/web"

	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
)

type DevicesTable struct {
	Table *tui.Table
	box   *tui.Box
}

type currentlyPlaying struct {
	Box      tui.Widget
	song     string
	Devices  DevicesTable
	Playback Playback
}

type Playback struct {
	Previous *tui.Button
	Next     *tui.Button
	Stop     *tui.Button
	Play     *tui.Button
	Box      *tui.Box
}

// NewPlayback creates data structure representing current spotify playback.
func NewPlayback(client SpotifyClient, playerStateChanges chan *web.WebPlaybackState, webPlayerID spotify.ID) currentlyPlaying {
	currentlyPlayingLabel := tui.NewLabel("")
	go func() {
		for {
			currentState := <-playerStateChanges
			labelText := fmt.Sprintf(
				"%s\n%s\n%s",
				currentState.CurrentTrackName,
				currentState.CurrentAlbumName,
				currentState.CurrentArtistName,
			)
			currentlyPlayingLabel.SetText(labelText)
		}
	}()

	updateCurrentlyPlayingLabel(client, currentlyPlayingLabel)

	// TODO handle error
	_ = transferPlaybackToDevice(client, webPlayerID)
	availableDevicesTable, err := createAvailableDevicesTable(client, webPlayerID)
	if err != nil {
		log.Fatalf("err occured: %v", err)
	}

	playbackButtons := createPlaybackButtons(client, currentlyPlayingLabel)

	currentlyPlayingBox := tui.NewHBox(currentlyPlayingLabel, availableDevicesTable.box, playbackButtons.Box)
	currentlyPlayingBox.SetBorder(true)
	currentlyPlayingBox.SetTitle("Currently playing")
	return currentlyPlaying{
		Box:      currentlyPlayingBox,
		Devices:  *availableDevicesTable,
		Playback: playbackButtons,
	}
}

func updateCurrentlyPlayingLabel(client SpotifyClient, label *tui.Label) {
	currentlyPlaying, err := client.PlayerCurrentlyPlaying()
	var currentSongName string
	if err != nil {
		log.Printf("could not fetch currently playing track - fallback to None, %s", err)
		currentSongName = "None"
	} else {
		currentSongName = getTrackRepr(currentlyPlaying.Item)
	}
	label.SetText(currentSongName)
}

func createPlaybackButtons(client SpotifyClient, currentlyPlayingLabel *tui.Label) Playback {
	playButton := tui.NewButton("[ ▷ Play]")
	stopButton := tui.NewButton("[ ■ Stop]")
	previousButton := tui.NewButton("[ |◄ Previous ]")
	nextButton := tui.NewButton("[ ►| Next ]")

	playButton.OnActivated(func(btn *tui.Button) {
		client.Play()
		time.Sleep(time.Millisecond * 500)
		updateCurrentlyPlayingLabel(client, currentlyPlayingLabel)
	})

	stopButton.OnActivated(func(*tui.Button) {
		client.Pause()
	})

	previousButton.OnActivated(func(*tui.Button) {
		client.Previous()
		time.Sleep(time.Millisecond * 500)
		updateCurrentlyPlayingLabel(client, currentlyPlayingLabel)
	})

	nextButton.OnActivated(func(*tui.Button) {
		client.Next()
		time.Sleep(time.Millisecond * 500)
		updateCurrentlyPlayingLabel(client, currentlyPlayingLabel)
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
		Play:     playButton,
		Stop:     stopButton,
		Previous: previousButton,
		Next:     nextButton,
		Box:      buttons,
	}
}

func createAvailableDevicesTable(client SpotifyClient, webPlayerID spotify.ID) (*DevicesTable, error) {
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
		table.AppendRow(
			tui.NewLabel(device.Name),
			tui.NewLabel(device.Type),
		)
		// we forced our web player to be the active one, but spotify backend
		// has delays thus, instead of highlighting active device (which might be
		// out of date), we highlight just our web player.
		if device.ID == webPlayerID {
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

	return &DevicesTable{box: tableBox, Table: table}, nil
}

func transferPlaybackToDevice(client SpotifyClient, id spotify.ID) error {
	return client.TransferPlayback(id, true)
}

func getTrackRepr(track *spotify.FullTrack) string {
	return fmt.Sprintf(
		"%s\n%s\n%s",
		track.Name,
		track.Album.Name,
		track.Artists[0].Name,
	)
}
