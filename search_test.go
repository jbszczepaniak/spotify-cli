package main

import (
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
	"testing"
	// "§fmt"
)

type FakePLayer struct {
	calls int
}

func (fp *FakePLayer) PlayOpt(opt *spotify.PlayOptions) error {
	fp.calls++
	return nil
}

func (fp *FakePLayer) Play() error {
	return nil
}

func TestOnItemActivatedCallbackCallsPlayOpt(t *testing.T) {
	fakePlayer := &FakePLayer{}

	spotifyClient := &DebugClient{Player: fakePlayer}

	results := NewSearchResults(spotifyClient)
	results.table.AppendRow(tui.NewLabel("Name"))
	results.data = append(results.data, "http://local")
	results.onItemActivatedCallback(results.table)

	if fakePlayer.calls != 1 {
		t.Errorf("Should be called 1 time, but it was called %s times", fakePlayer.calls)
	}
}
