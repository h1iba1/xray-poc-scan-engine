package cmd

import (
	"xray-poc-scan-engine/core"
)


func init() {
	err := core.LoadPackPoc()
	if err != nil{
		core.Log.Errorf("load poc error %v",err)
	}
	core.LoadAllYamlPoc()
}

//func TestRunPoc(t *testing.T) {
//	target :="https://h11ba1.com"
//	pocName :="xss"
//	RunPoc(target,pocName)
//}