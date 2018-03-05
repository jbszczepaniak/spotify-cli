package main

import "testing"
import "os"
import "io"

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
		t.Error("template.Execute should be called with test token", mock.ExecuteData)
	}

	os.Remove("index.html")
}
