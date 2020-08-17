package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestProcessWithPictureHavingCorrespondentCandidate(t *testing.T) {
	picturesDir := "pictures"
	if err := os.MkdirAll(picturesDir, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", picturesDir, err)
	}
	fakeSequencialCandidate := "260000003557"
	fakePictureFile := fmt.Sprintf("%s/%s.jpg", picturesDir, fakeSequencialCandidate)
	dumbMessage := "dumb file"
	if err := ioutil.WriteFile(fakePictureFile, []byte(dumbMessage), 0644); err != nil {
		t.Errorf("expected error nil when create file [%s] to simulate a picture file, got %q", fakePictureFile, err)
	}
	storageDir := "candidatures"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", storageDir, err)
	}
	fakeCandidatureFileName := fmt.Sprintf("%s/%s.zip", storageDir, fakeSequencialCandidate)
	if err := ioutil.WriteFile(fakeCandidatureFileName, []byte(dumbMessage), 0644); err != nil {
		t.Errorf("expected error nil when create file [%s] to simulate a candidature file, got %q", fakeCandidatureFileName, err)
	}
	if err := process(picturesDir, storageDir, nil); err != nil {
		t.Errorf("expected error nil when running process, error %q", err)
	}
	expectedFileToFindOnStorageDirAfterProcess := fmt.Sprintf("%s/%s.jpg", storageDir, fakeSequencialCandidate)
	_, err := os.Stat(expectedFileToFindOnStorageDirAfterProcess)
	if err != nil {
		t.Errorf("expected error nil when running stat on expected file, got %q", err)
	}
	if err := os.RemoveAll(storageDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
	if err := os.RemoveAll(picturesDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
}

func TestProcessWithPictureNotHavingCorrespondentCandidate(t *testing.T) {
	picturesDir := "pictures"
	if err := os.MkdirAll(picturesDir, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", picturesDir, err)
	}
	fakeSequencialCandidate := "260000003557"
	fakePictureFile := fmt.Sprintf("%s/%s.jpg", picturesDir, fakeSequencialCandidate)
	dumbMessage := "dumb file"
	if err := ioutil.WriteFile(fakePictureFile, []byte(dumbMessage), 0644); err != nil {
		t.Errorf("expected error nil when create file [%s] to simulate a picture file, got %q", fakePictureFile, err)
	}
	storageDir := "candidatures"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", storageDir, err)
	}
	logFileName := "logs.txt"
	logErrorFile, err := os.Create(logFileName)
	if err != nil {
		t.Errorf("exepected error nil when creating log file error, got %q", err)
	}
	defer logErrorFile.Close()
	if err := process(picturesDir, storageDir, logErrorFile); err != nil {
		t.Errorf("expected error nil when running process, error %q", err)
	}
	bytes, err := ioutil.ReadFile(logFileName)
	if err != nil {
		t.Errorf("expected err nil when reading the content of log file error, got %q", err)
	}
	if !strings.Contains(string(bytes), fakeSequencialCandidate) {
		t.Errorf("expected to find sequencial candidate [%s] on file [%s], but it wasn`t", fakeSequencialCandidate, logFileName)
	}
	if err := os.RemoveAll(storageDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
	if err := os.RemoveAll(logFileName); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
	if err := os.RemoveAll(picturesDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
}
