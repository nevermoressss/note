查看redis实例响应延迟情况
```text
redis-cli -h 127.0.0.1 -p 6379 --intrinsic-latency 60
```
查看一段时间内 Redis 的最小、最大、平均访问延迟
```text
redis-cli -h 127.0.0.1 -p 6379 --latency-history -i 1
// 每间隔 1 秒，采样 Redis 的平均操作耗时
```
慢日志查询(前提是要开启慢日志收集)
```text
SLOWLOG get 5
```
查找各个数据类型最大的bigkey(内部用scan命令实现的)
```text
redis-cli -h 127.0.0.1 -p 6379 --bigkeys -i 0.01
// -i 指定扫描频率 每次扫描后休息多久（秒）
```
一个细节：key主动过期的过程中，如果删除的是一个bigkey，耗时会久，同时这个操作的延迟不会记录在慢日志里

一些优化建议
```text
避免存储 bigkey，降低释放内存的耗时
淘汰策略改为随机淘汰，随机淘汰比 LRU 要快很多（视业务情况调整）
拆分实例，把淘汰 key 的压力分摊到多个实例上
如果使用的是 Redis 4.0 以上版本，开启 layz-free 机制，把淘汰 key 释放内存的操作放到后台线程中执行（配置 lazyfree-lazy-eviction = yes）
```
注意fork时的阻塞
```text
rdb持久化 bgsave
向slave首次同步数据 
AOF日志重写
```
绑核问题
```text
如果你把 Redis 进程只绑定了一个 CPU 逻辑核心上，那么当 Redis 在进行数据持久化时，fork 出的子进程会继承父进程的 CPU 使用偏好。(导致子进程会与主进程发生 CPU 争抢)
Redis 在 6.0 版本已经推出了功能，我们可以通过配置，对主线程、后台线程、后台 RDB 进程、AOF rewrite 进程，绑定固定的 CPU 逻辑核心
```