package main

import (
	"fmt"
	"testing"

	"github.com/marcusolsson/tui-go"

	"github.com/zmb3/spotify"
)

func TestNewSideBar(t *testing.T) {
	client := NewDebugClient()
	sideBar, err := NewSideBar(client)
	if err != nil {
		t.Fatalf("Unexpected error occured: %s", err)
	}
	if len(sideBar.albumList.albumsDescriptions) != 135 {
		// Because DebugClient's implementation of CurrentUsersAlbumsOpt fetches 135 Spotify Albums
		t.Fatalf("Should fetch 135 album descripitons, fetched %d", len(sideBar.albumList.albumsDescriptions))
	}
}

type CallConfig struct {
	executionError bool
	returnValue    *spotify.SavedAlbumPage
}

type AlbumFetcherMock struct {
	call        int
	callConfigs []CallConfig
}

func (fake *AlbumFetcherMock) CurrentUsersAlbumsOpt(opt *spotify.Options) (*spotify.SavedAlbumPage, error) {
	if fake.callConfigs[fake.call].executionError == true {
		fake.call++
		return nil, fmt.Errorf("err")
	}
	returnValue := fake.callConfigs[fake.call].returnValue
	fake.call++
	return returnValue, nil
}

func TestFetchUserAlbumListFetchesNoPages(t *testing.T) {
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}
	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: false,
			returnValue:    &spotify.SavedAlbumPage{Albums: make([]spotify.SavedAlbum, 0)},
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	albumList.fetchUserAlbums()

	if len(albumList.albumsDescriptions) != 0 {
		t.Fatalf("Expected albums descriptions to be empty, but it has length of %d", len(albumList.albumsDescriptions))
	}
}

func TestFetchUserAlbumListFetchesSinglePage(t *testing.T) {
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}

	saved := &spotify.SavedAlbumPage{Albums: constructNSpotifySavedAlbums(25)}
	saved.Total = 25 // Only one page

	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: false,
			returnValue:    saved,
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	albumsDescriptions, err := albumList.fetchUserAlbums()
	if err != nil {
		t.Fatalf("Did not expect to fail, but it did")
	}
	if len(albumsDescriptions) != 25 {
		t.Fatalf("Expected albums descriptions to have 25 elements, but have %d elements", len(albumsDescriptions))
	}
	if fetcherMock.call != 1 {
		t.Fatalf("Expected CurrentUsersAlbumsOpt() to be called once, but it was called %d times", fetcherMock.call)
	}
}
func TestFetchUserAlbumListFetchesManyPages(t *testing.T) {
	defer func() { spotifyAPIPageOffset = 25 }() // Reset after test
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}

	saved := &spotify.SavedAlbumPage{Albums: constructNSpotifySavedAlbums(25)}
	saved.Total = 50

	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: false,
			returnValue:    saved,
		},
		{
			executionError: false,
			returnValue:    saved,
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	albumsDescriptions, err := albumList.fetchUserAlbums()
	if err != nil {
		t.Fatalf("Did not expect to fail, but it did")
	}
	if len(albumsDescriptions) != 50 {
		t.Fatalf("Expected albums descriptions to have 50 elements, but have %d elements", len(albumsDescriptions))
	}
	if fetcherMock.call != 2 {
		t.Fatalf("Expected CurrentUsersAlbumsOpt() to be called twice, but it was called %d times", fetcherMock.call)
	}
}
func TestFetchUserAlbumListFailsOnFirstCall(t *testing.T) {
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}

	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: true,
			returnValue:    nil,
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	_, err := albumList.fetchUserAlbums()
	if err == nil {
		t.Fatalf("Expected to fail, but it didn't")
	}
	if fetcherMock.call != 1 {
		t.Fatalf("Expected CurrentUsersAlbumsOpt() to be called once, but it was called %d times", fetcherMock.call)
	}
}

func TestFetchUserAlbumListFailsWhenFetchingNotFirstPage(t *testing.T) {
	defer func() { spotifyAPIPageOffset = 25 }() // Reset after test
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}

	saved := &spotify.SavedAlbumPage{Albums: constructNSpotifySavedAlbums(25)}
	saved.Total = 50

	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: false,
			returnValue:    saved,
		},
		{
			executionError: true,
			returnValue:    nil,
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	_, err := albumList.fetchUserAlbums()
	if err == nil {
		t.Fatalf("Expected to fail, but it didn't")
	}
	if fetcherMock.call != 2 {
		t.Fatalf("Expected CurrentUsersAlbumsOpt() to be called twice, but it was called %d times", fetcherMock.call)
	}
}

