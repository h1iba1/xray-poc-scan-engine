package core

import (
	"fmt"
	"testing"
)

func TestParsePocByData(t *testing.T) {
	file :="/Users/h11ba1/Desktop/go/xray-poc-scan-engine/test.yml"
	pocYaml ,_:= ParsePocByFile(file)

	fmt.Println(pocYaml.GetDetail().Links)
}
