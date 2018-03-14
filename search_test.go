package main

import (
	"bytes"
	"fmt"
	"github.com/marcusolsson/tui-go"
	"github.com/zmb3/spotify"
	"log"
	"strings"
	"testing"
)

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
		errCallWithURI     bool
		errCallWithContext bool
		expectedCalls      int
		expectedLogs       string
	}{
		{
			errCallWithURI:     true,
			errCallWithContext: true,
			expectedCalls:      2,
			expectedLogs:       "Could not play searched URI: some:spotify:uri\n",
		},
		{
			errCallWithURI:     false,
			errCallWithContext: false,
			expectedCalls:      1,
			expectedLogs:       "Successfuly played searched URI: some:spotify:uri\n",
		},
		{
			errCallWithURI:     true,
			errCallWithContext: false,
			expectedCalls:      2,
			expectedLogs:       "Successfuly played searched URI: some:spotify:uri\n",
		},
		{
			errCallWithURI:     false,
			errCallWithContext: true,
			expectedCalls:      1,
			expectedLogs:       "Successfuly played searched URI: some:spotify:uri\n",
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

		results := NewSearchResults(client)
		results.table.AppendRow(tui.NewLabel("Name"))
		results.data = append(results.data, "some:spotify:uri")
		results.onItemActivatedCallback(results.table)

		if !strings.HasSuffix(str.String(), c.expectedLogs) {
			t.Errorf("Expect log to have suffix %s but log was %s", c.expectedLogs, str.String())
		}

		if fakePlayer.calls != c.expectedCalls {
			t.Errorf("Should be called %d times, but was called %d times", c.expectedCalls, fakePlayer.calls)
		}
	}
}
