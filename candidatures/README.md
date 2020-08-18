# Tutorial

Para coletar os arquivos necessários ao enriquecimento use o seguinte comando:
```
$ go run cli.go --sourceFile=${URL} --localDir=${OUTDIR}
```

Onde a URL é a fonte dos arquivos .zip, pondendo ser passado uma URL usando protocolo HTTP(S) ou file, e OUTDIR é o path do diretório para onde os arquivos coletados serão colocados localmente.

Após a etapa de coleta devemos executar o enriquecimento. Para executar o enriquecimento use o seguinte comando:
```
go run main.go --outDir=${BUCKET} --state=${ESTADO} --localDir=${OUTDIR}
```

Onde: 
+ BUCKET é o destino final dos arquivos de candidaturas processado, podendo ser um path local ou um bucket do GCS;
+ ESTADO é o código UF (SIGLA) do estado a ser enriquecido (seguindo a tabela abaixo);
+ OUTDIR o diretório usado na etapa de coleta;

Um exemplo completo de chamada é o seguinte:

```
go run main.go --outDir=gs://2016 --state=AL --localDir=/Users/user0/candidatos.info/enriquecedores/candidatures/cli/temp
```

| Estado | Sigla |
|:--:|:--|
|ACRE|AC|
|Alagoas|AL|
|Amazonas|AM|
|Amapá|AP|
|Bahia|BA|
|Ceará|CE|
|Espírito Santo|ES|
|Goias|GO|
|Maranhão|MA|
|Minas Gerais|MG|
|Mato Grosso do Sul|MS|    
|Mato Grosso|MT|
|Pará|PA|
|Paraíba|PB|
|Pernambuco|PE|
|Piauí|PI|
|Parana|PR|
|Rio de Janeiro|RJ|
|Rio Grande do Norte|RN|
|Rorâima|RO|
|Rio Grande do Sul|RS|
|Rondônia|RO|
|Roraima|RR|
|Santa Catarina|SC|
|São Paulo|SP|
|Sergipe|SE|
|Tocantins|TO|
