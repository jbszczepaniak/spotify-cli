package main

import "testing"
import "os"
import "io"

type TemplateMock struct {
	ExecuteCount int
}

func (tm *TemplateMock) Execute(wr io.Writer, data interface{}) error {
	tm.ExecuteCount++
	return nil
}

func TestInsertTokenToTemplate(t *testing.T) {
	tm := TemplateMock{}

	insertTokenToTemplate("token", &tm)

	if tm.ExecuteCount != 1 {
		t.Error("template.Execute should be called once. It was called", tm.ExecuteCount, "times.")
	}

	os.Remove("index.html")
}
