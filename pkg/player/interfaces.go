package player

import "github.com/zmb3/spotify"

// SpotifyClient is a wrapper interface around spotify.client
// used in order to improve testability of the code.
type SpotifyClient interface {
	UserAlbumFetcher
	Player
	Searcher
	Pause() error
	Previous() error
	Next() error
	PlayerCurrentlyPlaying() (*spotify.CurrentlyPlaying, error)
	PlayerDevices() ([]spotify.PlayerDevice, error)
	TransferPlayback(spotify.ID, bool) error
}

type Player interface {
	Play() error
	PlayOpt(opt *spotify.PlayOptions) error
}

type Searcher interface {
	Search(string, spotify.SearchType) (*spotify.SearchResult, error)
}

type UserAlbumFetcher interface {
	CurrentUsersAlbumsOpt(opt *spotify.Options) (*spotify.SavedAlbumPage, error)
}
