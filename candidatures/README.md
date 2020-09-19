# Tutorial

Para coletar os arquivos necessários ao enriquecimento use o seguinte comando:
```
$ go run cli.go --sourceFile=${URL} --localDir=${OUTDIR}
```

Onde a URL é a fonte dos arquivos .zip, pondendo ser passado uma URL usando protocolo HTTP(S) ou file, e OUTDIR é o path do diretório para onde os arquivos coletados serão colocados localmente.

Após a etapa de coleta devemos executar o enriquecimento. Para executar o enriquecimento use o seguinte comando:
```
go run main.go -candidaturesDir=${CANDIDATURES_DIR} -localCache=${LOCAL_CACHE} -credentials=${CREDENTIALS_FILE} -OAuthToken=${OAUTH_TOKEN_FILE} -year=${YEAR} -state=${STATE} -localDir=${OUTDIR} --offset=${OFFSET} --outputFile=${OUTPUT_FILE}
```

Onde: 
+ CANDIDATURES_DIR é o ID da pasta no Google Drive para onde os arquivos serão enviados;
+ LOCAL_CACHE é o path de um diretório para salvar os arquivos de candidaturas localmente;
+ CREDENTIALS_FILE é o path até o o arquivo de credenciais do Google Drive;
+ OAUTH_TOKEN_FILE é o path até o arquivo do token OAuth para o Google Drive;
+ YEAR é o ano da eleição;
+ STATE é a sigla do estado para ser processado (VEJA A TABELA ABAIXO COM AS SIGLAS DOS ESTADOS);
+ OUTDIR é o diretório onde os arquivos CSV foram colocados;
+ OFFSET é um inteiro que indica a linha de processamento que o programa deve começar a rodar. Pode ocorrer de um processamento terminar com erro no meio do caminho, e essa flag será printada via STDOUT para que você rode novamente o programa de onde parou;
+ OUTPUT_FILE é o path para o arquivo de logs. Caso não seja passado sera criado um novo;

Um exemplo completo de chamada é o seguinte:

```
go run main.go -candidaturesDir=505102302j02s10sj26969 -localCache=/Users/user0/candidatos.info/enriquecedoes/candidatures/cache -credentials=/Users/user0/candidatos.info/enriquecedores/credentials.json -OAuthToken=/Users/user0/candidatos.info/enriquecedores/token.json -year=2016 -state=AL -localDir=/Users/user0/Downloads/consulta_cand_2016/ --offset=0 -outputFile=/Users/user0/candidatos.info/enriquecedores/candidatures/candidatures_path-2020-AL.csv
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
