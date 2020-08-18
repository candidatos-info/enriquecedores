package main

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestCollect(t *testing.T) {
	currentPath, err := os.Getwd()
	if err != nil {
		t.Errorf("expected to have err nil when getting current path, got %q", err)
	}
	filesTestSource := fmt.Sprintf("file://%s/files_2016.zip", currentPath)
	outDir := "files"
	if err = os.MkdirAll(outDir, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", outDir, err)
	}
	if err := collect(filesTestSource, outDir); err != nil {
		t.Errorf("exepceted error nil when collecting test files, got %q", err)
	}
	expectedFileToFindAfterCollecting := fmt.Sprintf("%s/consulta_cand_2016_RR.csv", outDir)
	if _, err := os.Stat(expectedFileToFindAfterCollecting); err != nil {
		t.Errorf("expected to have err nil when running stat for expected file [%s], got %q", expectedFileToFindAfterCollecting, err)
	}
	if err := os.RemoveAll(outDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
}

func TestProcess(t *testing.T) {
	currentPath, err := os.Getwd()
	if err != nil {
		t.Errorf("expected to have err nil when getting current path, got %q", err)
	}
	filesTestSource := fmt.Sprintf("file://%s/files_2016.zip", currentPath)
	outDir := "files"
	if err = os.MkdirAll(outDir, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", outDir, err)
	}
	if err := collect(filesTestSource, outDir); err != nil {
		t.Errorf("exepceted error nil when collecting test files, got %q", err)
	}
	stateToTest := "SP"
	candidaturesRepository := newInMemoryRepository()
	if err := process(outDir, stateToTest, candidaturesRepository); err != nil {
		log.Fatalf("falha ao processar dados para enriquecimento do banco, erro %q", err)
	}
	emailExpectedToBeOnDB := "VEREADORPOLAQUE@HOTMAIL.COM"
	vc, err := candidaturesRepository.findCandidateByEmail(emailExpectedToBeOnDB)
	if err != nil {
		t.Errorf("expected err nil when looking for email [%s] on fake db, got %q", emailExpectedToBeOnDB, err)
	}
	if vc == nil {
		t.Errorf("expected to have found voting city, we've got a nil")
	}
	if err := os.RemoveAll(outDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
}
