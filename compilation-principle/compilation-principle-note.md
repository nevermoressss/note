为什么要学习编译原理？
```text
与工作息息相关
理解了底层机制，才能有更深入思考问题，以及深层次解决问题的能力，而不是只能盲目地搜索答案，从表面解决问题。而学习编译原理能让你从前端的语法维度、代码优化的维度、与硬件结合的维度几个方面，加深对计算机技术的理解，提升自己的竞争力。
```
编译器技术划分
```text
词法分析->语法分析->语义分析->生成中间代码->优化->生成目标代码
编译器前端：词法分析->语法分析->语义分析
编译器后端：生成中间代码->优化->生成目标代码
```
词法分析
```text
“正则文法”和“有限自动机”
分析整个程序的字符串，当遇到不同的字符时，会驱使它迁移到不同的状态。
词法分析是识别一个个的单词
```
语法分析
```text
语法分析就是在词法分析的基础上识别出程序的语法结构。这个结构是一个树状结构，是计算机容易理解和执行的。
一个程序就是一棵树，这棵树叫做抽象语法树（Abstract Syntax Tree，AST）。树的每个节点（子树）是一个语法单元，这个单元的构成规则就叫“语法”。每个节点还可以有下级节点。
```
语义分析
```text
语义分析就是要让计算机理解我们的真实意图，把一些模棱两可的地方消除掉。
比如：
某个表达式的计算结果是什么数据类型？如果有数据类型不匹配的情况，是否要做自动转换？

如果在一个代码块的内部和外部有相同名称的变量，我在执行的时候到底用哪个？

在同一个作用域内，不允许有两个名称相同的变量，这是唯一性检查。你不能刚声明一个变量 a，紧接着又声明同样名称的一个变量 a，这就不允许了。
```
上下文无关文法
```text
正则文法是上下文无关文法的一个子集。它们的区别呢，就是上下文无关文法允许递归调用，而正则文法不允许。
上下文无关的意思是，无论在任何情况下，文法的推导规则都是一样的。比如，在变量声明语句中可能要用到一个算术表达式来做变量初始化，而在其他地方可能也会用到算术表达式。不管在什么地方，算术表达式的语法都一样，都允许用加法和乘法，计算优先级也不变。
```
类型是什么
```text
类型是针对一组数值，以及在这组数值之上的一组操作。
类型是高级语言赋予的一种语义，有了类型这种机制，就相当于定了规矩，可以检查施加在数据上的操作是否合法。因此类型系统最大的好处，就是可以通过类型检查降低计算出错的概率。

```
语义分析本质
```text
语义分析的本质，就是针对上下文相关的情况做处理。
```