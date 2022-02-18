package core

import (
	"github.com/google/cel-go/common/types/ref"
	"reflect"
	"testing"
)

func TestCelFunc(t *testing.T) {
	//生成celEnv
	celEnv, err := NewCelEnv(nil)
	if err != nil {
		t.Error(err)
		return
	}
	// 解析输入生成pAst
	pAst, iss := celEnv.Parse("randomLowercase(4)")
	if iss != nil && iss.Err() != nil {
		t.Error(iss.Err())
		return
	}
	// 检查生成的pAst，生成cAst
	cAst, iss := celEnv.Check(pAst)
	if iss != nil && iss.Err() != nil {
		t.Error(iss.Err())
		return
	}
	// 程序在环境中生成Ast的可评估实例(Env)
	prg, err := celEnv.Program(cAst, NewCelFunctions(""))
	if err != nil {
		t.Error(err)
		return
	}

	var setTypes map[string]ref.Val

	// Eval返回Ast和环境对输入vars的求值结果。
	rv, _, _ :=prg.Eval(map[string]interface{}{})
	t.Log(reflect.TypeOf(rv))


	setTypes["1"]=rv

	//for k,v :=range setTypes{
	//	fmt.Println(k)
	//	fmt.Println(v.Value())
	//}

}











