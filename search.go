package main

import (
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
)

type WithURI struct {
	URI spotify.URI
}

type searchResults struct {
	table *tui.Table
	box   *tui.Box
	data  []WithURI
}

func NewSearchResults(client SpotifyClient) *searchResults {
	table := tui.NewTable(0, 0)
	data := make([]WithURI, 0)
	box := tui.NewVBox(table, tui.NewSpacer())

	box.SetTitle("Songs")
	box.SetBorder(true)

	results := &searchResults{
		table: table,
		box:   box,
		data:  data,
	}

	table.OnItemActivated(func(t *tui.Table) {
		selectedRow := t.Selected()
		trackURI := &results.data[selectedRow].URI
		err := client.PlayOpt(&spotify.PlayOptions{URIs: []spotify.URI{*trackURI}})
		if err != nil {
			err := client.PlayOpt(&spotify.PlayOptions{PlaybackContext: trackURI})
			if err != nil {
				panic("nie ma")
			}
		}
	})
	return results
}
