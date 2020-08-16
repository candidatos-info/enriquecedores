package main

import (
	"fmt"
	"io/ioutil"
	"os"
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
