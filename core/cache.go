package core

import (
	"sort"
)

var pocMapCache =make(map[string]*PocYaml)

// 根据name获取PocYaml
func GetPocYaml(name string) *PocYaml {
	if v,ok:= pocMapCache[name];ok{
		return v
	}
	return nil
}

// 获取内存所有pocName
func GetAllPocYamlName()[]string  {
	result :=make([]string,0,len(pocMapCache))
	for name:=range pocMapCache {
		result=append(result,name)
	}
	sort.Strings(result)
	return result
}

// 获取内存中所有poc
func GetAllPocYaml()map[string]*PocYaml {
	return pocMapCache
}

// 从文件中加载poc到pocMapCache
func LoadYamlPoc(file string)(string,error)  {
	yml,err:= ParsePocByFile(file)
	if err ==nil{
		pocMapCache[yml.Name]=yml
		return yml.Name,nil
	}
	return "",err
}

func LoadYamlPocByData(data []byte)error  {
	yml,err:= ParsePocByData(data)
	if err ==nil{
		pocMapCache[yml.Name]=yml
	}
	return err
}

