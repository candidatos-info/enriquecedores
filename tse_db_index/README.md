# README

A primeira etapa do enriquecimento é a de coleta dos arquivos de candidatura do TSE que podem ser obtidos por este link: http://www.tse.jus.br/eleicoes/estatisticas/repositorio-de-dados-eleitorais-1/repositorio-de-dados-eleitorais. De posse da fonte de dados das candidaturas basta rodar o seguinte comando:

```
go run main.go --localDir=${DIRETORIO} --collect=${URL}
```

Onde DIRETORIO é o diretório de saída onde os arquivos descomprimidos serão colocados localmente e URL é fonte do arquivo zip do TSE com as candidaturas. OBS: a implementação deste enriquecedor suporta tanto procolo HTTP(s) quanto file;

Um exemplo concreto de coleta usando protocolo file:

```
go run main.go -localDir=. -collect=file:///Users/abuarquemf/candidatos.info/enri/candidatures/files_2016.zip
```

Uma vez que os arquivos CSV foram baixados pode-se executar o enriquecimento do banco com o seguinte comando:
```
go run main.go --projectID=${PROJECT_ID} --localDir=${DIRETORIO} --state=${ESTADO} 
```

Onde PROJECT_ID é o ID do projeto no GCP e em caso de testes deve ser passada uma string vazia, DIRETORIO é o local onde foram baixados os arquivos CSV do TSE e ESTADO é o estado a ser trabalhado no enriquecimento. Um exemplo de uso desse cli é o seguinte:

```
go run main.go --projectID=505 --localDir=. --state=AL
```
