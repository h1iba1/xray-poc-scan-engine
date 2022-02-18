package core

import (
	"fmt"
	"testing"
)

func TestLoadAllYamlPoc(t *testing.T) {
	err := LoadPackPoc()
	if err!=nil{}
	LoadAllYamlPoc()
	for _, yaml := range GetAllPocYaml() {
		fmt.Println(yaml.Name)
	}
}

func TestGetPocYaml(t *testing.T) {
	err := LoadPackPoc()
	if err!=nil{}
	LoadAllYamlPoc()

	fmt.Println(GetPocYaml("poc-yaml-yapi-rce").Name)
}