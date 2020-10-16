package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/candidatos-info/enriquecedores/filestorage"
)

func TestProcessWithPictureHavingCorrespondentCandidate(t *testing.T) {
	picturesDir, err := ioutil.TempDir("", "picturesDir")
	if err != nil {
		t.Errorf("expected error nil when creating temporary dir for pictures, erro %v", err)
	}
	defer os.RemoveAll(picturesDir)
	fakeSequencialCandidate := "260000003557.jpg"
	fakePictureFile := fmt.Sprintf("%s/%s", picturesDir, fakeSequencialCandidate)
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
	picturesCache, err := ioutil.TempFile("", "picturesCache")
	if err != nil {
		t.Errorf("expected err nil when creating pictures cache file, go %v", err)
	}
	defer os.Remove(picturesCache.Name())
	defer logFile.Close()
	source := "tse"
	if err := process(source, picturesDir, storageDir, filestorage.NewLocalStorage(), logFile, picturesCache); err != nil {
		t.Errorf("expected error nil when running process, error %q", err)
	}
	contentOfCacheFile, err := ioutil.ReadFile(picturesCache.Name())
	if err != nil {
		t.Errorf("expected err nil when reading cache file, got %v", err)
	}
	if !strings.Contains(string(contentOfCacheFile), fakeSequencialCandidate) {
		t.Errorf("expected to find the name of fake file on cache file")
	}
}
