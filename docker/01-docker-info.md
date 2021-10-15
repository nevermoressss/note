#### docker技术
```text
隔离性： Linux Namespace(ns)
每个用户实例之间相互隔离, 互不影响。kernel namespace。
其中pid、net、ipc、mnt、uts、user等namespace将container的进程、网络、消息、文件系统、UTS("UNIX Time-sharing System")和用户空间隔离开。

可配额/可度量：Control Groups (cgroups)
cgroups 实现了对资源的配额和度量。 
groups可以限制blkio、cpu、cpuacct、cpuset、devices、freezer、memory、net_cls、ns九大子系统的资源。

便携性：AUFS
AUFS (AnotherUnionFS) 是一种 Union FS, 简单来说就是支持将不同目录挂载到同一个虚拟文件系统下(unite several directories into a single virtual filesystem)的文件系统
```