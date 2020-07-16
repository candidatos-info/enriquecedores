# Tutorial

Para coletar os arquivos necessários ao enriquecimento use o seguinte comando:
```
$ go run cli.go --coleta=${URL} --outdir=${OUTDIR}
```

Onde a URL é a fonte dos arquivos .zip, pondendo ser passado uma URL usando protocolo HTTP(S) ou file, e OUTDIR é o path do diretório para onde os arquivos coletados serão colocados.

Após a etapa de coleta devemos executar o enriquecimento. Para executar o enriquecimento use o seguinte comando:
```
go run cli.go go run cli.go -estado=${ESTADO} -ano=${ANO} -outdir=${OUTDIR} -remoteadd=${REMOTE_ADD} -cceadd=${CCE_ADD} -username=${USERNAME} -password=${PASSWORD}
```

Onde Estado é o código UF (SIGLA) do estado a ser enriquecido (seguindo a tabela abaixo), ANO é o ano da eleição a ser processadada, OUTDIR o diretório usado na etapa de coleta, REMOTE_ADD é o endereço o qual o mini servidor do cli está exposto, CCE_ADD é o endereço do CCE, USERNAME e PASSWORD são os parâmetros de basic auth. Um exemplo completo de chamada é o seguinte:
```
go run cli.go -estado=TO -ano=2016 -outdir=./baseDir -remoteadd=https://2f307e515c14.ngrok.io -cceadd=http://localhost:8877/cce -username=cands -password=pass
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
