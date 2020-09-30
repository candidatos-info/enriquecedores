# Tutorial

Para executar o detector de candidatura recorrentes execute os seguintes comandos:
```
$ go build
$ ./recurrence-detector -dbName=${DB_NAME} -dbURL=${DB_URL} -prevCandidaturesSheet=${PREV_CANDIDATURES_PATH} -currentCandidaturesSheet=${CURRENT_CANDIDATURES_PATH}
```

Onde:
+ DB_NAME é o nome do banco;
+ DB_URL é URL de conexão com o banco;
+ PREV_CANDIDATURES_PATH é o path para o arquivo csv das candidaturas antigas;
+ CURRENT_CANDIDATURES_PATH é o path para o arquivo csv das candidaturas atuais;

```
$ ./recurrence-detector -dbName=candidatos -dbURL=mongodb://localhost:27017/candidatos -prevCandidaturesSheet=/Users/user0/Downloads/consulta_cand_2016/consulta_cand_2016_AL.csv -currentCandidaturesSheet=/Users/user0/Downloads/consulta_cand_2020/consulta_cand_2020_AL.csv
```
