package core

import (
	"github.com/gobuffalo/packr/v2"
)



var PackBox *packr.Box

// 将./pocs目录下的poc加载到PackBox
func LoadPackPoc() error {
	PackBox =packr.New("poc","./pocs")
	return nil
}

func getPackFileBody(file string)([]byte,error){
	body,err:= PackBox.Find(file)
	if err!=nil{
		Log.Errorf("getPackFileBody:%v",err)
	}
	return body,nil
}

// 将PackBox中的所有poc加载到内存
func LoadAllYamlPoc()  {
	files := PackBox.List()
	for _,v:=range files{
		body,err := getPackFileBody(v)
		if err !=nil{
			Log.Errorf("LoadAllYamlPoc:%v",err)
			continue
		}
		err = LoadYamlPocByData(body)
		if err !=nil{
			Log.Errorf("LoadYamlPocByData:%v",err)
		}
	}
}






