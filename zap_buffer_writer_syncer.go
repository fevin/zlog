package zlog

import (
	"bufio"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

type bufferWriterSyncer struct {
	ws           zapcore.WriteSyncer
	bufferWriter *bufio.Writer
}

// defaultBufferSize sizes the buffer associated with each WriterSync.
const defaultBufferSize = 256 * 1024

// defaultFlushInterval means the default flush interval
const defaultFlushInterval = 30 * time.Second

// Buffer wraps a WriteSyncer in a buffer to improve performance,
// if bufferSize = 0, we set it to defaultBufferSize
// if flushInterval = 0, we set it to defaultFlushInterval
func newBufferWriteSyncer(ws zapcore.WriteSyncer, bufferSize int, flushInterval time.Duration) zapcore.WriteSyncer {
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
	ws = &lockedWriteSyncer{ws: ws}

	// flush buffer every interval
	// we do not need exit this goroutine explicitly
	go func() {
		select {
		case <-time.NewTicker(flushInterval).C:
			if err := ws.Sync(); err != nil {
				return
			}
		}
	}()

	return ws
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
	ws zapcore.WriteSyncer
}

func (s *lockedWriteSyncer) Write(bs []byte) (int, error) {
	s.Lock()
	n, err := s.ws.Write(bs)
	s.Unlock()
	return n, err
}

func (s *lockedWriteSyncer) Sync() error {
	s.Lock()
	err := s.ws.Sync()
	s.Unlock()
	return err
}
