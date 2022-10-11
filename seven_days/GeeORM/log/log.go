package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// [info ] 颜色为蓝色，[error] 为红色。
// 使用 log.Lshortfile 支持显示文件名和代码行号。
// 暴露 Error，Errorf，Info，Infof 4个方法。

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

// log methods
var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

// log levels
const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

// SetLevel controls log level
// 这一部分的实现非常简单，三个层级声明为三个常量，通过控制 Output，来控制日志是否打印。
// 如果设置为 ErrorLevel，infoLog 的输出会被定向到 ioutil.Discard，即不打印该日志。
func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if ErrorLevel < level {
		errorLog.SetOutput(ioutil.Discard)
	}
	if InfoLevel < level {
		infoLog.SetOutput(ioutil.Discard)
	}
}
