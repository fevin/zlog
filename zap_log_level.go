package zlog

import (
	"fmt"
	"go.uber.org/zap/zapcore"
)

var (
	zapLevelMap map[int8]zapcore.Level
)

func init() {
	zapLevelMap = make(map[int8]zapcore.Level, 5)
	zapLevelMap[-1] = zapcore.DebugLevel
	zapLevelMap[0] = zapcore.InfoLevel
	zapLevelMap[1] = zapcore.WarnLevel
	zapLevelMap[2] = zapcore.ErrorLevel
	zapLevelMap[3] = zapcore.FatalLevel
}

func getZapLevel(level int8) zapcore.Level {
	if zapLevel, isOK := zapLevelMap[level]; isOK {
		return zapLevel
	}

	panic(fmt.Sprintf("zlog level is error: the level[%d] doesnot exist!", level))
}
