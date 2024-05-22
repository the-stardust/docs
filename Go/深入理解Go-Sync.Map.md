# 前言

在 golang 开发过程中,很多时候用到了协程,并想要在协程处理任务之后,把结果聚合到 map 中,但是 go 的原生 map 是非线程安全的,所以方案就有
- map+mutex,写入前加锁,写完后释放锁 (在读多写少的场景下,锁的粒度太大存在效率问题:影响其他的元素操作)
- sync.map (减少加锁时间,读写分离,降低锁粒度,空间换时间,降低影响范围)

# 数据结构
```go
// sync.Map的核心数据结构
type Map struct {
    mu Mutex                        // 对 dirty 加锁保护，线程安全
    read atomic.Value                 // readOnly 只读的 map，充当缓存层
    dirty map[interface{}]*entry     // 负责写操作的 map，当misses = len(dirty)时，将其赋值给read
    misses int                        // 未命中 read 时的累加计数，每次+1
}

// 上面read字段的数据结构
type readOnly struct {
    m  map[interface{}]*entry // 
    amended bool // Map.dirty的数据和这里read中 m 的数据不一样时，为true
}

// 上面m字段中的entry类型
type entry struct {
    // 可见value是个指针类型，虽然read和dirty存在冗余情况（amended=false），但是由于是指针类型，存储的空间应该不是问题
    p unsafe.Pointer // *interface{}
}
```

## sync.Map 原理分析

![upload successful](../images/sync.map.png)

数据结构中的read 好比整个sync.Map的一个“高速缓存”，当goroutine从sync.Map中读数据时，sync.Map会首先查看read这个缓存层是否有用户需要的数据（key是否命中），
如果有（key命中）， 则通过原子操作将数据读取并返回，这是sync.Map推荐的快路径(fast path)，也是sync.Map的读性能极高的原因。

- 写操作：直接写入dirty（负责写的map）
- 读操作：先读read（负责读操作的map），没有再读dirty（负责写操作的map）

![upload successful](../images/sync.map2.png)


1. 通过 read 和 dirty 两个字段实现数据的读写分离，读的数据存在只读字段 read 上，将最新写入的数据则存在 dirty 字段上
2. 读取时会先查询 read，不存在再查询 dirty，写入时则只写入 dirty
3. 读取 read 并不需要加锁，而读或写 dirty 则需要加锁
4. 另外有 misses 字段来统计 read 被穿透的次数（被穿透指需要读 dirty 的情况），超过一定次数则将 dirty 数据更新到 read 中（触发条件：misses=len(dirty)）

## 优缺点

- 优点：Go官方所出；通过读写分离，降低锁时间来提高效率；
- 缺点：不适用于大量写的场景，这样会导致 read map 读不到数据而进一步加锁读取，同时dirty map也会一直晋升为read map，整体性能较差，甚至没有单纯的 map+mutex 高。
- 适用场景：读多写少的场景。

## 新增/修改过程

新增和修改操作比较复杂,需要先看下 read 中值的状态,然后再看 dirty 中的状态, 在 dirty == nil 的时候,还涉及搬迁过程

![upload successful](../images/sync.map4.png)


## 查询过程

>图片是网上找的,过程其实可以看源码,核心就是双重加锁 + miss 计数后替换

![upload successful](../images/sync.map3.png)

## 删除过程

删除过程氛围两部,第一步是直接删除 dirty 里面的数据,第二步其实是软删除,先标记read 里面的值为 nil,后删除

# 总结

通过阅读源码我们发现sync.Map是通过冗余的两个数据结构(read、dirty),实现性能的提升。为了提升性能，load、delete、store等操作尽量使用只读的read；
为了提高read的key击中概率，采用动态调整，将dirty数据提升为read；对于数据的删除，采用延迟标记删除法，只有在提升dirty的时候才删除。