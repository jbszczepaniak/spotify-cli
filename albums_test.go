package main

import "testing"

func TestRenderAlbumListPage(t *testing.T) {

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
