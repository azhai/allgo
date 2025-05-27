package main

import (
	"fmt"
	"net/http"

	"github.com/alexflint/go-arg"
	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/logutil"
	"github.com/azhai/allgo/services/db"
	"github.com/azhai/allgo/services/log"
)

var args struct {
	GenModel *GenModelCmd `arg:"subcommand:model" help:"生成Model文件"`
	Verbose  bool         `arg:"-v,--verbose" help:"输出详细信息"`
	ServerOpts
}

// ServerOpts 服务配置
type ServerOpts struct {
	Host string `arg:"-s,--host" default:"" help:"运行IP"`     // 运行IP
	Port int    `arg:"-p,--port" default:"8080" help:"运行端口"` // 运行端口
}

// GetServerAddr 获取服务地址
func (t ServerOpts) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

func init() {
	arg.MustParse(&args)
}

func main() {
	env := config.Prepare(256)
	if err := log.OpenService(env); err != nil {
		panic(err)
	}
	defer log.CloseService()
	if err := db.OpenService(env); err != nil {
		panic(err)
	}
	defer db.CloseService()

	if args.GenModel != nil {
		args.GenModel.Run()
		return
	}

	addr := args.GetServerAddr()
	logutil.Infof("Server start at %s ...", addr)
	mux := http.NewServeMux()
	mux.HandleFunc("/", Homepage)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logutil.Fatal(err)
	}
}
