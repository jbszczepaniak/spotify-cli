package main

import (
	"github.com/marcusolsson/tui-go"
	"log"
)

type Album struct {
	artist string
	title  string
}

var albums []Album

func main() {
	client := authenticate()

	salbums, err := client.CurrentUsersAlbums()

	if err != nil {
		log.Fatal(err)
	}

	for _, album := range salbums.Albums {
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
	albumsList.Select(3)

	box := tui.NewVBox(
		albumsList,
		tui.NewSpacer(),
	)

	box.SetTitle("SPOTIFY CLI")

	ui, err := tui.New(box)
	if err != nil {
		panic(err)
	}

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		panic(err)
	}
}
