# README

Para realizar o enriquecimento de fotos antes é necessario obter os arquivos por meio de alguma fonte como brasil.io (https://brasil.io/dataset/eleicoes-brasil/candidatos/). Uma vez de posse das fotos basta rodar o seguinte commando:
```
go run main.go -inDir=${IN_DIR} -outDir=${OUT_DIR} -year=${YEAR} -state=${STATE} -handledPicturesFile=${HANDLED_PICTURES_FILE} -outputFile=${OUTPUT_FILE}
```

Onde:
+ IN_DIR é o path para o diretório onde as fotos estão localmente;
+ OUT_DIR é o ID da pasta no Google Drive onde as fotos serão enviadas;
+ YEAR é o ano da eleição;
+ STATE é o estado da eleição;
+ HANDLED_PICTURES_FILE é o path para um arquivo que contem o nome das fotos já processadas;
+ OUTPUT_FILE é o path para o arquivo que será usado pelo resumidor do banco de dados. Caso você não passe esse parâmetro o programa criará um automaticamente. Caso uma execução termine em erro vocé deverá rodar novamente passando o mesmo arquivo como parâmetro;

Um exemplo completo de chamada é o seguinte:

```
go run main.go -inDir=/Users/abuarquemf/Downloads/2016/AL -outDir=13v-RiEaq7ZFV1nnQVkEeBaDnUMKH-2Us -year=2020 -state=AL 
```
