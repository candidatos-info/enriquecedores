package main

import (
	"flag"
	"log"
)

func main() {
	stateDir := flag.String("inDir", "", "diretório onde as fotos do estado estão")
	candidaturesBucket := flag.String("candidatesDir", "", "bucket onde estão as as candidaturas")
	picturesBucket := flag.String("outDir", "", "bucket onde as fotos devem ser salvas")
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
