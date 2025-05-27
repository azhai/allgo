package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var MB = 1024 * 1024 // 1MB

// Prepare 初始化环境，包括日志、配置等
func Prepare(size int) *Environ {
	if IsRunTest() {
		return NewWithFile(".env.testing")
	}
	if size > 0 { // 压舱石，阻止频繁GC
		ballast := make([]byte, size*MB)
		runtime.KeepAlive(ballast)
	}
	return New()
}

// IsRunTest 是否测试模式下
func IsRunTest() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

// BackToDir 退回上层目录
func BackToDir(back int) (err error) {
	if back == 0 {
		return
	} else if back < 0 {
		back = 0 - back
	}
	dir := strings.Repeat("../", back)
	if dir, err = filepath.Abs(dir); err == nil {
		err = os.Chdir(dir)
	}
	return
}

// SetUsage 使用帮助
func SetUsage(tpl string, args ...any) {
	out := flag.CommandLine.Output()
	_, _ = fmt.Fprintf(out, tpl, args...)
	flag.PrintDefaults()
}
