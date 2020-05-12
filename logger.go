package zlog

type zlogger interface {
	Sync() error
	Close() error

	// Log*
	LogStart(logLevel, info string, startTimeNS int64)
	Log(logLevel, obj, info string)
	LogData(logLevel, obj string, data interface{})
	LogErr(logLevel, obj, info string, err interface{})
	LogThirdPart(logLevel, obj, host, info string, startTimeNS int64)
	LogPanic(obj, info string, err interface{})

	// LogReq*
	LogReq(logLevel, obj, reqId, info string)
	LogReqData(logLevel, obj, reqId string, data interface{})
	LogReqErr(logLevel, obj, reqId, info string, err interface{})
	LogReqThirdPart(logLevel, obj, reqId, host, info string, startTimeNS int64)
	LogReqBegin(logLevel, reqId, reqClientIP, reqUri, reqParams string, startTimeNS int64)
	LogReqEnd(logLevel, reqId, retData string, startTimeNS int64)
}
