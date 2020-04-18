package zlog

const (
	// log level
	LL_DEBUG = "[DEBUG]"
	LL_INFO  = "[INFO]"
	LL_WARN  = "[WARN]"
	LL_ERROR = "[ERROR]"
	LL_FATAL = "[FATAL]"

	// obj
	OBJ_INIT        = "INIT"
	OBJ_START       = "START"
	OBJ_LOAD_CONFIG = "LOAD_CONFIG" // 加载配置文件
	OBJ_RB          = "RB"          // 请求开始
	OBJ_REQ         = "REQ"         // 请求处理过程中
	OBJ_RE          = "RE"          // 请求结束

	// log key
	LK_TIMESTAMP    = "ts"
	LK_FILE         = "file"
	LK_LOG_LEV      = "logLev"
	LK_OBJ          = "obj"
	LK_HOST         = "host"
	LK_INFO         = "info"
	LK_DATA         = "data" // 离线数据标识
	LK_ERR          = "err"
	LK_COST         = "cost"
	LK_REQ_ID       = "reqId"
	LK_REQ_CLIENTIP = "reqClientIP"
	LK_REQ_HOST     = "reqHost"
	LK_REQ_URI      = "reqUri"
	LK_REQ_PARAMS   = "reqParams"
	LK_RET_DATA     = "retData"
	LK_RET_CODE     = "retCode"
)
