
## 前言

所谓垃圾就是不再需要的内存块，这些垃圾如果不清理就没办法再次被分配使用，在不支持垃圾回收的编程语言里，这些垃圾内存就是泄露的内存。

Golang的垃圾回收（GC）也是内存管理的一部分，了解垃圾回收最好先了解前面介绍的内存分配原理。

## 垃圾回收算法

- 引用计数：对每个对象维护一个引用计数，当引用该对象的对象被销毁时，引用计数减1，当引用计数器为0是回收该对象。
	- 优点：对象可以很快的被回收，不会出现内存耗尽或达到某个阀值时才回收。
	- 缺点：不能很好的处理循环引用，而且实时维护引用计数，有也一定的代价。
	- 代表语言：Python、PHP、Swift
- 标记-清除：从根变量开始遍历所有引用的对象，引用的对象标记为"被引用"，没有被标记的进行回收。
	- 优点：解决了引用计数的缺点。
	- 缺点：需要STW，即要暂时停掉程序运行。
	- 代表语言：Golang(其采用三色标记法混合写屏障)
- 分代回收：按照对象生命周期长短划分不同的代空间，生命周期长的放入老年代，而短的放入新生代，不同代有不能的回收算法和回收频率。
	- 优点：回收性能好
	- 缺点：算法复杂
	- 代表语言： JAVA
## Golang垃圾回收

### 垃圾回收原理

简单来说，垃圾回收的核心就是标记出哪些内存还在使用中（即被引用到），哪些内存不再使用了（即未被引用到）把未被引用到的内存回收掉，以供后续内存分配时使用

下图展示了一段内存，内存心中既有以分配掉的内存，也有为分配的内存，垃圾回收的目标就是把哪些已经分配但是没有对象引用的内存找出来并回收掉

