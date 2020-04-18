# zlog

此日志库基于 [zap](https://github.com/uber-go/zap) 封装，以 `k=v` 的形式打印日志    
支持功能：
* 日志分级
* 支持 k=v tab 分割的日志格式
* ERROR/FATAL 日志单独一份输出到 err log 文件
* 自动轮转，基于 [lumberjack](https://github.com/natefinch/lumberjack)
* 指定单个日志文件大小
* 指定轮转的日志个数
* err log 文件实时写，其他日志文件缓冲写
* 支持写日志是使用 buffer writer，批量写（性能是不使用 buffer 的 6~7 倍）

----

  * [日志格式](#日志格式)
     * [日志文件名](#日志文件名)
     * [日志格式约定](#日志格式约定)
  * [日志配置](#日志配置)
     * [配置项说明](#配置项说明)
     * [MaxLogLevel 说明](#maxloglevel-说明)
  * [使用方法](#使用方法)
  * [日志接口说明](#日志接口说明)

## 日志格式
### 日志文件名
会同时写两种日志文件：
* program.log：所有级别日志都会往此文件写，并且使用 buffer writer 提升写入性能
* error-program.log：ERROR / FATAL 级别日志会额外往此文件写入一份，写入时不使用 buffer writer

示例：
```
➜  zlog git:(master) ✗ go test -v .

// zlog.test 为当前进程名
✗ ll logtemp
total 24
-rw-r--r--  1 javin  staff    25B  4 18 19:16 README.md
-rw-r--r--  1 javin  staff   104B  4 18 20:51 error-zlog.test.log
-rw-r--r--  1 javin  staff   535B  4 18 20:51 zlog.test.log

➜  zlog git:(master) ✗ cat logtemp/error-zlog.test.log
ts=04-18T20:51:43.656	file=zlog/log_test.go:35	logLev=FATAL		obj=TEST_OBJ	info=get version	err=time out

➜  zlog git:(master) ✗ cat logtemp/zlog.test.log
ts=04-18T20:51:43.656	file=zlog/log_test.go:22	logLev=INFO		obj=START	info=start done	cost=1587214303656
ts=04-18T20:51:43.656	file=zlog/log_test.go:23	logLev=INFO		obj=LOAD_CONFIG	info=load var xxx
ts=04-18T20:51:43.656	file=zlog/log_test.go:33	logLev=INFO		obj=TEST_OBJ	data={"key":"test_key","value":"test_val"}
ts=04-18T20:51:43.656	file=zlog/log_test.go:35	logLev=FATAL		obj=TEST_OBJ	info=get version	err=time out
ts=04-18T20:51:43.657	file=zlog/log_test.go:36	logLev=INFO		obj=TEST_OBJ	host=127.0.0.1:80	info=	cost=1587214303657
```

### 日志格式约定
基本格式： ts / file / logLev / obj 是每条日志必须有的 field  
```
ts=xxx	file=xxx	logLev=xxx		obj=xxx
```

## 日志配置
### 配置项说明

具体配置项，见：[log_config.go](./log_config.go#L17)

### MaxLogLevel 说明

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
