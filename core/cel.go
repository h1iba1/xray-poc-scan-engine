package core

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"xray-poc-scan-engine/utils"
)



var mapStrStr =decls.NewMapType(decls.String,decls.String)

type UrlType struct {
	Scheme               string   `protobuf:"bytes,1,opt,name=scheme,proto3" json:"scheme,omitempty"`
	Domain               string   `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	Host                 string   `protobuf:"bytes,3,opt,name=host,proto3" json:"host,omitempty"`
	Port                 string   `protobuf:"bytes,4,opt,name=port,proto3" json:"port,omitempty"`
	Path                 string   `protobuf:"bytes,5,opt,name=path,proto3" json:"path,omitempty"`
	Query                string   `protobuf:"bytes,6,opt,name=query,proto3" json:"query,omitempty"`
	Fragment             string   `protobuf:"bytes,7,opt,name=fragment,proto3" json:"fragment,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type Reverse struct {
	Url                  *UrlType `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Domain               string   `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	Ip                   string   `protobuf:"bytes,3,opt,name=ip,proto3" json:"ip,omitempty"`
	IsDomainNameServer   bool     `protobuf:"varint,4,opt,name=is_domain_name_server,json=isDomainNameServer,proto3" json:"is_domain_name_server,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

var defaultDecls=[]*exprpb.Decl{
	// NewIdent 创建具有可选文字值的命名标识符声明。文字值通常仅与枚举标识符相关联。
	decls.NewIdent("request.url.scheme", decls.String, nil),
	decls.NewIdent("request.url.domain", decls.String, nil),
	decls.NewIdent("request.url.host", decls.String, nil),
	decls.NewIdent("request.url.port", decls.String, nil),
	decls.NewIdent("request.url.path", decls.String, nil),
	decls.NewIdent("request.url.query", decls.String, nil),
	decls.NewIdent("request.url.fragment", decls.String, nil),

	decls.NewIdent("response.status", decls.Int, nil),
	decls.NewIdent("response.content_type", decls.String, nil),
	decls.NewIdent("response.body", decls.Bytes, nil),
	decls.NewIdent("response.headers", mapStrStr, nil),
	decls.NewIdent("response.url", decls.String, nil),


	decls.NewFunction("bcontains", decls.NewInstanceOverload("bytes_bcontains_bytes",
			[]*exprpb.Type{decls.Bytes, decls.Bytes}, decls.Bool)),
	decls.NewFunction("bmatches",
		decls.NewInstanceOverload("string_bmatches_bytes",
			[]*exprpb.Type{decls.String, decls.Bytes},
			decls.Bool)),
	decls.NewFunction("md5",
		decls.NewOverload("md5_string",
			[]*exprpb.Type{decls.String},
			decls.String)),
	decls.NewFunction("randomInt",
		decls.NewOverload("randomInt_int_int",
			[]*exprpb.Type{decls.Int, decls.Int},
			decls.Int)),
	decls.NewFunction("randomLowercase",
		decls.NewOverload("randomLowercase_int",
			[]*exprpb.Type{decls.Int},
			decls.String)),
	decls.NewFunction("base64",
		decls.NewOverload("base64_string",
			[]*exprpb.Type{decls.String},
			decls.String)),
	decls.NewFunction("base64",
		decls.NewOverload("base64_bytes",
			[]*exprpb.Type{decls.Bytes},
			decls.String)),
	decls.NewFunction("base64Decode",
		decls.NewOverload("base64Decode_string",
			[]*exprpb.Type{decls.String},
			decls.String)),
	decls.NewFunction("base64Decode",
		decls.NewOverload("base64Decode_bytes",
			[]*exprpb.Type{decls.Bytes},
			decls.String)),
	decls.NewFunction("urlencode",
		decls.NewOverload("urlencode_string",
			[]*exprpb.Type{decls.String},
			decls.String)),
	decls.NewFunction("urlencode",
		decls.NewOverload("urlencode_bytes",
			[]*exprpb.Type{decls.Bytes},
			decls.String)),
	decls.NewFunction("urldecode",
		decls.NewOverload("urldecode_string",
			[]*exprpb.Type{decls.String},
			decls.String)),
	decls.NewFunction("urldecode",
		decls.NewOverload("urldecode_bytes",
			[]*exprpb.Type{decls.Bytes},
			decls.String)),
	decls.NewFunction("substr",
		decls.NewOverload("substr_string_int_int",
			[]*exprpb.Type{decls.String, decls.Int, decls.Int},
			decls.String)),
	decls.NewFunction("wait",
		decls.NewInstanceOverload("reverse_wait_int",
			[]*exprpb.Type{decls.Any, decls.Int},
			decls.Bool)),
	decls.NewFunction("icontains",
		decls.NewInstanceOverload("icontains_string",
			[]*exprpb.Type{decls.String, decls.String},
			decls.Bool)),
}

// fuction实现
var defaultFunctions=[]*functions.Overload{
	{
		Operator: "bytes_bcontains_bytes",
		Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
			v1, ok := lhs.(types.Bytes)
			if !ok {
				return types.ValOrErr(lhs, "unexpected type '%v' passed to bcontains", lhs.Type())
			}
			v2, ok := rhs.(types.Bytes)
			if !ok {
				return types.ValOrErr(rhs, "unexpected type '%v' passed to bcontains", rhs.Type())
			}
			return types.Bool(bytes.Contains(v1, v2))
		},
	},
	{
		Operator: "string_bmatch_bytes",
		Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
			v1, ok := lhs.(types.String)
			if !ok {
				return types.ValOrErr(lhs, "unexpected type '%v' passed to bmatch", lhs.Type())
			}
			v2, ok := rhs.(types.Bytes)
			if !ok {
				return types.ValOrErr(rhs, "unexpected type '%v' passed to bmatch", rhs.Type())
			}
			ok, err := regexp.Match(string(v1), v2)
			if err != nil {
				return types.NewErr("%v", err)
			}
			return types.Bool(ok)
		},
	},
	{
		Operator: "md5_string",
		Unary: func(value ref.Val) ref.Val {
			v, ok := value.(types.String)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to md5_string", value.Type())
			}
			return types.String(fmt.Sprintf("%x", md5.Sum([]byte(v))))
		},
	},
	{
		Operator: "randomInt_int_int",
		Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
			from, ok := lhs.(types.Int)
			if !ok {
				return types.ValOrErr(lhs, "unexpected type '%v' passed to randomInt", lhs.Type())
			}
			to, ok := rhs.(types.Int)
			if !ok {
				return types.ValOrErr(rhs, "unexpected type '%v' passed to randomInt", rhs.Type())
			}
			min, max := int(from), int(to)
			return types.Int(rand.Intn(max-min) + min)
		},
	},
	{
		Operator: "randomLowercase_int",
		Unary: func(value ref.Val) ref.Val {
			n, ok := value.(types.Int)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to randomLowercase", value.Type())
			}
			return types.String(utils.RandomLowercase(int(n)))
		},
	},
	{
		Operator: "base64_string",
		Unary: func(value ref.Val) ref.Val {
			v, ok := value.(types.String)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to base64_string", value.Type())
			}
			return types.String(base64.StdEncoding.EncodeToString([]byte(v)))
		},
	},
	{
		Operator: "base64_bytes",
		Unary: func(value ref.Val) ref.Val {
			v, ok := value.(types.Bytes)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to base64_bytes", value.Type())
			}
			return types.String(base64.StdEncoding.EncodeToString(v))
		},
	},
	{
		Operator: "base64Decode_string",
		Unary: func(value ref.Val) ref.Val {
			v, ok := value.(types.String)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to base64Decode_string", value.Type())
			}
			decodeBytes, err := base64.StdEncoding.DecodeString(string(v))
			if err != nil {
				return types.NewErr("%v", err)
			}
			return types.String(decodeBytes)
		},
	},
	{
		Operator: "base64Decode_bytes",
		Unary: func(value ref.Val) ref.Val {
			v, ok := value.(types.Bytes)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to base64Decode_bytes", value.Type())
			}
			decodeBytes, err := base64.StdEncoding.DecodeString(string(v))
			if err != nil {
				return types.NewErr("%v", err)
			}
			return types.String(decodeBytes)
		},
	},
	{
		Operator: "urlencode_string",
		Unary: func(value ref.Val) ref.Val {
			v, ok := value.(types.String)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to urlencode_string", value.Type())
			}
			return types.String(url.QueryEscape(string(v)))
		},
	},
	{
		Operator: "urlencode_bytes",
		Unary: func(value ref.Val) ref.Val {
			v, ok := value.(types.Bytes)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to urlencode_bytes", value.Type())
			}
			return types.String(url.QueryEscape(string(v)))
		},
	},
	{
		Operator: "urldecode_string",
		Unary: func(value ref.Val) ref.Val {
			v, ok := value.(types.String)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to urldecode_string", value.Type())
			}
			decodeString, err := url.QueryUnescape(string(v))
			if err != nil {
				return types.NewErr("%v", err)
			}
			return types.String(decodeString)
		},
	},
	{
		Operator: "urldecode_bytes",
		Unary: func(value ref.Val) ref.Val {
			v, ok := value.(types.Bytes)
			if !ok {
				return types.ValOrErr(value, "unexpected type '%v' passed to urldecode_bytes", value.Type())
			}
			decodeString, err := url.QueryUnescape(string(v))
			if err != nil {
				return types.NewErr("%v", err)
			}
			return types.String(decodeString)
		},
	},
	{
		Operator: "substr_string_int_int",
		Function: func(values ...ref.Val) ref.Val {
			if len(values) == 3 {
				str, ok := values[0].(types.String)
				if !ok {
					return types.NewErr("invalid string to 'substr'")
				}
				start, ok := values[1].(types.Int)
				if !ok {
					return types.NewErr("invalid start to 'substr'")
				}
				length, ok := values[2].(types.Int)
				if !ok {
					return types.NewErr("invalid length to 'substr'")
				}
				runes := []rune(str)
				if start < 0 || length < 0 || int(start+length) > len(runes) {
					return types.NewErr("invalid start or length to 'substr'")
				}
				return types.String(runes[start : start+length])
			} else {
				return types.NewErr("too many arguments to 'substr'")
			}
		},
	},
	{
		Operator: "icontains_string",
		Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
			v1, ok := lhs.(types.String)
			if !ok {
				return types.ValOrErr(lhs, "unexpected type '%v' passed to bcontains", lhs.Type())
			}
			v2, ok := rhs.(types.String)
			if !ok {
				return types.ValOrErr(rhs, "unexpected type '%v' passed to bcontains", rhs.Type())
			}
			// 不区分大小写包含
			return types.Bool(strings.Contains(strings.ToLower(string(v1)), strings.ToLower(string(v2))))
		},
	},
}

// 创建新的标识
func NewCelEnv(options map[string]ref.Val) (cel.Env, error) {
	// 创建decl切片
	addDecls := make([]*exprpb.Decl, 0, len(defaultDecls)+len(options))
	// 将defaultDecl添加到decls中
	addDecls = append(addDecls, defaultDecls...)

	// 遍历options 创建标识
	for k, v := range options {
		var c *exprpb.Decl
		// 根据类型创建标识
		switch v.Type() {
		case types.StringType:
			c = decls.NewIdent(k, decls.String, nil)
		case types.IntType:
			c = decls.NewIdent(k, decls.Int, nil)
		case types.UintType:
			c = decls.NewIdent(k, decls.Uint, nil)
		case types.BoolType:
			c = decls.NewIdent(k, decls.Bool, nil)
		case types.BytesType:
			c = decls.NewIdent(k, decls.Bytes, nil)
		case types.DoubleType:
			c = decls.NewIdent(k, decls.Double, nil)
		default:
			c = decls.NewIdent(k, decls.String, nil)
		}

		addDecls = append(addDecls, c)
	}

	/*
		NewEnv 创建一个 Env 实例，适用于针对一组用户定义的常量、变量和函数解析和检查表达式。
		默认情况下启用宏和标准内置函数。
		有关可用于配置环境的选项，请参阅 EnvOptions。
	*/
	// Declarations 声明选项扩展了在环境中配置的声明集。注意：如果两者一起使用，则必须在 ClearBuiltIns 之后指定此选项。
	return cel.NewEnv(cel.Declarations(addDecls...))
}

func NewCelFunctions(dnsPrefix string) cel.ProgramOption {
	// 创建新的functions切片 并将defaultFunctions添加到fs切片
	fs := make([]*functions.Overload, 0, len(defaultFunctions)+1)
	fs = append(fs, defaultFunctions...)
	//
	fs = append(fs,
		&functions.Overload{
			Operator: "waitReverse",
			Unary: func(value ref.Val) ref.Val {
				timeout, ok := value.Value().(int64)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to waitReverse", value.Type())
				}
				return types.Bool(waitReverse(dnsPrefix, timeout))
			},
			Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
				randStr, ok := lhs.Value().(string)
				if !ok {
					return types.ValOrErr(lhs, "unexpected string '%v' passed to waitReverse", lhs.Type())
				}
				timeout, ok := rhs.Value().(int64)
				if !ok {
					return types.ValOrErr(rhs, "unexpected int '%v' passed to waitReverse", rhs.Type())
				}
				return types.Bool(waitReverse(randStr, timeout))
			},
		})

	return cel.Functions(fs...)
}

// 等待dns反连
func waitReverse(prefix string, timeout int64) bool {
	return false
}