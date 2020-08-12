package main

import (
	"fmt"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	d := "28/09/1972"
	nascimentoAsTime, err := time.Parse("02/01/2006", d)
	if err != nil {
		t.Errorf("falha ao fazer parse de data de nascimento de candidato %s para time.Time, erro %q", d, err)
	}
	fmt.Println(nascimentoAsTime)

}
