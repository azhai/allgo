package log

import (
	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/logutil"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

// OpenService 根据配置初始化日志单例
func OpenService(env *config.Environ) error {
	if logFile := env.Get("ACCESS_LOG_FILE"); logFile != "" {
		logger = logutil.NewLoggerURL("info", logFile)
	} else if logDir := env.Get("LOG_DIR"); logDir != "" {
		logger = logutil.NewLogger(logDir)
	} else {
		logger = zap.NewNop().Sugar()
	}
	logutil.SetLogger(logger)
	return nil
}

// CloseService 关闭服务
func CloseService() {
	if logger != nil {
		_ = logger.Sync()
		logger = nil
	}
}
