package filestorage

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestUpload(t *testing.T) {
	content := "o amor é um fogo que arde sem se ver, é ferida que dói e não se sente, é dor que desatina sem doer"
	fileStorage := NewLocalStorage()
	dir := "dir"
	fileName := "aname.txt"
	path, err := fileStorage.Upload([]byte(content), dir, fileName)
	if err != nil {
		t.Errorf("expected erro nil when writing a file, got %q", err)
	}
	fileContent, err := ioutil.ReadFile(path)
	if err != nil {
		t.Errorf("expected err nil when reading file, got %q", err)
	}
	if content != string(fileContent) {
		t.Errorf("expected content to be \"%s\", got %s", content, string(fileContent))
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
}
