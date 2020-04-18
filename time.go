package zlog

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// startTimeNS 单位 纳秒
func getCost(startTimeNS int64) int64 {
	var cost int64 = 0
	if startTimeNS > 0 {
		cost = time.Now().UnixNano()/1e6 - startTimeNS/1e6
	}
	return cost
}

func encodeTimeLayout(t time.Time, layout string, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

func dayMilliTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {

	encodeTimeLayout(t, "01-02T15:04:05.000", enc)
}
