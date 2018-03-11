package main

import "testing"
import "github.com/zmb3/spotify"

func TestGetTrackRepr(t *testing.T) {
	var tests = []struct {
		track *spotify.FullTrack
		repr  string
	}{
		{
			&spotify.FullTrack{
				spotify.SimpleTrack{
					Name:    "Name",
					Artists: []spotify.SimpleArtist{{Name: "art1"}, {Name: "art2"}},
				},
				spotify.SimpleAlbum{},
				nil,
				0,
			}, "Name (art1, art2)",
		},
		{
			&spotify.FullTrack{
				spotify.SimpleTrack{
					Name:    "Name",
					Artists: []spotify.SimpleArtist{{Name: "art"}},
				},
				spotify.SimpleAlbum{},
				nil,
				0,
			}, "Name (art)",
		},
	}
	for _, test := range tests {
		got := getTrackRepr(test.track)
		if got != test.repr {
			t.Errorf("Got: %v, want: %v", got, test.repr)
		}
	}
}
