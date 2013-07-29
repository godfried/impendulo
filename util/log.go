package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var errLogger, infoLogger *SyncLogger

//init sets up the loggers.
func init() {
	logDir := filepath.Join(BaseDir(), "logs")
	y, m, d := time.Now().Date()
	dir := filepath.Join(logDir, strconv.Itoa(y), m.String(), strconv.Itoa(d))
	err := os.MkdirAll(dir, DPERM)
	if err != nil {
		panic(err)
	}
	errLogger, err = NewLogger(filepath.Join(dir, "E_"+time.Now().String()+".log"))
	if err != nil {
		panic(err)
	}
	infoLogger, err = NewLogger(filepath.Join(dir, "I_"+time.Now().String()+".log"))
	if err != nil {
		panic(err)
	}
}

func SetErrorConsoleLogging(enable bool){
	errLogger.console = enable
}

func SetInfoConsoleLoggin(enable bool){
	infoLogger.console = enable
}

//SyncLogger allows for concurrent logging.
type SyncLogger struct {
	logger *log.Logger
	lock   *sync.Mutex
	console bool
}

//Log safely logs data to this logger's log file.
func (this SyncLogger) Log(vals ...interface{}) {
	this.lock.Lock()
	this.logger.Print(vals)
	if this.console{
		fmt.Println(vals)
	}
	this.lock.Unlock()
}

//NewLogger creates a new SyncLogger which writes its logs to the specified file.
func NewLogger(fname string) (*SyncLogger, error) {
	fo, err := os.Create(fname)
	if err != nil {
		return nil, err
	}
	return &SyncLogger{log.New(fo, "", log.LstdFlags), new(sync.Mutex), false}, nil
}

//Log sends data to be logged to the appropriate logger.
func Log(v ...interface{}) {
	if len(v) > 0 {
		if _, ok := v[0].(error); ok {
			errLogger.Log(v)
		} else {
			infoLogger.Log(v)
		}
	}
}
