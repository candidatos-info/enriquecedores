package main

import (
	"flag"
	"log"
)

func main() {
	source := flag.String("coleta", "", "fonte do arquivo zip")
	outDir := flag.String("outdir", "", "diretório de arquivo zip a ser usado pelo CCE")
	year := flag.Int("ano", 0, "ano da eleição")
	state := flag.String("estado", "", "estado a ser processado")
	flag.Parse()
	if *source != "" {
		if *outDir == "" {
			log.Fatal("informe diretório de saída")
		}
		err := collect(*source, *outDir)
		if err != nil {
			log.Fatal("falha ao executar coleta, erro %q", err)
		}
	} else {
		if *state == "" {
			log.Fatal("informe estado a ser processado")
		}
		if *year == 0 {
			log.Fatal("informe ano a ser processado")
		}
		if *outDir == "" {
			log.Fatal("informe diretório de saída")
		}
		err := process(*state, *outDir, *year)
		if err != nil {
			log.Fatal("falha ao executar enriquecimento, erro %q", err)
		}
	}
}

func collect(source, outDir string) error {
	// TODO implement
	return nil
}

func process(state, outDir string, year int) error {
	// TODO implement
	return nil
}
