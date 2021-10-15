进程
```text
计算机里的数据和状态的总和
```
容器技术的核心
```text
通过约束和修改进程的动态表现，为其创造出一个“边界“
Cgroups:制造约束的主要手段
Namespace:修改进程视图的主要方法（隔离）
```
Cgroups
```text
/sys/fs/cgroup/
Linux Cgroups 的全称是 Linux Control Group。它最主要的作用，就是限制一个进程组能够
使用的资源上限，包括 CPU、内存、磁盘、网络带宽等等。此外，Cgroups 还能够对进程进行优先级设置、审计，以及将进程挂起和恢复等操作。
ps:/proc 文件系统并不知道用户通过 Cgroups 给这个容器做了什么样的资源限制
```
docker核心
```text
为待创建的用户进程：
    1.启用Linux Namespace配置
    2.设置指定的Cgroups参数
    3.切换进程的根目录（change Root）
```
```text
同一台机器上的所有容器，都共享宿主机操作系统的内核。  
Docker 在镜像的设计中，引入了层（layer）的概念。  
用户制作镜像的 每一步操作，都会生成一个层，也就是一个增量 rootfs。 
```
联合文件系统UnionFS   
```text

```