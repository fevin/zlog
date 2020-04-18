package zlog

import (
	"flag"
	"os"
	"path/filepath"
)

var (
	defaultMaxLogLevel   int8 = 0
	defaultMaxLogSizeMB  int  = 1024
	defaultMaxLogFileNum int  = 10
	defaultLogDirName         = flag.String("log_dir", "", "default log file dir")
	defaultLogFileName        = filepath.Base(os.Args[0]) + ".log"
)

type LogConfig struct {
	MaxLogLevel      int8   `json:"MaxLogLevel"`      // 输出的日志级别
	MaxLogSizeMB     int    `json:"MaxLogSizeMB"`     // 单个日志文件大小
	MaxLogFileNum    int    `json:"MaxLogFileNum"`    // 保留日志文件个数
	LogDirName       string `json:"LogDirName"`       // 日志输出目录
	LogFileName      string `json:"LogFileName"`      // 日志文件名（内容包含各个 level 的日志）
	ErrorLogFileName string `json:"ErrorLogFileName"` // 错误日志文件名（内容那个包含 ERROR/FATAL 日志）
}

func (this *LogConfig) Reset(conf *LogConfig) {
	this.MaxLogLevel = defaultMaxLogLevel
	if conf.MaxLogLevel != 0 {
		this.MaxLogLevel = conf.MaxLogLevel
	}

	this.MaxLogSizeMB = defaultMaxLogSizeMB
	if conf.MaxLogSizeMB != 0 {
		this.MaxLogSizeMB = conf.MaxLogSizeMB
	}

	this.MaxLogFileNum = defaultMaxLogFileNum
	if conf.MaxLogFileNum != 0 {
		this.MaxLogFileNum = conf.MaxLogFileNum
	}

	this.LogDirName = *defaultLogDirName
	if conf.LogDirName != "" {
		this.LogDirName = conf.LogDirName
	}

	this.LogFileName = defaultLogFileName
	if conf.LogFileName != "" {
		this.LogFileName = conf.LogFileName
	}

	this.ErrorLogFileName = "error-" + this.LogFileName
	if conf.ErrorLogFileName != "" {
		this.ErrorLogFileName = conf.ErrorLogFileName
	}
}

func (this *LogConfig) GetLogFilePath() string {
	return this.LogDirName + string(filepath.Separator) + this.LogFileName
}

func (this *LogConfig) GetErrorLogFilePath() string {
	return this.LogDirName + string(filepath.Separator) + this.ErrorLogFileName
}
