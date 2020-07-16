package main

import (
	"flag"
	"log"
)

func main() {
	stateDir := flag.String("statedir", "", "diretório onde as fotos do estado estão")
	candidaturesBucket := flag.String("candidatesbucket", "", "bucket onde estão as as candidaturas")
	picturesBucket := flag.String("picturesbucket", "", "bucket onde as fotos devem ser salvas")
	flag.Parse()
	if *stateDir == "" {
		log.Fatal("informe o diretório onde as fotos do estão estão")
	}
	if *candidaturesBucket == "" {
		log.Fatal("informe o bucket onde as candidaturas estão")
	}
	if *picturesBucket == "" {
		log.Fatal("informe o bucket onde as fotos devem ser salvs")
	}
	// TODO implementar script de enriquecimento
}
