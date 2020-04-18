package zlog

// go test -bench .

import (
	"errors"
	"testing"
)

func init() {
	mockConf := &LogConfig{
		MaxLogLevel:   0,
		MaxLogSizeMB:  10,
		MaxLogFileNum: 3,
		LogDirName:    "./logtemp",
	}
	Init(mockConf)
}

func TestLog(t *testing.T) {
	defer Sync()
	LogStart(LL_INFO, "start done", 1)
	Log(LL_INFO, OBJ_LOAD_CONFIG, "load var xxx")
	type Data struct {
		Key string `json:"key"`
		Val string `json:"value"`
	}
	dt := Data{
		Key: "test_key",
		Val: "test_val",
	}
	testObj := "TEST_OBJ"
	LogData(LL_INFO, testObj, dt)
	err := errors.New("time out")
	LogErr(LL_FATAL, testObj, "get version", err)
	LogThirdPart(LL_INFO, testObj, "127.0.0.1:80", "", 1)
}

func BenchmarkInfoLog(b *testing.B) {
	defer Sync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Log(LL_INFO, "test", "is ok")
	}
}

func BenchmarkErrorLog(b *testing.B) {
	defer Sync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Log(LL_ERROR, "test", "is ok")
	}
}

func BenchmarkInfoLogParallel(b *testing.B) {
	defer Sync()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Log(LL_INFO, "test", "is ok")
		}
	})
}

func BenchmarkErrorLogParallel(b *testing.B) {
	defer Sync()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Log(LL_ERROR, "test", "is ok")
		}
	})
}
