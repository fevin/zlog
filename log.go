// 此文件提供外界调用的公开方法
//
// 我们对 log function 做如下约定：
// - Log* 开头的方法，用于打印通用日志
// - LogReq* 开头的方法都会带有 reqId field，用于打印处理请求过程中产生的日志
// - *Err 结尾的方法，用于打印类型为 error 的信息
//
// 日志格式约定：
// - 基本格式： ts / file / logLev / obj 是每条日志必须有的 field
//		 ts=xxx	file=xxx	logLev=xxx		obj=xxx
// - 在线追踪用的日志信息，描述都用 info=xxx 表示
// - 离线收集用的日志信息，数据都用 data=xxx 表示，对应特定方法 LogData / LogReqData
//
// 日志格式示例：
// ts=04-07T21:19:38.653	file=zlog/zlogger.go:43	logLev=[INFO]		obj=START	info=start done	cost=1586265578652
// ts=04-07T21:19:38.654	file=zlog/zlogger.go:47	logLev=[INFO]		obj=LOAD_CONFIG	info=load var xxx
// ts=04-07T21:19:38.654	file=zlog/zlogger.go:57	logLev=[INFO]		obj=MYSQL	info=get version	err=time out
// ts=04-07T21:19:38.654	file=zlog/zlogger.go:62	logLev=[INFO]		obj=MYSQL	host=127.0.0.1:80	info=	cost=1586265578653
//
// 使用此日志库之前，必须先通过 zlog.Init() 方法进行初始化

package zlog

var (
	logger zlogger
)

func Init(conf *LogConfig) error {
	logConf := new(LogConfig)
	logConf.Reset(conf)
	logger = newZapLogger(logConf)
	return nil
}

func Sync() error {
	return logger.Sync()
}

// 记录服务启动耗时
// startTimeNS 单位：纳秒
func LogStart(logLevel, info string, startTimeNS int64) {
	logger.LogStart(logLevel, info, startTimeNS)
}

func Log(logLevel, obj, info string) {
	logger.Log(logLevel, obj, info)
}

// 用于打印离线数据， data=xxx
// 如果 data 是 struct/map 最终会被 json.Marshal 成字符串
func LogData(logLevel, obj string, data interface{}) {
	logger.LogData(logLevel, obj, data)
}

func LogErr(logLevel, obj, info string, err interface{}) {
	logger.LogErr(logLevel, obj, info, err)
}

// （不带reqId）请求了其他组件，比如 mysql、redis、cnd 等
func LogThirdPart(logLevel, obj, host, info string, startTimeNS int64) {
	logger.LogThirdPart(logLevel, obj, host, info, startTimeNS)
}

// FATAL log and panic
// 打印 FATAL 日志并触发 panic
// 此方法可用于记录初始化失败
func LogPanic(obj, info string, err interface{}) {
	logger.LogPanic(obj, info, err)
}

func LogReq(logLevel, obj, reqId, info string) {
	logger.LogReq(logLevel, obj, reqId, info)
}

// 用于打印离线数据， data=xxx
// 如果 data 是 struct/map 最终会被 json.Marshal 成字符串
func LogReqData(logLevel, obj, reqId string, data interface{}) {
	logger.LogReqData(logLevel, obj, reqId, data)
}

func LogReqErr(logLevel, obj, reqId, info string, err interface{}) {
	logger.LogReqErr(logLevel, obj, reqId, info, err)
}

// （带reqId）业务请求中，请求了其他组件，比如 mysql、redis、cnd 等
func LogReqThirdPart(logLevel, obj, reqId, host, info string, startTimeNS int64) {
	logger.LogReqThirdPart(logLevel, obj, reqId, host, info, startTimeNS)
}

// 完整接收到请求数据之后，打印此日志
// startTimeNS 指开始接受请求的时间
func LogReqBegin(logLevel, reqId, reqClientIP, reqUri, reqParams string, startTimeNS int64) {
	logger.LogReqBegin(logLevel, reqId, reqClientIP, reqUri, reqParams, startTimeNS)
}

// 请求处理结束，打印此日志
// startTimeNS 指开始接受请求的时间，同 LogReqBegin 中的 startTimeNS
// retData 返回的数据
func LogReqEnd(logLevel, reqId, retData string, startTimeNS int64) {
	logger.LogReqEnd(logLevel, reqId, retData, startTimeNS)
}
