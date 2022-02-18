package main

import (
	"xray-poc-scan-engine/cmd"
	"xray-poc-scan-engine/core"
)

func init() {
	err := core.LoadPackPoc()
	if err != nil{
		core.Log.Errorf("load poc error %v",err)
	}
	core.LoadAllYamlPoc()
}

func main()  {
	cmd.RunCmd()
}
