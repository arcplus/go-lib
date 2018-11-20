## Env
```
log_level = "info" // 日志级别 debug, info, warn, error
log_async = true // 异步日志开启
log_std_disable = true // 关闭 std 日志

# redis 日志相关，只有两者同不为空时，redis 日志才有效
log_rds_dsn = "xxxx" // redis 地址
log_rds_key = "xxxx" // redis 日志的 key
log_rds_level = "info" // redis 日志的 log level
```