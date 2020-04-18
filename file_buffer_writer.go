package zlog

import (
	"bufio"
	"fmt"
	"io"
	stdlog "log"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const bufferSize = 256 * 1024
const flushInterval = 10 * time.Second

func newBufFileWriteSyncer(w io.Writer, fileMaxSizeMB int) *bufFileWriteSyncer {
	syncer := new(bufFileWriteSyncer)
	syncer.Writer = bufio.NewWriterSize(w, bufferSize)
	syncer.fileMaxSize = uint64(fileMaxSizeMB * 1024 * 1024)
	go syncer.flushDaemon()
	return syncer
}

type bufFileWriteSyncer struct {
	*bufio.Writer
	logger      *lumberjack.Logger
	nbytes      uint64 // 已经写到这个文件的字节数
	fileMaxSize uint64 // 单文件最大字节数
	mu          sync.Mutex
}

func (this *bufFileWriteSyncer) Sync() error {
	return this.lockAndFlush()
}

func (this *bufFileWriteSyncer) Write(p []byte) (n int, err error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.nbytes+uint64(len(p)) >= this.fileMaxSize {
		if err := this.flushToFile(); err != nil {
			panic(fmt.Sprintf("zlog buf flush to file error:%v", err))
		}
	}
	n, _ = this.Writer.Write(p)
	this.nbytes += uint64(n)
	return
}

func (this *bufFileWriteSyncer) flushToFile() error {
	this.nbytes = 0
	return this.Writer.Flush()
}

func (this *bufFileWriteSyncer) flushDaemon() {
	ticker := time.NewTicker(flushInterval)
	for {
		select {
		case <-ticker.C:
			if err := this.lockAndFlush(); err != nil {
				stdlog.Println("[FATAL] zlog flushDaemon flush fail, error:", err)
			}
		}
	}
}

func (this *bufFileWriteSyncer) lockAndFlush() error {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.flushToFile()
}
