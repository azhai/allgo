package logutil_test

import (
	"testing"
	"time"

	"github.com/azhai/allgo/logutil"
)

var (
	sinkURL = "rotate://../logs/access.log?level=info&min=1&comp=0"
	logger  = logutil.NewLoggerURL(sinkURL)
)

func NowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Test11Info(t *testing.T) {
	for i := 1; i <= 20000; i++ {
		logger.Infof("999 888 %05d", i)
	}
	logger.Errorf("now is %s", NowTime())
	// assert.NoError(t, err)
}
