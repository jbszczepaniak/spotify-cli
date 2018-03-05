package main

import "testing"
import "os"
import "io"
import "net/http/httptest"
import "github.com/zmb3/spotify"

type TemplateMock struct {
	ExecuteCount int
	ExecuteData  interface{}
}

func (tm *TemplateMock) Execute(wr io.Writer, data interface{}) error {
	tm.ExecuteCount++
	tm.ExecuteData = data
	return nil
}

func TestInsertTokenToTemplate(t *testing.T) {
	mock := TemplateMock{}

	insertTokenToTemplate("test token", &mock)

	if mock.ExecuteCount != 1 {
		t.Error("template.Execute should be called once. It was called", mock.ExecuteCount, "times.")
	}

	if mock.ExecuteData.(tokenToInsert).Token != "test token" {
		t.Error("template.Execute should be called with test token. It was called with: ", mock.ExecuteData)
	}

	os.Remove("index.html")
}

func TestAuthCallBack(t *testing.T) {
	ch = make(chan *spotify.Client)

	go func() {
		client := <-ch
		if token, _ := client.Token(); token.AccessToken == "123" {
			t.Error("asd")
		}

	}()
	authCallback(&httptest.ResponseRecorder{}, httptest.NewRequest("GET", "/", nil))
}
