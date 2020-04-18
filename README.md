# zlog

   * [zlog](#zlog)
      * [简介](#简介)
         * [日志配置](#日志配置)
            * [配置项说明](#配置项说明)
            * [MaxLogLevel 说明](#maxloglevel-说明)
      * [使用方法](#使用方法)
      * [日志接口说明](#日志接口说明)

## 简介
此日志库基于 [zap](https://github.com/uber-go/zap) 封装，以 `k=v` 的形式打印日志    
支持功能：
* 日志分级
* 支持 k=v tab 分割的日志格式
* ERROR/FATAL 日志单独一份输出到 err log 文件
* 自动轮转
* 指定单个日志文件大小
* 指定轮转的日志个数
* err log 文件实时写，其他日志文件缓冲写

### 日志格式说明
日志格式约定：
- 基本格式： ts / file / logLev / obj 是每条日志必须有的 field
 	 `ts=xxx	file=xxx	logLev=xxx		obj=xxx`
- 在线追踪用的日志信息，描述都用 info=xxx 表示
- 离线收集用的日志信息，数据都用 data=xxx 表示，对应特定方法 LogData / LogReqData

```
ts=04-07T21:19:38.653	file=zlog/zlogger.go:43	logLev=[INFO]		obj=START	info=start done	cost=1586265578652
ts=04-07T21:19:38.654	file=zlog/zlogger.go:47	logLev=[INFO]		obj=LOAD_CONFIG	info=load var xxx
ts=04-07T21:19:38.654	file=zlog/zlogger.go:62	logLev=[INFO]		obj=MYSQL	host=127.0.0.1:3306	info=	cost=1586265578653
```

### 日志配置
#### 配置项说明

具体配置项，见：[log_config.go](./log_config.go#L17)

#### MaxLogLevel 说明

| 数值 | 级别名 |
| ---- | ---- |
| -1 | DEBUG |
| 0 | INFO |
| 1 | WARN |
| 2 | ERROR |
| 3 | FATAL |

## 使用方法

```
go get code.aliyun.com/module-go/zlog
```

代码示例：

```
import (
    "time"

    "code.aliyun.com/module-go/zlog"
)

type YOU_PROJECT_CONFIG struct {
    Log *zlog.LogConfig
}

func main() {
    startTimeNS := time.Now().UnixNano()
    // get config from configfile
    conf := getConfig()

    // init zlog
    zlog.Init(conf.Log)
    defer zlog.Sync()

    // 初始化失败，打 FATAL 并 触发 panic
    zlog.LogPanic(zlog.OBJ_INIT, "init idb manager", err)

    // 记录服务启动耗时
    // 注意：关于耗时的统计，只需传入开始时间的纳秒级时间戳即可 time.Now().UnixNano()
    zlog.LogStart(LL_INFO, "start", startTimeNS)

    // use zlog
    zlog.Log(LL_INFO, "test", "is ok")
    zlog.Log(LL_WARN, "test", "is ok")
    zlog.Log(LL_ERROR, "test", "is ok")
    zlog.Log(LL_FATAL, "test", "is ok")
}
```

## 日志接口说明

具体接口参见：[log.go](./log.go)    
使用时，直接通过 `zlog.LogFuncxxx` 的形式调用即可。    

接口命名、使用约定参见：[log.go](./log.go)  
