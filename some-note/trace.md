
![服务图](https://github.com/nevermoressss/studygo/blob/master/pic/trace/1.png)
分布式系统中，一个请求往往需要调用多个服务，调用多次存储系统才能完成。在这一些列的调用中，有并行的有串行的，在这种情况下我们如何才能确定整个请求链路中调用了什么模块，每个模块的耗时是多少，当需要排查问题时如何快速的定位问题？


这就是涉及到链路追踪。

链路追踪是什么？
链路追踪是分布式系统下的一个概念，最早提出这个词是在2010年谷歌发布的一遍论文[Dapper : a Large-Scale Distributed Systems Tracing Infrastructure](https://static.googleusercontent.com/media/research.google.com/zh-CN//archive/papers/dapper-2010-1.pdf)中出现的，它的目的就是为了解决上面的问题，将一次分布式的请求还原成调用链路，将一次分布式请求的调用情况集中展示出来,为排查故障和分析性能提供数据支持。

追踪与跨度
Dapper中提出了“追踪”与“跨度”两个概念。
从客户端发起请求抵达系统的边界开始，记录请求流经的每一个服务，直到到向客户端返回响应为止，这整个过程就称为一次“追踪”（trace）
每次开始调用服务前都要先埋入一个调用记录，这个记录称为一个“跨度”（Span）
每一次 Trace 实际上都是由若干个有顺序、有层级关系的 Span 所组成一颗“追踪树”（Trace Tree）

三个主要构成元素
Span  基本工作单元 ，Span 在不断地启动和停止，同时记录了时间信息，当你创建一个 Span，你必须在未来的某个时刻停止它。

Trace tree  在分布式追踪系统中使用Trace表示对一次请求完整调用链的追踪。可以看出每一次跟踪 Trace 都是一个树型结构，Span 可以体现出服务之间的具体依赖关系。

Annotation  用来及时记录一个事件的存在，一些核心 Annotation 用来定义一个请求的开始和结束

![链路图](https://github.com/nevermoressss/studygo/blob/master/pic/trace/2.png)

数据收集
目前可以分为三种主流的实现方式

基于日志的追踪
基于日志的追踪的思路是将 Trace、Span 等信息直接输出到应用日志中，然后随着所有节点的日志归集过程汇聚到一起，再从全局日志信息中反推出完整的调用链拓扑关系。对应用程序只有很少量的侵入性，对性能影响也非常低。但其缺点是直接依赖于日志归集过程。

基于服务的追踪
是目前最常见的追踪实现方式， 被jaeger、Zipkin、SkyWalking、Pinpoint等主流追踪系统广泛采用，实现思路是通过某些手段给目标应用注入追踪探针，有专门的数据收集协议，把从目标系统中监控得到的服务调用信息，通过另一次独立的 HTTP 或者 RPC 甚至 UDP 请求发送给追踪系统。

基于边车代理的追踪
是服务网格的专属方案，它对应用完全透明，无论是日志还是服务本身都不会有任何变化。只要是通过http或者各种RPC来访问服务就可以被追踪到。Envoy就是边车代理的代表。边车代理本身的对应用透明的工作原理决定了它只能实现服务调用层面的追踪。



jaeger VS zipkin
zipkin

![zipkin](https://github.com/nevermoressss/studygo/blob/master/pic/trace/zipkin.png)

Twitter

zipkin是比较早期的产品，相对来说会更加成熟，核心组件是java编写的相对来说会比较稳定。资源占用大。

Zipkin的架构中包含Reporter，Transport，Colletor，Storage，API，UI几个部分。

其中Reporter集成在每个服务的代码中，负责Span的生成，带内数据(traceid等)的传递，带外数据(span)的上报，采样控制。Transport部分为带外数据上报的通道，zipkin支持http和kafka两种方式。Colletor负责接收带外数据，并插入到集中存储中。Storage为存储组件，适配底层的存储系统，zipkin提供默认的in-memory存储，并支持Mysql，Cassandra，ElasticSearch存储系统。API提供查询、分析和上报链路的接口。接口的定义见zipkin-api。UI用于展示页面展示。

jaeger

![jaeger](https://github.com/nevermoressss/studygo/blob/master/pic/trace/jaeger.png)
Uber

jaeger相对来说比较年轻，采用go语言编写，支持动态采样。

Jaeger Client - 为不同语言实现了符合 OpenTracing 标准的 SDK。应用程序通过 API 写入数据，client library 把 trace 信息按照应用程序指定的采样策略传递给 jaeger-agent。
Agent - 它是一个监听在 UDP 端口上接收 span 数据的网络守护进程，它会将数据批量发送给 collector。它被设计成一个基础组件，部署到所有的宿主机上。Agent 将 client library 和 collector 解耦，为 client library 屏蔽了路由和发现 collector 的细节。
Collector - 接收 jaeger-agent 发送来的数据，然后将数据写入后端存储。Collector 被设计成无状态的组件，因此您可以同时运行任意数量的 jaeger-collector。
Data Store - 后端存储被设计成一个可插拔的组件，支持将数据写入 cassandra、elastic search。
Query - 接收查询请求，然后从后端存储系统中检索 trace 并通过 UI 进行展示。Query 是无状态的，您可以启动多个实例，把它们部署在 nginx 这样的负载均衡器后面。

对比分析
架构上：jaeger将jaeger-agent从业务应用中抽出，部署在宿主机或容器中，专门负责向collector异步上报调用链跟踪数据，这样做将业务应用与collector解耦了，同时也减少了业务应用的第三方依赖。

语言上：jaeger整体是用go语言编写的，在并发性能、对系统资源的消耗上也对基于java的zipkin是有优势的。

采样策略上：zipkin 不支持动态采样（根据请求量来调控采样的频率）

是否支持更换
选择了jaeger之后能否更换zipkin，选择了zipkin之后能否更换jaeger 。 可以的，因为都实现了OpenTracing规范（B站就做过zipkin到jaeger的转换）

OpenTracing 为了解决不同的分布式追踪系统 API 不兼容的问题，诞生了 OpenTracing 规范。
OpenTracing  是一个轻量级的标准化层，它位于应用程序/类库和追踪或日志分析程序之间。

![openTracing](https://github.com/nevermoressss/studygo/blob/master/pic/trace/3.png)

非功能性的挑战
低性能损耗：损耗肯定是有的，分布式追踪不能对服务本身产生明显的性能负担。追踪的主要目的之一就是为了寻找性能缺陷。

对应用透明：应该尽量以非侵入或者少侵入的方式来实现追踪，对开发人员做到透明化。（这块是重点，必要的入侵还是会有的，但是应当尽量减少）

随应用扩缩：现代的分布式服务集群都有根据流量压力自动扩缩的能力，这要求当业务系统扩缩时，追踪系统也能自动跟随，不需要运维人员人工参与。（这块需要运维的支持，服务慢慢接入追踪，前期可以暂时不需要）

存储问题：多集群架构下，不同集群收集到的数据应集中存储。（如果分开存储涉及跨集群调用的时候链路就不完整了）

需要改造的点
核心点：需要对http、grpc、redis操作、mongodb操作、log 中加入span上报相关的代码，以及对span进行一个传递
