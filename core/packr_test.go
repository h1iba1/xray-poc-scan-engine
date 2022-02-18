package core

import (
	"fmt"
	"testing"
)

func TestLoadPackPoc(t *testing.T) {
	LoadPackPoc()
	test ,err := PackBox.FindString("74cms-sqli.yml")
	if err != nil {
		Log.Errorf("pack 读取错误%v",err)
	}

	fmt.Printf(test)
}