![upload successful](http://blogs.xinghe.host/images/pasted-102.png)

上图中，内存块1、2、4号位上的内存块已被分配(数字1代表已被分配，0 未分配)。变量a, b为一指针，指向内存的1、2号位。内存块的4号位曾经被使用过，但现在没有任何对象引用了，就需要被回收掉。

垃圾回收开始时从root对象开始扫描，把root对象引用的内存标记为"被引用"，考虑到内存块中存放的可能是指针，所以还需要递归的进行标记，全部标记完成后，只保留被标记的内存，未被标记的全部标识为未分配即完成了回收。

### 内存标记(Mark)
前面介绍内存分配时，介绍过span数据结构，span中维护了一个个内存块，并由一个位图allocBits表示每个内存块的分配情况。在span数据结构中还有另一个位图gcmarkBits用于标记内存块被引用情况。

![upload successful](http://blogs.xinghe.host/images/pasted-103.png)

如上图所示，allocBits记录了每块内存分配情况，而gcmarkBits记录了每块内存标记情况。标记阶段对每块内存进行标记，有对象引用的的内存标记为1(如图中灰色所示)，没有引用到的保持默认为0.

allocBits和gcmarkBits数据结构是完全一样的，标记结束就是内存回收，回收时将allocBits指向gcmarkBits，则代表标记过的才是存活的，gcmarkBits则会在下次标记时重新分配内存，非常的巧妙。

### 三色标记法

前面介绍了对象标记状态的存储方式，还需要一个队列来存放标记的对象，可以简单想象成把对象从标记队列中取出，将对象的引用状态标记在span的gcmarkBits位图，把对象引用到的其他对象再放入到白哦即队列中等待被标记

三色只是抽象出来的颜色

- 灰色：对象还在标记队列中等待
- 黑色：对象已被标记，gcmarkBits对应的位为1（该对象不会在本次GC中被清理）
- 白色：对象未被标记，gcmarkBits对应的位为0（该对象将会在本次GC中被清理）

![upload successful](http://blogs.xinghe.host/images/pasted-104.png)

初始状态下都是白色

接着开始扫描根对象a、b

![upload successful](http://blogs.xinghe.host/images/pasted-105.png)

由于根对象引用到了AB，那么AB两个变色灰色，接下来开始分析灰色对象，分析A时，A没有引用其他对象迅速转成黑色，B引用了D，B转入黑色的同时把D放在标记队列中，也就是把D标记成灰色，进行接下来的分析：

![upload successful](http://blogs.xinghe.host/images/pasted-106.png)

上图中灰色对象只有D，由于D没有引用其他对象，所以D转入黑色，标记过程结束

![upload successful](http://blogs.xinghe.host/images/pasted-107.png)

最终黑色的对象会被保留下来，八色对象会被回收掉

### Stop The World

印度电影《苏丹》中有句描述摔跤的一句台词是：“所谓摔跤，就是把对手控制住，然后摔倒他。”

对于垃圾回收来说，回收过程中也需要控制住内存的变化，否则回收过程中指针传递会引起内存引用关系变化，如果错误的回收了还在使用的内存，结果将是灾难性的。

Golang中的STW（Stop The World）就是停掉所有的goroutine，专心做垃圾回收，待垃圾回收结束后再恢复goroutine。

STW时间的长短直接影响了应用的执行，时间过长对于一些web应用来说是不可接受的，这也是广受诟病的原因之一。

为了缩短STW的时间，Golang不断优化垃圾回收算法，这种情况得到了很大的改善。

## 垃圾回收优化

### 写屏障(Write Barrier)

前面说过STW目的是防止GC扫描时内存变化而停掉goroutine，而写屏障就是让goroutine与GC同时运行的手段。虽然写屏障不能完全消除STW，但是可以大大减少STW的时间。

写屏障类似一种开关，在GC的特定时机开启，开启后指针传递时会把指针标记，即本轮不回收，下次GC时再确定。

GC过程中新分配的内存会被立即标记，用的并不是写屏障技术，也即GC过程中分配的内存不会在本轮GC中回收。

### 辅助GC(Mutator Assist)

为了防止内存分配过快，在GC执行过程中，如果goroutine需要分配内存，那么这个goroutine会参与一部分GC的工作，即帮助GC做一部分工作，这个机制叫作Mutator Assist。

## 垃圾回收触发时机

### 内存分配量达到阀值触发GC

每次内存分配时都会检查当前内存分配量是否已达到阀值，如果达到阀值则立即启动GC。

`阀值 = 上次GC内存分配量 * 内存增长率`

内存增长率由环境变量GOGC控制，默认为100，即每当内存扩大一倍时启动GC。

### 手动触发

程序代码中也可以使用runtime.GC()来手动触发GC，这主要用于Gc性能测试和统计

### GC性能优化

GC性能与对象数量负相关，对象越多GC性能越差，对程序影响也越大

所以GC性能优化的思路之一就是减少对象分配个数，比如对象复用或使用大对象组合多个小对象等待

另外，由于内存逃逸现象，有些隐式的内存分配也会产生，也有可能成为GC的负担

关于GC性能优化的具体方法，后面单独结束

## 逃逸分析

所谓逃逸分析（Escape analysis）是指由编译器决定内存分配的位置，不需要程序员指定。 函数中申请一个新的对象

- 如果分配在栈中，则函数执行结束可自动将内存回收；
- 如果分配在堆中，则函数执行结束可交给GC（垃圾回收）处理;

有了逃逸分析，返回函数局部变量将变得可能，除此之外，逃逸分析还跟闭包息息相关，了解哪些场景下对象会逃逸至关重要。

### 逃逸策略

每当函数中申请新的对象，编译器会跟据该对象是否被函数外部引用来决定是否逃逸： 
1. 如果函数外部没有引用，则优先放到栈中； 
2. 如果函数外部存在引用，则必定放到堆中；

**注意，对于函数外部没有引用的对象，也有可能放到堆中，比如内存过大超过栈的存储能力**

### 逃逸场景

#### 指针逃逸
我们知道GC可以返回局部变量的指针，这其实是一个典型的变量逃逸案例，代码如下：
    
    package main

    type Student struct {
        Name string
        Age  int
    }

    func StudentRegister(name string, age int) *Student {
        s := new(Student) //局部变量s逃逸到堆

        s.Name = name
        s.Age = age

        return s
    }

    func main() {
        StudentRegister("Jim", 18)
    }
    
函数StudentRegister()内部s为局部变量，其值通过函数返回值返回，s本身为一指针，其指向的内存地址不会是栈而是堆，这就是典型的逃逸案例。

通过编译参数-gcflag=-m可以查年编译过程中的逃逸分析：

    D:\SourceCode\GoExpert\src>go build -gcflags=-m
    # _/D_/SourceCode/GoExpert/src
    .\main.go:8: can inline StudentRegister
    .\main.go:17: can inline main
    .\main.go:18: inlining call to StudentRegister
    .\main.go:8: leaking param: name
    .\main.go:9: new(Student) escapes to heap
    .\main.go:18: main new(Student) does not escape
    
可见在StudentRegister()函数中，也即代码第9行显示"escapes to heap"，代表该行内存分配发生了逃逸现象。

#### 栈空间不足逃逸

    package main

    func Slice() {
        s := make([]int, 1000, 1000)

        for index, _ := range s {
            s[index] = index
        }
    }

    func main() {
        Slice()
    }
    
上面代码Slice()函数中分配了一个1000个长度的切片，是否逃逸取决于栈空间是否足够大。 直接查看编译提示，如下：

    D:\SourceCode\GoExpert\src>go build -gcflags=-m
    # _/D_/SourceCode/GoExpert/src
    .\main.go:4: Slice make([]int, 1000, 1000) does not escape
    
我们发现此处并没有发生逃逸。那么把切片长度扩大10倍即10000会如何呢?

    D:\SourceCode\GoExpert\src>go build -gcflags=-m
    # _/D_/SourceCode/GoExpert/src
    .\main.go:4: make([]int, 10000, 10000) escapes to heap
    
我们发现当切片长度扩大到10000时就会逃逸。

实际上**当栈空间不足以存放当前对象时或无法判断当前切片长度时**会将对象分配到堆中。    

#### 动态类型逃逸

很多函数参数为interface类型，比如fmt.Println(a ...interface{})，编译期间很难确定其参数的具体类型，也人产生逃逸。 如下代码所示：

    package main

    import "fmt"

    func main() {
        s := "Escape"
        fmt.Println(s)
    }
    
上述代码s变量只是一个string类型变量，调用fmt.Println()时会产生逃逸：

    D:\SourceCode\GoExpert\src>go build -gcflags=-m
    # _/D_/SourceCode/GoExpert/src
    .\main.go:7: s escapes to heap
    .\main.go:7: main ... argument does not escape

#### 闭包引用逃逸

某著名的开源框架实现了某个返回Fibonacci数列的函数：

    func Fibonacci() func() int {
        a, b := 0, 1
        return func() int {
            a, b = b, a+b
            return a
        }
    }
    
该函数返回一个闭包，闭包引用了函数的局部变量a和b，使用时通过该函数获取该闭包，然后每次执行闭包都会依次输出Fibonacci数列。 完整的示例程序如下所示：

    package main

    import "fmt"

    func Fibonacci() func() int {
        a, b := 0, 1
        return func() int {
            a, b = b, a+b
            return a
        }
    }

    func main() {
        f := Fibonacci()

        for i := 0; i < 10; i++ {
            fmt.Printf("Fibonacci: %d\n", f())
        }
    }
    
上述代码通过Fibonacci()获取一个闭包，每次执行闭包就会打印一个Fibonacci数值。输出如下所示：

    D:\SourceCode\GoExpert\src>src.exe
    Fibonacci: 1
    Fibonacci: 1
    Fibonacci: 2
    Fibonacci: 3
    Fibonacci: 5
    Fibonacci: 8
    Fibonacci: 13
    Fibonacci: 21
    Fibonacci: 34
    Fibonacci: 55
    
Fibonacci()函数中原本属于局部变量的a和b由于闭包的引用，不得不将二者放到堆上，以致产生逃逸：

    D:\SourceCode\GoExpert\src>go build -gcflags=-m
    # _/D_/SourceCode/GoExpert/src
    .\main.go:7: can inline Fibonacci.func1
    .\main.go:7: func literal escapes to heap
    .\main.go:7: func literal escapes to heap
    .\main.go:8: &a escapes to heap
    .\main.go:6: moved to heap: a
    .\main.go:8: &b escapes to heap
    .\main.go:6: moved to heap: b
    .\main.go:17: f() escapes to heap
    .\main.go:17: main ... argument does not escape
    
### 逃逸总结

- 栈上分配的对象比在堆中分配的有更高的效率
- 栈上分配的内存不需要GC处理
- 堆上分配的内存使用完毕后会交给GC处理
- 逃逸分析目的是决定内存分配地址是栈还是堆
- 逃逸分析在编译阶段完成

## 编程Tips

思考一下这个问题：函数传递指针真的比传值效率高吗？ 我们知道传递指针可以减少底层值的拷贝，可以提高效率，但是如果拷贝的数据量小，由于指针传递会产生逃逸，可能会使用堆，也可能会增加GC的负担，所以传递指针不一定是高效的。

## 参考

https://rainbowmango.gitbook.io/