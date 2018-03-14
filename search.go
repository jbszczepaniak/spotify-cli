package main

import (
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
	// "log"
)

type URIName struct {
	URI  spotify.URI
	Name string
}

type searchResults struct {
	table                   *tui.Table
	box                     *tui.Box
	data                    []spotify.URI
	onItemActivatedCallback func(*tui.Table) // Part of the struct in order to test it
}

func NewSearchResults(client SpotifyClient) *searchResults {
	table := tui.NewTable(0, 0)
	data := make([]spotify.URI, 0)
	box := tui.NewVBox(table, tui.NewSpacer())

	box.SetTitle("Songs")
	box.SetBorder(true)

	results := &searchResults{
		table: table,
		box:   box,
		data:  data,
	}

	callback := func(t *tui.Table) {
		selectedRow := t.Selected()
		trackURI := &results.data[selectedRow]
		err := client.PlayOpt(&spotify.PlayOptions{URIs: []spotify.URI{*trackURI}})
		if err != nil {
			err := client.PlayOpt(&spotify.PlayOptions{PlaybackContext: trackURI}) // Fallback to these if previous vall won't work parameters.
			if err != nil {
				panic("Could not get search result")
			}
		}
	}
	table.OnItemActivated(callback)
	results.onItemActivatedCallback = callback
	return results
}

func (sr *searchResults) appendSearchResult(uriName URIName) {
	sr.table.AppendRow(tui.NewLabel(uriName.Name))
	sr.data = append(sr.data, uriName.URI)
}

func (sr *searchResults) resetSearchResults() {
	sr.table.RemoveRows()
	sr.data = sr.data[:0]
}
