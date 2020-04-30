package zlog

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type _TYPE_ZAP_LOG_fUNC func(msg string, fields ...zap.Field)

var (
	zapEnableErrLogLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
)

func newZapLogger(logConf *LogConfig) zlogger {
	// encoder
	zapEncoderConf := zap.NewProductionEncoderConfig()
	zapEncoderConf.MessageKey = LK_LOG_LEV
	zapEncoderConf.TimeKey = LK_TIMESTAMP
	zapEncoderConf.CallerKey = LK_FILE
	zapEncoderConf.EncodeLevel = zapcore.CapitalLevelEncoder
	zapEncoderConf.EncodeTime = dayMilliTimeEncoder
	zapEncoder := newZapKVTabEncoder(zapEncoderConf)

	// writer
	// normal log write use buffer
	allLogger := &lumberjack.Logger{
		Filename:   logConf.GetLogFilePath(),
		MaxSize:    logConf.MaxLogSizeMB,
		MaxBackups: logConf.MaxLogFileNum,
		LocalTime:  true,
	}
	allLevelWriteSyncer := newBufferWriteSyncer(zapcore.AddSync(allLogger), 0, 20*time.Second)

	errLogFileName := logConf.GetErrorLogFilePath()
	errLevelWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   errLogFileName,
		MaxSize:    logConf.MaxLogSizeMB,
		MaxBackups: logConf.MaxLogFileNum,
		LocalTime:  true,
	})

	// logger
	dLevel := zap.NewAtomicLevelAt(getZapLevel(logConf.MaxLogLevel))
	core := zapcore.NewTee(
		zapcore.NewCore(zapEncoder, allLevelWriteSyncer, dLevel),
		zapcore.NewCore(zapEncoder, errLevelWriteSyncer, zapEnableErrLogLevel),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	zlogger := new(zapLogger)
	zlogger.logger = logger
	zlogger.logFuncMap = make(map[string]_TYPE_ZAP_LOG_fUNC, 5)
	zlogger.logFuncMap[LL_DEBUG] = logger.Debug
	zlogger.logFuncMap[LL_INFO] = logger.Info
	zlogger.logFuncMap[LL_WARN] = logger.Warn
	zlogger.logFuncMap[LL_ERROR] = logger.Error
	zlogger.logFuncMap[LL_FATAL] = logger.Error // zap FATAL will exec os.Exit
	return zlogger
}

type zapLogger struct {
	logger     *zap.Logger
	logFuncMap map[string]_TYPE_ZAP_LOG_fUNC
}

// 强制刷新日志到日志文件中
func (this *zapLogger) Sync() error {
	return this.logger.Sync()
}

func (this *zapLogger) getLogFunc(logLevel string) _TYPE_ZAP_LOG_fUNC {
	logFunc, isOK := this.logFuncMap[logLevel]
	if !isOK {
		this.logger.Fatal("zlog", zap.String(LK_INFO, "logLevel is error:"+logLevel))
		logFunc = this.logger.Info
	}
	return logFunc
}

// 记录服务启动耗时
// startTimeNS 单位：纳秒
func (this *zapLogger) LogStart(logLevel, info string, startTimeNS int64) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, OBJ_START),
		zap.String(LK_INFO, info),
		zap.Int64(LK_COST, getCost(startTimeNS)),
	)
}

func (this *zapLogger) Log(logLevel, obj, info string) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, obj),
		zap.String(LK_INFO, info),
	)
}

// 用于打印离线数据， data=xxx
// 如果 data 是 struct/map 最终会被 json.Marshal 成字符串
func (this *zapLogger) LogData(logLevel, obj string, data interface{}) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, obj),
		zap.Any(LK_DATA, data),
	)
}

func (this *zapLogger) LogErr(logLevel, obj, info string, err interface{}) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, obj),
		zap.String(LK_INFO, info),
		zap.Any(LK_ERR, err),
	)
}

func (this *zapLogger) LogThirdPart(logLevel, obj, host, info string, startTimeNS int64) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, obj),
		zap.String(LK_HOST, host),
		zap.String(LK_INFO, info),
		zap.Int64(LK_COST, getCost(startTimeNS)),
	)
}

func (this *zapLogger) LogPanic(obj, info string, err interface{}) {
	this.getLogFunc(LL_FATAL)(LL_FATAL,
		zap.String(LK_OBJ, obj),
		zap.String(LK_INFO, info),
		zap.Any(LK_ERR, err),
	)
	panic(fmt.Sprintf("info=%s\terr=%v", info, err))
}

func (this *zapLogger) LogReq(logLevel, obj, reqId, info string) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, obj),
		zap.String(LK_REQ_ID, reqId),
		zap.String(LK_INFO, info),
	)
}

// 用于打印离线数据， data=xxx
// 如果 data 是 struct/map 最终会被 json.Marshal 成字符串
func (this *zapLogger) LogReqData(logLevel, obj, reqId string, data interface{}) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, obj),
		zap.String(LK_REQ_ID, reqId),
		zap.Any(LK_DATA, data),
	)
}

func (this *zapLogger) LogReqErr(logLevel, obj, reqId, info string, err interface{}) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, obj),
		zap.String(LK_REQ_ID, reqId),
		zap.String(LK_INFO, info),
		zap.Any(LK_ERR, err),
	)
}

func (this *zapLogger) LogReqThirdPart(logLevel, obj, reqId, host, info string, startTimeNS int64) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, obj),
		zap.String(LK_REQ_ID, reqId),
		zap.String(LK_HOST, host),
		zap.String(LK_INFO, info),
		zap.Int64(LK_COST, getCost(startTimeNS)),
	)
}

func (this *zapLogger) LogReqBegin(logLevel, reqId, reqClientIP, reqUri, reqParams string, startTimeNS int64) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, OBJ_RB),
		zap.String(LK_REQ_ID, reqId),
		zap.String(LK_REQ_CLIENTIP, reqClientIP),
		zap.String(LK_REQ_URI, reqUri),
		zap.String(LK_REQ_PARAMS, reqParams),
		zap.Int64(LK_COST, getCost(startTimeNS)),
	)
}

// 请求处理结束，打印此日志
// startTimeNS 指开始接受请求的时间，同 RequestBegin 中的 startTimeNS
// retData 返回的数据
func (this *zapLogger) LogReqEnd(logLevel, reqId, retData string, startTimeNS int64) {
	this.getLogFunc(logLevel)(logLevel,
		zap.String(LK_OBJ, OBJ_RE),
		zap.String(LK_REQ_ID, reqId),
		zap.String(LK_RET_DATA, retData),
		zap.Int64(LK_COST, getCost(startTimeNS)),
	)
}
