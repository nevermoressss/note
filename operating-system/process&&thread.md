什么是进程
```text
用户角度：进程就是一段程序的执行过程
系统角度：系统分配资源的最小单位
```
进程的结构
```text
代码段（程序）
数据段（数据）
堆栈段（进程控制块PCB）
```
进程控制块PCB（进程描述信息、进程控制、进程管理信息，资源分配清单）
```text
标识符：进程PID
状态：任务状态，退出代码，退出信号等
优先级：进程的优先级
程序计数器：即将被执行的下一条指令的地址
内存指针：程序代码和进程相关数据的指针，其他进程共享的内存块指针
上下文数据：进程执行时寄存器中的数据
IO状态：分配给进程的IO设备和进程使用的文件列表
```
进程三态
```text
运行态 占用CPU
就绪态 可运行，但是因为其他进程正在运行而暂停CPU
阻塞态 等待某个事件停止运行
```
进程上下文切换
```text
CPU上下文：指 C P U 寄存器 和 程序计数器
CPU上下文切换：把前一个任务的 C P U上下文  保存起来，然后在加载当前任务的 C P U上下文，最后再跳转到 程序计数器 所指的新位置，运行任务。
进程的上下文切换：发生在内核态，用户空间资源（虚拟内存、栈、全局变量等）与内核空间资源（内核堆栈、寄存器等）。
```
什么是线程
````text
线程是进程中的一个实体
系统调度和分派的基本单位
拥有独立一套的寄存器和栈
线程并不拥有资源
````
线程的上下文切换
```text
不同进程的线程切换：切换过程和进程上线文切换一样
同进程内线程的切换：内存是共享，所以切换的时候只需要切换私有数据，寄存器和栈
```
线程模型
```text
内核线程：
    内核管理 TCB在内核空间
    优点：
        1.由内核空间管理，创建销毁调度等全自动管理
        2.能够利用cpu的多核特性，实现并行执行
        3.阻塞不会影响其他内核线程
    缺点：
        1.大部分操作都涉及内核态，需要用户切换到内核态执行
        2.内核资源有限，无法大量创建内核线程
用户线程：
    在用户控件 操作系统看不到TCB
    优点：
        1.不走内核态，速度快
        2.不消耗内核资源
    缺点：
        1.创建销毁调度等需要自己管理
        2.用户线程阻塞会导致整个进程内的其他用户线程阻塞（整个进程阻塞），因为内核感知不到用户线程，所以无法去调度其他用户线程
        3.无法利用cpu多核特性，还是因为内核感知不到用户线程
一对一模型：
    进程 只需要创建使用L W P（轻量级进程） ，因为一个 L W P 由一个 内核线 程支持，所以最终是内核管理线程，可以调度到其他处理器上
一对多模型：
    多个用 户级线程 对用到同一个 L W P 上实现，因为是用户态通过用户空间的线程库对线程管理，所以速度特别快，不会涉及到用户态与内核态的转换
多对多模型（m:n）：
    多用户线程 可以绑定不同的内核线程 ，这点与 一对一模型 类似，其次又区别于一对一模型，进程内的 多用户线程 与 内核线程 不是一对一绑定，而是动态绑定，当某个 内核线程 因绑定的 用户线程 执行阻塞操作，让出 C P U 时，绑定该 内核线程 的其他 用户线程 可以解绑，重新绑定到其他 内核线程 继续运行。
```
调度算法
```text
先来先服务算法（First Come First Severd, FCFS）：
    谁先来，谁先被 C P U 执行，后到的就乖乖排队等待，十分简单的算法，C P U每次调度 就绪队列 的第一个进程，直到进程退出或阻塞，才会把该进程入队到队尾
    F C F S对长作业有利，适用于 C P U 繁忙型作业的系统，而不适用于 I/O 繁忙型作业的系统。
最短作业优先算法（Shortest Job First, SJF）
    它会优先选择运行时间最短的进程，有助于提高系统吞吐量。但是对长作业不利，所以很容易造成一种极端现象。比如，一个 长作业 在就绪队列等待运行，而这个就绪队列有非常多的短作业，最终使得 长作业 不断的往后推，周转时间变长，致使长作业长期不会被运行（适用于 I/O 繁忙型作业的系统）
高响应比优先算法 （Highest Response Ratio Next, HRRN）
    优先权=（等待时间+要求服务时间）/ 要求服务时间
    如果两个进程的「等待时间」相同时，「要求的服务时间」越短，「优先权」就越高，这样短作业的进程容易被选中运行
    如果两个进程「要求的服务时间」相同时，「等待时间」越长，「优先权」就越高，这就兼顾到了长作业进程，因为进程的响应比可以随时间等待的增加而提高，当其等待时间足够长时，其响应比便可以升到很高，从而获得运行的机会
时间片轮转算法（Round Robin, RR）
    给每个进程分配相同时间片（Quantum），允许进程在该时间段中运行。
最高优先级（Highest Priority First，HPF）算法
    静态优先级：创建进程时候，已经确定优先级，整个运行时间优先级都不会变化
    动态优先级：根据进程的动态变化调整优先级，比如进程运行时间增加，则降低其优先级，如果进程等待时间（就绪队列的等待时间）增加，则提高优先级。
多级反馈队列（Multilevel Feedback Queue）算法
    「多级」表示有多个队列，每个队列优先级从高到低，优先级越高的队列拥有的时间片越短
    「反馈」 表示有新的进程进入优先级高的队列时，停止当前运行进程，去运行优先级高的队列
```