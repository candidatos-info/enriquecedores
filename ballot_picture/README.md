# README

Para realizar o enriquecimento de fotos antes é necessario obter os arquivos por meio de alguma fonte como brasil.io (https://brasil.io/dataset/eleicoes-brasil/candidatos/). Uma vez de posse das fotos basta rodar o seguinte commando:
```
go run main --inDir=${FOTOS} --outDir=${DESTINO}
```

Onde FOTOS é o path para o diretório local onde as fotos estão e DESTINO é o diretório para onde as fotos devem ir, podendo ser um path local para fins de teste ou um bucket no GCS. Vale citar o fato de que os arquivos de fotos e candidaturas ficam armazenados no mesmo bucket/diretoriório (se for o caso de um diretório deve-se informar o path local, e no caso de um bucket deve-se informar informar o nome do bucket form prefixo gs://). Um exemplo completo de chamada é o seguinte:
```
go run main --inDir=/Users/user0/candidatos.info/fotos --outDir=gs://2016
```
