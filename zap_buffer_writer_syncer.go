package zlog

import (
	"bufio"
	"bytes"
	"context"
	"io"
	stdlog "log"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

type bufferWriterSyncer struct {
	ws           zapcore.WriteSyncer
	bufferWriter *bufio.Writer
}

const (
	// 缓冲写，默认缓冲大小。超过此大小，会触发写磁盘
	defaultBufferSize = 256 * 1024

	// 定时刷磁盘的时间间隔
	defaultFlushInterval = 30 * time.Second

	// 异步写日志，异步的 buffer 大小，即异步队列中最多缓存几条数据
	defaultAsyncBufferSize = 10000
)

// Buffer wraps a WriteSyncer in a buffer to improve performance,
// if bufferSize = 0, we set it to defaultBufferSize
// if flushInterval = 0, we set it to defaultFlushInterval
func newBufferWriteSyncer(ws zapcore.WriteSyncer, bufferSize int, flushInterval time.Duration) (zapcore.WriteSyncer, io.Closer) {
	if bufferSize == 0 {
		bufferSize = defaultBufferSize
	}

	if flushInterval == 0 {
		flushInterval = defaultFlushInterval
	}

	// bufio is not goroutine safe, so add lock writer here
	ws = &bufferWriterSyncer{
		bufferWriter: bufio.NewWriterSize(ws, bufferSize),
	}

	ctx, cancel := context.WithCancel(context.Background())
	lws := &lockedWriteSyncer{
		ws:        ws,
		bsBufPool: NewBytesBufferPool(1024),
		bsBufChan: make(chan *bytes.Buffer, defaultAsyncBufferSize),
		ctx:       ctx,
		ctxCancel: cancel,
	}

	go lws.consume()

	// flush buffer every interval
	// we do not need exit this goroutine explicitly
	go func() {
		ticker := time.NewTicker(flushInterval)
		for {
			select {
			case <-ticker.C:
				if err := lws.Sync(); err != nil {
					return
				}
			}
		}
	}()

	return lws, lws
}

func (s *bufferWriterSyncer) Write(bs []byte) (int, error) {
	// there are some logic internal for bufio.Writer here:
	// 1. when the buffer is enough, data would not be flushed.
	// 2. when the buffer is not enough, data would be flushed as soon as the buffer fills up.
	// this would lead to log spliting, which is not acceptable for log collector
	// so we need to flush bufferWriter before writing the data into bufferWriter
	if len(bs) > s.bufferWriter.Available() && s.bufferWriter.Buffered() > 0 {
		err := s.bufferWriter.Flush()
		if err != nil {
			return 0, err
		}
	}

	return s.bufferWriter.Write(bs)
}

func (s *bufferWriterSyncer) Sync() error {
	return s.bufferWriter.Flush()
}

type lockedWriteSyncer struct {
	sync.Mutex
	ws        zapcore.WriteSyncer
	bsBufPool *BytesBufferPool
	bsBufChan chan *bytes.Buffer
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (s *lockedWriteSyncer) Close() error {
	s.ctxCancel()
	s.cleanBufChan()
	return s.Sync()
}

func (s *lockedWriteSyncer) cleanBufChan() {
	tm := 100 * time.Millisecond
	timer := time.NewTimer(tm)
	for {
		select {
		case bsBuf, _ := <-s.bsBufChan:
			if bsBuf == nil {
				goto CLEAN_EXIT
			}
			s.doWrite(bsBuf)
		case <-timer.C:
			goto CLEAN_EXIT
		}
		timer.Reset(tm)
	}

CLEAN_EXIT:
}

func (s *lockedWriteSyncer) consume() {
	for {
		select {
		case bsBuf, isOK := <-s.bsBufChan:
			if bsBuf == nil && !isOK {
				stdlog.Println("[zlog] channel close, zlog async consume exit!")
				return
			}
			s.doWrite(bsBuf)
		case <-s.ctx.Done():
			stdlog.Println("[zlog] context done, zlog async consume exit!")
			return
		}
	}
}

func (s *lockedWriteSyncer) Write(bs []byte) (int, error) {
	bsBuf := s.bsBufPool.Get()
	bsBuf.Write(bs)
	select {
	case s.bsBufChan <- bsBuf:
		return len(bs), nil
	default:
		return s.doWrite(bsBuf)
	}
}

func (s *lockedWriteSyncer) doWrite(bsBuf *bytes.Buffer) (int, error) {
	bs := bsBuf.Bytes()
	defer s.bsBufPool.Put(bsBuf)
	if len(bs) == 0 {
		return 0, nil
	}

	s.Lock()
	defer s.Unlock()
	n, err := s.ws.Write(bs)
	return n, err
}

func (s *lockedWriteSyncer) Sync() error {
	s.Lock()
	err := s.ws.Sync()
	s.Unlock()
	return err
}
