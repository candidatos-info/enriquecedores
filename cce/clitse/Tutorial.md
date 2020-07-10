# Tutorial

Para coletar os arquivos necessários ao enriquecimento use o seguinte comando:
```
$ go run cli.go --coleta=${URL} --outdir=${OUTDIR}
```

Onde a URL é a fonte dos arquivos .zip, pondendo ser passado uma URL usando protocolo HTTP(S) ou file, e OUTDIR é o path do diretório para onde os arquivos coletados serão colocados.

Após a etapa de coleta devemos executar o enriquecimento. Para executar o enriquecimento use o seguinte comando:
```
go run cli.go --estado=${ESTADO} --ano=${ANO} --outdir=${OUTDIR}
```

Onde Estado é o código UF (SIGLA) do estado a ser enriquecido (seguindo a tabela abaixo), ANO é o ano da eleição a ser processadada e OUTDIR o diretório usado na etapa de coleta.

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
