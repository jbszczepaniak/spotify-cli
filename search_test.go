package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
)

func TestNewSearch(t *testing.T) {
	client := &DebugClient{}
	search := NewSearch(client)
	if len(search.focusables) != 4 {
		t.Fatalf("Expected to have 4 focusables elements, got %d", len(search.focusables))
	}
	if (search.box.Length()) != 2 {
		t.Fatalf("Expected to have 2 elements in search box, got %d", search.box.Length())
	}
}

type FakePlayer struct {
	calls                     int
	playOptErrCallWithURI     bool
	playOptErrCallWithContext bool
}

func (fp *FakePlayer) PlayOpt(opt *spotify.PlayOptions) error {
	fp.calls++

	if fp.playOptErrCallWithContext && opt.PlaybackContext != nil {
		return fmt.Errorf("")
	}
	if fp.playOptErrCallWithURI && opt.URIs != nil {
		return fmt.Errorf("")
	}

	return nil
}

func (fp *FakePlayer) Play() error {
	return nil
}

func TestOnItemActivatedCallback(t *testing.T) {
	cases := []struct {
		errCallWithURI          bool
		errCallWithContext      bool
		expectedPlayOptNumCalls int
		expectedLogs            string
	}{
		{
			errCallWithURI:          true,
			errCallWithContext:      true,
			expectedPlayOptNumCalls: 2,
			expectedLogs:            "Could not play searched URI: some:spotify:uri\n",
		},
		{
			errCallWithURI:          false,
			errCallWithContext:      false,
			expectedPlayOptNumCalls: 1,
			expectedLogs:            "Successfuly played searched URI: some:spotify:uri\n",
		},
		{
			errCallWithURI:          true,
			errCallWithContext:      false,
			expectedPlayOptNumCalls: 2,
			expectedLogs:            "Successfuly played searched URI: some:spotify:uri\n",
		},
		{
			errCallWithURI:          false,
			errCallWithContext:      true,
			expectedPlayOptNumCalls: 1,
			expectedLogs:            "Successfuly played searched URI: some:spotify:uri\n",
		},
	}

	for _, c := range cases {
		var str bytes.Buffer
		log.SetOutput(&str)

		client := &DebugClient{}
		fakePlayer := &FakePlayer{
			playOptErrCallWithURI:     c.errCallWithURI,
			playOptErrCallWithContext: c.errCallWithContext,
		}
		client.Player = fakePlayer

		results := NewSearchResults(client, "Results")

		results.table.AppendRow(tui.NewLabel("Name"))
		results.data = append(results.data, "some:spotify:uri")

		results.onItemActivatedCallback(results.table)

		if !strings.HasSuffix(str.String(), c.expectedLogs) {
			t.Errorf("Expect log to have suffix %s but log was %s", c.expectedLogs, str.String())
		}

		if fakePlayer.calls != c.expectedPlayOptNumCalls {
			t.Errorf("Should be called %d times, but was called %d times", c.expectedPlayOptNumCalls, fakePlayer.calls)
		}
	}
}

func TestAppendRemoveSearchResults(t *testing.T) {
	client := &DebugClient{}
	results := NewSearchResults(client, "Results")
	testUriName := URIName{URI: "test:spotify:uri", Name: "Test Name"}

	results.appendSearchResult(testUriName)
	if resultsItemsCount := len(results.data); resultsItemsCount != 1 {
		t.Fatal("Expect results to have 1 item, but results have %d items", resultsItemsCount)
	}

	results.resetSearchResults()
	if resultsItemsCount := len(results.data); resultsItemsCount != 0 {
		t.Fatal("Expect results to have 0 item, but results have %d items", resultsItemsCount)
	}
}
