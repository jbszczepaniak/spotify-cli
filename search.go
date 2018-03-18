package main

import (
	"log"

	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
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

type search struct {
	focusables []tui.Widget
	box        *tui.Box
}

func searchInputOnSubmit(client SpotifyClient, searchedSongs, searchedAlbums, searchedArtists *searchResults) func(*tui.Entry) {
	return func(entry *tui.Entry) {
		result, _ := client.Search(
			entry.Text(),
			spotify.SearchTypeAlbum|spotify.SearchTypeTrack|spotify.SearchTypeArtist,
		)

		searchedAlbums.resetSearchResults()
		for _, i := range result.Albums.Albums {
			searchedAlbums.appendSearchResult(URIName{Name: i.Name, URI: i.URI})
		}

		searchedSongs.resetSearchResults()
		for _, i := range result.Tracks.Tracks {
			searchedSongs.appendSearchResult(URIName{Name: i.Name, URI: i.URI})
		}

		searchedArtists.resetSearchResults()
		for _, i := range result.Artists.Artists {
			searchedArtists.appendSearchResult(URIName{Name: i.Name, URI: i.URI})
		}

	}
}

func NewSearch(client SpotifyClient) *search {
	searchedSongs := NewSearchResults(client, "Songs")
	searchedAlbums := NewSearchResults(client, "Albums")
	searchedArtists := NewSearchResults(client, "Artists")

	searchInput := tui.NewEntry()
	searchInput.SetSizePolicy(tui.Preferred, tui.Minimum)
	searchInput.OnSubmit(searchInputOnSubmit(client, searchedSongs, searchedAlbums, searchedArtists))

	searchInputBox := tui.NewHBox(searchInput, tui.NewSpacer())
	searchInputBox.SetTitle("Search")
	searchInputBox.SetBorder(true)

	searchResults := tui.NewVBox(searchedSongs.box, searchedAlbums.box, searchedArtists.box)
	searchResults.SetTitle("Search Results")
	searchResults.SetBorder(true)

	return &search{
		focusables: []tui.Widget{searchInput, searchedSongs.table, searchedAlbums.table, searchedArtists.table},
		box:        tui.NewVBox(searchInputBox, searchResults),
	}

}

func NewSearchResults(client SpotifyClient, name string) *searchResults {
	table := tui.NewTable(0, 0)
	data := make([]spotify.URI, 0)
	box := tui.NewVBox(table, tui.NewSpacer())

	box.SetTitle(name)
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
				log.Printf("Could not play searched URI: %s\n", *trackURI)
				return
			}
		}
		log.Printf("Successfuly played searched URI: %s\n", *trackURI)
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
