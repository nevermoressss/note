docker 架构
![docker-architecture](https://github.com/nevermoressss/studygo/blob/master/pic/docker/docker-architecture.jpg)

用户是使用Docker Client与Docker Daemon建立通信，并发送请求给后者。  
而Docker Daemon作为Docker架构中的主体部分，首先提供Server的功能使其可以接受Docker Client的请求；而后Engine执行Docker内部的一系列工作，每一项工作都是以一个Job的形式的存在。
Job的运行过程中，当需要容器镜像时，则从Docker Registry中下载镜像，并通过镜像管理驱动graphdriver将下载镜像以Graph的形式存储；当需要为Docker创建网络环境时，通过网络管理驱动networkdriver创建并配置Docker容器网络环境；当需要限制Docker容器运行资源或执行用户指令等操作时，则通过execdriver来完成。
而libcontainer是一项独立的容器管理包，networkdriver以及execdriver都是通过libcontainer来实现具体对容器进行的操作。

```text
docker Client
Docker Client可以通过以下三种方式和Docker Daemon建立通信：tcp://host:port，unix://path_to_socket和fd://socketfd。为了简单起见，本文一律使用第一种方式作为讲述两者通信的原型。与此同时，与Docker Daemon建立连接并传输请求的时候，Docker Client可以通过设置命令行flag参数的形式设置安全传输层协议(TLS)的有关参数，保证传输的安全性。

Docker Client发送容器管理请求后，由Docker Daemon接受并处理请求，当Docker Client接收到返回的请求相应并简单处理后，Docker Client一次完整的生命周期就结束了。当需要继续发送容器管理请求时，用户必须再次通过docker可执行文件创建Docker Client。
```
```text
Docker Daemon
Docker Daemon的架构，大致可以分为三部分：Docker Server、Engine和Job。

Docker Server:
接受并调度分发Docker Client发送的请求。
Docker Server的运行在Docker的启动过程中，是靠一个名为"serveapi"的job的运行来完成的。原则上，Docker Server的运行是众多job中的一个，但是为了强调Docker Server的重要性以及为后续job服务的重要特性，将该"serveapi"的job单独抽离出来分析，理解为Docker Server。

Engine:
Engine是Docker架构中的运行引擎，同时也Docker运行的核心模块。它扮演Docker container存储仓库的角色，并且通过执行job的方式来操纵管理这些容器。
在Engine数据结构的设计与实现过程中，有一个handler对象。该handler对象存储的都是关于众多特定job的handler处理访问。

Job:
一个Job可以认为是Docker架构中Engine内部最基本的工作执行单元。
Docker可以做的每一项工作，都可以抽象为一个job。
```

```text
Docker Registry
Docker Registry是一个存储容器镜像的仓库。而容器镜像是在容器被创建时，被加载用来初始化容器的文件架构与目录。
在Docker的运行过程中，Docker Daemon会与Docker Registry通信，并实现搜索镜像、下载镜像、上传镜像三个功能，这三个功能对应的job名称分别为"search"，"pull" 与 "push"。
```
```text
Graph
Graph在Docker架构中扮演已下载容器镜像的保管者，以及已下载容器镜像之间关系的记录者。一方面，Graph存储着本地具有版本信息的文件系统镜像，另一方面也通过GraphDB记录着所有文件系统镜像彼此之间的关系。
```

```text
Driver
Driver是Docker架构中的驱动模块。通过Driver驱动，Docker可以实现对Docker容器执行环境的定制。由于Docker运行的生命周期中，并非用户所有的操作都是针对Docker容器的管理，另外还有关于Docker运行信息的获取，Graph的存储与记录等。因此，为了将Docker容器的管理从Docker Daemon内部业务逻辑中区分开来，设计了Driver层驱动来接管所有这部分请求。

在Docker Driver的实现中，可以分为以下三类驱动：graphdriver、networkdriver和execdriver。
```
docker run 执行流程
![docker-run](https://github.com/nevermoressss/studygo/blob/master/pic/docker/docker-run.png)

docker 源码分布
```text
api：定义API，使用了Swagger2.0这个工具来生成API，配置文件在api/swagger.yaml
builder：用来build docker镜像的包，看来历史比较悠久了
bundles：这个包是在进行docker源码编译和开发环境搭建的时候用到的，编译生成的二进制文件都在这里。
cli：使用cobra工具生成的docker客户端命令行解析器。
client：接收cli的请求，调用RESTful API中的接口，向server端发送http请求。
cmd：其中包括docker和dockerd两个包，他们分别包含了客户端和服务端的main函数入口。
container：容器的配置管理，对不同的platform适配。
contrib：这个目录包括一些有用的脚本、镜像和其他非docker core中的部分。
daemon：这个包中将docker deamon运行时状态expose出来。
distribution：负责docker镜像的pull、push和镜像仓库的维护。
dockerversion：编译的时候自动生成的。
docs：文档。这个目录已经不再维护，文档在另一个仓库里。
experimental：从docker1.13.0版本起开始增加了实验特性。
hack：创建docker开发环境和编译打包时用到的脚本和配置文件。
image：用于构建docker镜像的。
integration-cli：集成测试
layer：管理 union file system driver上的read-only和read-write mounts。
libcontainerd：访问内核中的容器系统调用。
man：生成man pages。
migrate：将老版本的graph目录转换成新的metadata。
oci：Open Container Interface库
opts：命令行的选项库。
pkg：
plugin：docker插件后端实现包。
profiles：里面有apparmor和seccomp两个目录。用于内核访问控制。
project：项目管理的一些说明文档。
reference：处理docker store中镜像的reference。
registry：docker registry的实现。
restartmanager：处理重启后的动作。
runconfig：配置格式解码和校验。
vendor：各种依赖包。
volume：docker volume的实现。
```