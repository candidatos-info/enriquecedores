# Tutorial

Para realizar o enriquecimento de fotos antes é necessario obter os arquivos por meio de alguma fonte como brasil.io. Uma vez de posse das fotos basta rodar o seguinte commando:
```
go run main --inDir=${FOTOS} --destinationDir=${DESTINO}
```

Onde FOTOS é o path para o diretório local onde as fotos estão e DESTINO é o diretório para onde as fotos devem ir, podendo ser um path local para fins de teste ou um bucket no GCS. Um exemplo completo de chamada é o seguinte:
```
go run main --inDir=/Users/user0/candidatos.info/fotos --destinationDir=gs://2016

```
