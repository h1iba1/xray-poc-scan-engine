package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strconv"

	"github.com/alexeyco/simpletable"
	"regexp"
	"xray-poc-scan-engine/core"
)


var (
	Target  string
	PocName string
	FilePath string
)

var (
	Poc string
)

var (
	Pocs []string
	targets []string
)

var rootCmd = &cobra.Command{
	Use: "xray_poc",
	Short: "xray poc扫描器",
	Long: `xray poc规则发生了一次变化,导致之前的poc扫描器不能使用，故重新写一版。
			xray v2版本的poc规则:https://docs.xray.cool/#/guide/poc/v2`,
}

var scanCmd =&cobra.Command{
	Use: "scan",
	Short: "xray_poc scan poc扫描模块",
	Long :`xray poc 扫描器：
	Author:h11ba1/https://github.com/h1iba1
	scan -h :
	--u 指定扫描目标
	--poc 指定需要调用的poc。不指定poc默认使用所有poc
	--f 批量扫描。`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if PocName !="" {
			Pocs = getPocByName(PocName)
		}else {
			Pocs = core.GetAllPocYamlName()
			core.Log.Info("all poc")
		}

		if FilePath != "" {
			targets =LoadFile(FilePath)
		}

		if Target!=""{
			targets=append(targets,Target)
		}

	},
	Run: func(cmd *cobra.Command, args []string){
		var reqs []*core.PocRequest

		// 多个poc的情况
		// 多个target
		for _,pocName :=range Pocs{
			for _,target :=range targets{
				req := &core.PocRequest{
					URL: target,
					YmlName: pocName,
					Headers: http.Header{},
					Method: "get",
					PostData:"",
					Port:80,
				}
				reqs =append(reqs,req)
			}
		}

		userAgent :=[]string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169"}
		httpConfig := core.HttpConfig{
			DialTimeout:     5,
			ReadTimeout:     30,
			FailRetries:     1,
			MaxQPS:          500,
			MaxRedirect:     5,
			MaxConnsPerHost: 50,
			MaxRespBodySize: 8388608,
			Headers:         core.HeaderConfig{UserAgent: userAgent},
		}
		execPoc(reqs,httpConfig)
	},
}

var pocCmd = &cobra.Command{
	Use: "poc",
	Short: "xray_poc search poc查看模块",
	Long: `xray poc 扫描器：
	Author:h11ba1/https://github.com/h1iba1
	poc -h:
	--search 搜索存在的poc，支持模糊匹配。如：xray_poc poc --search "thinkphp"`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if Poc!=""{
			Pocs =getPocByName(Poc)
		}else{
			core.Log.Errorf("get poc by name err")
		}
	},
	Run: func(cmd *cobra.Command, args []string){
		table := simpletable.New()
		table.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: "#"},
				{Align: simpletable.AlignLeft, Text: "Poc Name"},
			},
		}

		for k, v := range Pocs {
			r := []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d", k)},
				{Align: simpletable.AlignLeft, Text: v},
			}

			table.Body.Cells = append(table.Body.Cells, r)
		}
		table.SetStyle(simpletable.StyleCompactLite)
		fmt.Println(table.String())
	},
}

func RunCmd()  {
	rootCmd.AddCommand(scanCmd,pocCmd)

	scanCmd.Flags().StringVar(&Target,"u","","xray_poc scan target")
	scanCmd.Flags().StringVar(&PocName,"poc","","xray_poc scan poc")
	scanCmd.Flags().StringVar(&FilePath,"f","","xray_poc scan filePath")

	pocCmd.Flags().StringVar(&Poc,"search","","xray_poc poc search")

	err := rootCmd.Execute()
	if err != nil {
		core.Log.Errorf("%v",err)
	}
}

func getPocByName(pocName string)[]string  {
	var result []string

	reg, err := regexp.Compile(pocName)
	pocNames := core.GetAllPocYamlName()
	if err != nil {
		core.Log.Errorf("search regexp syntax error: %v", err)
	}
	for _, v := range pocNames {
		if reg.MatchString(v) {
			result = append(result, v)
		}
	}
	return result
}


func execPoc(reqs []*core.PocRequest,httpConfig core.HttpConfig)  {

	var pocResults []*core.PocResult

	pocExecuteManager := core.NewPocExecuteManager(httpConfig)

	// 批量执行
	for _,req :=range reqs{
		if pocResult,err :=pocExecuteManager.DoScan(req); err ==nil {
			pocResults =append(pocResults,pocResult)
		}else if err !=nil {
			core.Log.Errorf("DoScan req exec err %v",err)
		}
	}

	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "#"},
			{Align: simpletable.AlignLeft, Text: "target-url"},
			{Align: simpletable.AlignLeft, Text: "poc-name"},
			{Align: simpletable.AlignLeft, Text: "status"},
		},
	}

	for k,pocResult :=range pocResults{
		r := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d", k)},
			{Align: simpletable.AlignLeft, Text: pocResult.Target.URL},
			{Align: simpletable.AlignLeft, Text: pocResult.PocInfo.PocName},
			{Align: simpletable.AlignLeft, Text: strconv.FormatBool(pocResult.Vulnerable)},
		}
		table.Body.Cells = append(table.Body.Cells, r)
	}
	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
}

// 读取文件 返回url列表
func LoadFile(file string) []string {
	var result []string
	f, err := os.Open(file)
	if err != nil {
		core.Log.Errorf("load file error %v", err)
		return result
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	return result
}