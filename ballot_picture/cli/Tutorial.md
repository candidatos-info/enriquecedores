# Tutorial

Para ativar o enriquecimento de fotos use o seguinte comando:
```
go run main.go --inDir=${IN_DIR} --candidatesDir=${CANDIDATES_DIR} --outDir=${OUTPUT_DIR}
```

Onde IN_DIR é o diretório onde as fotos do estado estão, CANDIDATES_DIR o local onde os arquivos de candidaturas estão e OUTPUT_DIR o local onde as fotos devem ser salvas, estes dois últimos podendo ser um path local ou de um bucket no GCS.

Um exemplo completo de chamada é o seguinte:
```
go run main.go --inDir=/Users/abuarquemf/Downloads/2016/SE --candidatesDir=gs://2016 --outDir=gs://2016
```
