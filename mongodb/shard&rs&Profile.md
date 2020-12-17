2020-12-17  
问题描述
```text
了解到业务很多场景会利用mongodb的副本集来做数据的读取
那么问题来了
mongodb在分片、副本集的状态下
在副本集中产生的慢查询profile日志是否会同步到PRIMARY？在mongos又是否能查到？
如果是在PRIMARY产生的呢？
```
验证思路
```text
1.测试分片副本集是否工作正常（前提）
2.验证日志级别能否同步（顺带验证）
3.验证慢查询日志是否会同步（主要目的）
```
节点选取
```text
选取一个mongos节点，一个分片的PRIMARY节点，以及PRIMARY对应的一个SECONDARY节点
```
前置操作
```text
通过在mongos:sh.status()查看分片信息
选取一个PRIMARY:rs.status()得到副本集的信息
选取一个SECONDARY:rs.slaveOk()设置副本集可读
得到三个测试节点（mongos，PRIMARY,SECONDARY）
```
副本集是否工作正常
```text
用一个新db和一个新collection
1.在PRIMARY插入数据1，SECONDARY能找到对应的数据，SECONDARY工作正常。
2.去mongos查询数据1，奇怪的知识来了，查不到相应的数据同时发现collection也找不到，但是DB能看到
(猜测：在PRIMARY节点不会触发配置服务器数据分布表的新增和修改)
3.在mongos插入数据2，发现之前在mongos查不到的数据1能查到了，同时在PRIMARY和SECONDARY查到数据2
(猜测：在mongos插入会更新配置服务器)
4.再次在PRIMARY插入数据3，去mongos查找，能查找到新插入的数据3
（猜测应该八九不离十了）
```
验证日志级别是否能同步
```text
先看下3个测试节点分别日志级别状态db.getProfilingLevel()
mongos/PRIMARY/SECONDARY 都是0
（分别修改测试节点的日志隔离级别db.setProfilingLevel(1,200)）
1.修改SECONDARY，SECONDARY日志级别改为1，PRIMARY/mongos不受影响
2.修改PRIMARY，PRIMARY日志级别改为1，SECONDARY/mongos不受影响
3.修改mongos，报错，mongos日志级别只能是0

小结：日志级别不会同步
```
验证慢查询日志是否会同步
```text
验证思路：
1 去到测试db，分别更改3个节点的日志级别db.setProfilingLevel(1,200) 
2 执行一条慢查询db.collection.find({"$where":"sleep(100)||true"}).count()；
3 观察现象
4 恢复db.setProfilingLevel(0) 
验证：
1.修改SECONDARY日志级别执行慢查询，SECONDARY有对应的慢查询日志，PRIMARY没有，通过mongos查询也差不到
2.修改PRIMARY日志级别执行慢查询，PRIMARY有对应的慢查询日志，SECONDARY没有，通过mongos可以查询得到！
3.mongos不能修改日志级别（mongos实际不存储数据）
小结：
查询日志profile是不会同步的
mongos默认情况下读的是PRIMARY节点的数据，所以PRIMARY节点产生的profile在mongos能够看到

如果某些查询都是指定去副本集里读取的话，如果发生慢查询在mongos和PRIMARY是看不到的，要额外注意。
```