package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/candidatos-info/descritor"
	"github.com/golang/protobuf/proto"
)

func TestCollect(t *testing.T) {
	currentPath, err := os.Getwd()
	if err != nil {
		t.Errorf("expected to have err nil when getting current path, got %q", err)
	}
	testFilePath := fmt.Sprintf("file://%s/files_2016.zip", currentPath)
	outputDir := "output"
	if err = os.MkdirAll(outputDir, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", outputDir, err)
	}
	if err := collect(testFilePath, outputDir); err != nil {
		t.Errorf("expected to have err nil when collecting files, got %q", err)
	}
	fileExpectedToExistAfterCollecting := "consulta_cand_2016_RS.csv"
	pathToLookForExpectedFile := fmt.Sprintf("%s/%s/%s", currentPath, outputDir, fileExpectedToExistAfterCollecting)
	if _, err := os.Stat(pathToLookForExpectedFile); err != nil {
		t.Errorf("expected to have err nil when running stat for expected file [%s], got %q", pathToLookForExpectedFile, err)
	}
	if err := os.RemoveAll(outputDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
}

func TestFullProcess(t *testing.T) {
	currentPath, err := os.Getwd()
	if err != nil {
		t.Errorf("expected to have err nil when getting current path, got %q", err)
	}
	testFilePath := fmt.Sprintf("file://%s/files_2016.zip", currentPath)
	outputDir := "output"
	if err = os.MkdirAll(outputDir, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", outputDir, err)
	}
	if err := collect(testFilePath, outputDir); err != nil {
		t.Errorf("expected to have err nil when collecting files, got %q", err)
	}
	stateToTest := "SP"
	dirToSaveCandidatures := "candidatures"
	if err = os.MkdirAll(dirToSaveCandidatures, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", dirToSaveCandidatures, err)
	}
	localCacheDir := "localCache"
	offset := 0
	logFileName := "candidatures_path-2016-SP.csv"
	logFile, err := os.Create(logFileName)
	if err != nil {
		t.Errorf("expected err nil when creating log file, error %v", err)
	}
	defer logFile.Close()
	if err := process(stateToTest, outputDir, dirToSaveCandidatures, localCacheDir, "", "", offset, logFile); err != nil {
		t.Errorf("expected err nil when processing files, got %q", err)
	}
	expectedSequancialCandidateToHave := "260000003557"
	fileToCheckIfExists := fmt.Sprintf("./%s/%s_%s.pb", dirToSaveCandidatures, stateToTest, expectedSequancialCandidateToHave)
	if _, err := os.Stat(fileToCheckIfExists); err != nil {
		t.Errorf("expected to have err nil when running stat for expected file [%s], got %q", fileToCheckIfExists, err)
	}
	bytesToUnmarshal, err := ioutil.ReadFile(fileToCheckIfExists)
	if err != nil {
		t.Errorf("expected err nil when reading bytes of file [%s], got %q", fileToCheckIfExists, err)
	}
	var candidature descritor.Candidatura
	if err = proto.Unmarshal(bytesToUnmarshal, &candidature); err != nil {
		t.Errorf("expected error nil when unmarshaling bytes of file [%s], got %q", fileToCheckIfExists, err)
	}
	exepectedGenderOnSample := "FEMININO"
	expectedVotingIDOnSample := "000412212100"
	expectedEmailOnSample := "PSB@PSBSERGIPE.COM.BR"
	if candidature.Candidato.Email != expectedEmailOnSample {
		t.Errorf("expected to find email [%s], got [%s]", expectedEmailOnSample, candidature.Candidato.Email)
	}
	if candidature.Candidato.TituloEleitoral != expectedVotingIDOnSample {
		t.Errorf("expected to find voting id [%s], got [%s]", expectedVotingIDOnSample, candidature.Candidato.TituloEleitoral)
	}
	if candidature.Candidato.Genero != exepectedGenderOnSample {
		t.Errorf("expected to find gender [%s], got [%s]", exepectedGenderOnSample, candidature.Candidato.Genero)
	}
	if err := os.RemoveAll(outputDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
	if err := os.RemoveAll(dirToSaveCandidatures); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
	expectedFilesToHaveOnLocalCache := map[string]struct{}{
		"SP_260000002646.pb": struct{}{},
		"SP_260000003056.pb": struct{}{},
		"SP_260000003557.pb": struct{}{},
		"SP_260000003630.pb": struct{}{},
		"SP_260000004485.pb": struct{}{},
		"SP_260000005823.pb": struct{}{},
		"SP_260000006327.pb": struct{}{},
		"SP_260000006766.pb": struct{}{},
		"SP_260000006899.pb": struct{}{},
	}
	filesOnLocalCache, err := ioutil.ReadDir(localCacheDir)
	if err != nil {
		t.Errorf("expected err nil when listing files of localCacheDir, got error %v", err)
	}
	for _, file := range filesOnLocalCache {
		if _, ok := expectedFilesToHaveOnLocalCache[file.Name()]; !ok {
			t.Errorf("expected to find file [%s] on local cache dir [%s]", file.Name(), localCacheDir)
		}
	}
	if err := os.RemoveAll(localCacheDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
	expectedLinesToHaveOnOutputFile := map[string]struct{}{
		"candidatures/SP_260000003630.pb,localCache/SP_260000003630.pb": struct{}{},
		"candidatures/SP_260000006766.pb,localCache/SP_260000006766.pb": struct{}{},
		"candidatures/SP_260000003056.pb,localCache/SP_260000003056.pb": struct{}{},
		"candidatures/SP_260000004485.pb,localCache/SP_260000004485.pb": struct{}{},
		"candidatures/SP_260000006899.pb,localCache/SP_260000006899.pb": struct{}{},
		"candidatures/SP_260000002646.pb,localCache/SP_260000002646.pb": struct{}{},
		"candidatures/SP_260000003557.pb,localCache/SP_260000003557.pb": struct{}{},
		"candidatures/SP_260000005823.pb,localCache/SP_260000005823.pb": struct{}{},
		"candidatures/SP_260000006327.pb,localCache/SP_260000006327.pb": struct{}{},
	}
	outputFile, err := os.Open(logFileName)
	if err != nil {
		t.Errorf("expected err nil when opening output file, erro %v", err)
	}
	scanner := bufio.NewScanner(outputFile)
	for scanner.Scan() {
		if _, ok := expectedLinesToHaveOnOutputFile[scanner.Text()]; !ok {
			t.Errorf("expected to have output [%s] on expectedLinesToHaveOnOutputFile", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Errorf("expected err nil scanning file, got %v", err)
	}
	if err := os.RemoveAll(logFileName); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
}

func TestOffset(t *testing.T) {
	currentPath, err := os.Getwd()
	if err != nil {
		t.Errorf("expected to have err nil when getting current path, got %q", err)
	}
	testFilePath := fmt.Sprintf("file://%s/files_2016.zip", currentPath)
	outputDir := "output"
	if err = os.MkdirAll(outputDir, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", outputDir, err)
	}
	if err := collect(testFilePath, outputDir); err != nil {
		t.Errorf("expected to have err nil when collecting files, got %q", err)
	}
	stateToTest := "SP"
	dirToSaveCandidatures := "candidatures"
	if err = os.MkdirAll(dirToSaveCandidatures, 0755); err != nil {
		t.Errorf("expected error nil when creating a test directory [%s], got %q", dirToSaveCandidatures, err)
	}
	localCacheDir := "localCache"
	offset := 3
	logFileName := "candidatures_path-2016-SP.csv"
	logFile, err := os.Create(logFileName)
	if err != nil {
		t.Errorf("expected err nil when creating log file, error %v", err)
	}
	defer logFile.Close()
	if err := process(stateToTest, outputDir, dirToSaveCandidatures, localCacheDir, "", "", offset, logFile); err != nil {
		t.Errorf("expected err nil when processing files, got %q", err)
	}
	if err := os.RemoveAll(outputDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
	if err := os.RemoveAll(dirToSaveCandidatures); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
	expectedFilesToHaveOnLocalCache := map[string]struct{}{
		"SP_260000002646.pb": struct{}{},
		"SP_260000003557.pb": struct{}{},
		"SP_260000004485.pb": struct{}{},
		"SP_260000005823.pb": struct{}{},
		"SP_260000006327.pb": struct{}{},
		"SP_260000006899.pb": struct{}{},
	}
	filesOnLocalCache, err := ioutil.ReadDir(localCacheDir)
	if err != nil {
		t.Errorf("expected err nil when listing files of localCacheDir, got error %v", err)
	}
	for _, file := range filesOnLocalCache {
		if _, ok := expectedFilesToHaveOnLocalCache[file.Name()]; !ok {
			t.Errorf("expected to find file [%s] on local cache dir [%s]", file.Name(), localCacheDir)
		}
	}
	expectedLinesToHaveOnOutputFile := map[string]struct{}{
		"candidatures/SP_260000004485.pb,localCache/SP_260000004485.pb": struct{}{},
		"candidatures/SP_260000006899.pb,localCache/SP_260000006899.pb": struct{}{},
		"candidatures/SP_260000002646.pb,localCache/SP_260000002646.pb": struct{}{},
		"candidatures/SP_260000003557.pb,localCache/SP_260000003557.pb": struct{}{},
		"candidatures/SP_260000005823.pb,localCache/SP_260000005823.pb": struct{}{},
		"candidatures/SP_260000006327.pb,localCache/SP_260000006327.pb": struct{}{},
	}
	outputFile, err := os.Open(logFileName)
	if err != nil {
		t.Errorf("expected err nil when opening output file, erro %v", err)
	}
	scanner := bufio.NewScanner(outputFile)
	for scanner.Scan() {
		if _, ok := expectedLinesToHaveOnOutputFile[scanner.Text()]; !ok {
			t.Errorf("expected to have output [%s] on expectedLinesToHaveOnOutputFile", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Errorf("expected err nil scanning file, got %v", err)
	}
	if err := os.RemoveAll(logFileName); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
	if err := os.RemoveAll(localCacheDir); err != nil {
		t.Errorf("expected erro nil when removing created files")
	}
}
