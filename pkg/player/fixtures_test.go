package player

import (
	"testing"

	"github.com/zmb3/spotify"
)

func TestFixturesForDebugMode(t *testing.T) {
	debugClient := NewDebugClient()

	err := debugClient.PlayOpt(&spotify.PlayOptions{})
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}

	err = debugClient.Play()
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}

	albumsPage, err := debugClient.CurrentUsersAlbumsOpt(&spotify.Options{})
	expectedAlbumsCount := 135 // 3*pageSize
	if len(albumsPage.Albums) != expectedAlbumsCount {
		t.Errorf("Expected to have %d fake albums, have %d", expectedAlbumsCount, len(albumsPage.Albums))
	}
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}

	err = debugClient.Previous()
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}

	err = debugClient.Pause()
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}

	err = debugClient.Next()
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}

	_, err = debugClient.PlayerCurrentlyPlaying()
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}

	devices, err := debugClient.PlayerDevices()
	expectedDevicesCount := 3
	if len(devices) != expectedDevicesCount {
		t.Errorf("Expected to have %d fake devices, have %d", expectedDevicesCount, len(devices))
	}
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}

	err = debugClient.TransferPlayback("id", true)
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}

	// _, err = debugClient.CurrentUser()
	// if err != nil {
	// 	t.Errorf("Expected not to return error, but got %v", err)
	// }

	// _, err = debugClient.Token()
	// if err != nil {
	// 	t.Errorf("Expected not to return error, but got %v", err)
	// }

	_, err = debugClient.Search("query", spotify.SearchTypeArtist)
	if err != nil {
		t.Errorf("Expected not to return error, but got %v", err)
	}
}
