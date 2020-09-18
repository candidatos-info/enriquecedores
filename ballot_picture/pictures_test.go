package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/candidatos-info/enriquecedores/filestorage"
)

func TestProcessWithPictureHavingCorrespondentCandidate(t *testing.T) {
	picturesDir, err := ioutil.TempDir("", "picturesDir")
	if err != nil {
		t.Errorf("expected error nil when creating temporary dir for pictures, erro %v", err)
	}
	defer os.RemoveAll(picturesDir)
	fakeSequencialCandidate := "260000003557"
	fakePictureFile := fmt.Sprintf("%s/%s.jpg", picturesDir, fakeSequencialCandidate)
	dumbMessage := "dumb file"
	if err := ioutil.WriteFile(fakePictureFile, []byte(dumbMessage), 0644); err != nil {
		t.Errorf("expected error nil when create file [%s] to simulate a picture file, got %q", fakePictureFile, err)
	}
	storageDir, err := ioutil.TempDir("", "storageDir")
	if err != nil {
		t.Errorf("expected error nil when creating temporary dir for storage, erro %v", err)
	}
	defer os.RemoveAll(storageDir)
	fakeCandidatureFileName := fmt.Sprintf("%s/%s.zip", storageDir, fakeSequencialCandidate)
	if err := ioutil.WriteFile(fakeCandidatureFileName, []byte(dumbMessage), 0644); err != nil {
		t.Errorf("expected error nil when create file [%s] to simulate a candidature file, got %q", fakeCandidatureFileName, err)
	}
	logFileName := "log_file.txt"
	logFile, err := ioutil.TempFile("", logFileName)
	if err != nil {
		t.Errorf("expected err nil when creating log file, go %v", err)
	}
	defer os.Remove(logFile.Name())
	defer logFile.Close()
	if err := process(picturesDir, storageDir, filestorage.NewLocalStorage(), logFile); err != nil {
		t.Errorf("expected error nil when running process, error %q", err)
	}
	expectedFileToFindOnStorageDirAfterProcess := fmt.Sprintf("%s/%s_%s.jpg", storageDir, filepath.Base(picturesDir), fakeSequencialCandidate)
	_, err = os.Stat(expectedFileToFindOnStorageDirAfterProcess)
	if err != nil {
		t.Errorf("expected error nil when running stat on expected file, got %q", err)
	}
	if err := os.RemoveAll(logFileName); err != nil {
		t.Errorf("expected erro nil when removing log file, erro %v", err)
	}
}