type fakeDataFetcher struct {
	ExecutionError bool
}

func (fake *fakeDataFetcher) fetchUserAlbums() ([]albumDescription, error) {
	if fake.ExecutionError == true {
		return nil, fmt.Errorf("error")
	}
	return []albumDescription{{artist: "Artist", title: "Title", uri: "uri"}}, nil
}

type fakePageRenderer struct {
	ExecutionError bool
}

func (fake *fakePageRenderer) renderPage([]albumDescription, int, int) error {
	if fake.ExecutionError == true {
		return fmt.Errorf("error")
	}
	return nil
}

func TestRenderFailsWhenFetchingUserAlbumsFail(t *testing.T) {
	albumList := &AlbumList{}
	albumList.DataFetcher = &fakeDataFetcher{ExecutionError: true}
	err := albumList.render()
	if err == nil {
		t.Fatalf("Expected to fail but it didn't")
	}
}

func TestRenderFailsWhenPageRenderingFail(t *testing.T) {
	albumList := &AlbumList{}
	albumList.DataFetcher = &fakeDataFetcher{ExecutionError: false}
	albumList.PageRenderer = &fakePageRenderer{ExecutionError: true}
	err := albumList.render()
	if err == nil {
		t.Fatalf("Expected to fail but it didn't")
	}
}

func TestRenderSucceds(t *testing.T) {
	albumList := &AlbumList{
		DataFetcher:  &fakeDataFetcher{ExecutionError: false},
		PageRenderer: &fakePageRenderer{ExecutionError: false},
		table:        &tui.Table{},
	}
	err := albumList.render()
	if err != nil {
		t.Fatalf("Did not expect to fail but it did with %#v", err)
	}
	// Here it should test that table has row added, but I don't see a way to this (query for table size would be nice)
}

func TestNextPage(t *testing.T) {
	albumList := &AlbumList{table: &tui.Table{}}
	cases := []struct {
		lastTwoSelected    []int
		shouldOpenNextPage bool
	}{
		{[]int{44, 45}, true}, // We are on last item, we should go to next page
		{[]int{0, 1}, false},  // We are on the beginning on the list, we should not go to next page
	}
	for _, c := range cases {
		albumList.lastTwoSelected = c.lastTwoSelected
		if shouldOpenNextPage := albumList.nextPage(); shouldOpenNextPage != c.shouldOpenNextPage {
			t.Fatalf("Got %v, but wanted %v for next page", shouldOpenNextPage, c.shouldOpenNextPage)
		}
	}
}

func TestPreviousPage(t *testing.T) {
	albumList := &AlbumList{table: &tui.Table{}}
	cases := []struct {
		lastTwoSelected        []int
		selectedTableItem      int
		currDataIdx            int
		shouldOpenPreviousPage bool
	}{
		{[]int{0, 1}, 0, 46, true}, // Only conditions where next page will be displayed
		{[]int{44, 45}, 0, 1, false},
		{[]int{0, 1}, 1, 46, false},
		{[]int{0, 1}, 0, 0, false},
	}
	for _, c := range cases {
		albumList.lastTwoSelected = c.lastTwoSelected
		albumList.currDataIdx = c.currDataIdx       // Pretend this is current data index
		albumList.table.Select(c.selectedTableItem) // Pretend that this item is selected
		if shouldOpenPreviousPage := albumList.previousPage(); shouldOpenPreviousPage != c.shouldOpenPreviousPage {
			t.Fatalf("Got %v, but wanted %v for previous page", shouldOpenPreviousPage, c.shouldOpenPreviousPage)
		}
	}
}

func TestTrimCommasIfTooLong(t *testing.T) {
	text := "Some text"
	cases := []struct {
		length         int
		expectedResult string
	}{
		{
			len(text),
			"Some text",
		},
		{
			len(text) - 1,
			"Some tex...",
		},
		{
			len(text) + 1,
			"Some text",
		},
	}
	for _, c := range cases {
		if result := trimWithCommasIfTooLong(text, c.length); result != c.expectedResult {
			t.Fatalf("Expected result to be %s, but it was %s", c.expectedResult, result)
		}
	}

}
