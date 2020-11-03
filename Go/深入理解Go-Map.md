## Map的数据结构
>本文图片全部来自于 [Go专家编程](https://rainbowmango.gitbook.io/) 一书，非常幸运能发现这本书

Go语言的map底层是利用哈希表实现的，一个哈希表里面可以有多个节点，也就是bucket，而每个bucket就保存了map里面的一组键值对

map数据结构有src/runtime/map.go 中的hmap定义

	type hmap struct {
        count     int 		// map中当前保存元素的个数
        flags     uint8		// 标记哈希读写状态，竟态检测
        B         uint8  	// buckets数组的大小，可以容纳 2 ^ N 个bucket
        noverflow uint16 	// 溢出的bucket个数
        hash0     uint32 	// 哈希因子
        buckets   unsafe.Pointer // bucket数组指针，数组的大小为2^B
        noverflow uint16 		  // 溢出的哈希的数量的近似值
        oldbuckets unsafe.Pointer // map扩容时的老数据会存在这
        nevacuate  uintptr        // 扩容时已迁移的bucket个数
        extra *mapextra 		  // 保存溢出map的链表和未使用的溢出map数组的首地址	
    }
    type mapextra struct {
        overflow    *[]*bmap	// 溢出的bucket地址，bmap是bucket的数据结构
        oldoverflow *[]*bmap	

        nextOverflow *bmap
    }
    
<!--more--> 

下面展示一个简单的map

![upload successful](http://blogs.xinghe.host/images/pasted-72.png)

此map含有四个bucket， hmap.B = 2, 元素经过哈希函数运算过后会落到其中一个bucket中进行存储，查找过程类似

bucket很多时候被翻译为桶，所谓的**哈希桶**，时间上就是bucket

## bucket的数据结构

	type bmap struct {
      
        tophash [bucketCnt]uint8    // //存储哈希值的高8位
      
    }

每个bucket可以存储8个键值对
- tophash是个长度为8的数组，哈希值相同（准确的说是哈希值低位相同）存入当前bucket时会将哈希值的高位存储在数组中，以便以后匹配
- key-value数据，存放顺序是key/key/key/...value/value/value，如此存放是为了节省字节对齐带来的空间浪费（这点我也不懂）

下图展示bucket存储8个key-value：
![upload successful](http://blogs.xinghe.host/images/pasted-73.png)

## 哈希冲突

当有两个或以上数量的键被哈希到同一个bucket中，就是发生了哈希冲突，go语言是利用链表法来解决哈希冲突的，由于每个bucket可以存储8个key-value，所以同一个bucket存放超过8个key-value的时候，就会再次创建一个bucket，用类似链表的方式将bucket串联起来

下图展示发生哈希冲突后的map：

![upload successful](http://blogs.xinghe.host/images/pasted-74.png)

bucket的数据结构指示下一个bucket的指针成为overflow bucket，意为当前bucket溢出的部分，事实上哈希冲突不是什么好事，好的哈希算法可以保证哈希值的随机性，避免冲突过多，冲突过多就会进行扩容

## 负载因子

负载因子是形容哈希表冲突的，公式为
	
    负载因子 = 所有键数量 / bucket数量
    
对于一个bucket树立等于4 包含4个键值的哈希表来说，负载因子就等于1

哈希表需要将负载因子控制在合适的大小，超过设定的阀值，就会进行rehash
- 哈希因子过小，说明空间利用率低
- 哈希因子过大，说明冲突严重

每个哈希表的实现对负载因子容忍程度不同，比如redis中实现负载因子大于1的时候，就会触发rehash，而**Go语言负载因子超过6.5的时候，才会rehash**，因为Go语言的bucket可以存储8个key-value，而redis每个bucket只能存储一个key-value，所以Go可以容忍更高的负载因子

## 渐进式扩容

### 扩容的前提条件

为了保证访问效率，当前新的元素添加到map的时候，都会检查是否需要扩容，扩容就是以空间换时间的手段，触发扩容有两个条件
- 负载因子大于6.5的时候
- overflow与 2^15的时候，即overflow超过32768的时候

### 增量扩容

当负载因子过大的时候， 就会新建一个bucket，新的bucket的大小是原来的两倍，考虑如果mao存储了数以亿计的key-value，一次性搬迁会造成比较大的延时，所以每次访问map时会触发搬迁，每次搬迁只会搬迁2个key-value

下图展示了包含一个bucket满载的map(为了描述方便，图中bucket省略了value区域):

![upload successful](http://blogs.xinghe.host/images/pasted-75.png)

当前map存储了7个key-value，只有一个bucket，此时的负载因子是7，当再次发生冲突的时候，就会发生扩容现象，扩容之后再将新插入的键插入新的bucket

当第8个key-value插入的时候，将会发生扩容，如图所示

![upload successful](http://blogs.xinghe.host/images/pasted-76.png)

hmap中的oldbuckets成员指身原来的bucket，而buckets指向了新申请的bucket，新的key-value被插入到了新的bucket中，后续对map的访问操作会触发迁移，将oldbuckets中的key-value慢慢搬迁到新的bucket中，直到搬迁完毕，删除oldbuckets

搬迁完成如图所示

![upload successful](http://blogs.xinghe.host/images/pasted-77.png)

数据搬迁过程中，原来的key-value会存储在新的bucket前面，新插入的key-value会村雨新的bucket的后面，实际搬迁过程非常复杂

### 等量扩容

所谓等量扩容，实际上不是扩大容量，是一种类似增量扩容的搬迁操作，因为有一种情况是，不断的增删，而键值对正好集中在一小部分的bucket，这样会造成overflow的bucket非常的多，而每个bucket又很稀疏，负载因子就会偏小，无法进行增量扩容，如下图所示

![upload successful](http://blogs.xinghe.host/images/pasted-78.png)

上图中，overflow的bucket大部分都是空的，访问效率会很差，此时进行一次等量扩容，即**buckets的树立不变，经过重新组织后的overflow的bucket会减少**，即节省了空间又增加了访问效率

## 查找过程

查找过程如下：
1. 根据key值计算出哈希值
2. 取哈希值的低位，与hmap.B取模确定bucket的位置
3. 去哈希值的高位在tophash中查找
4. 如果tophash[i]中存储的值也哈希值相等，则去找改bucket中与key值比较，
5. 当前bucket中没找到，就去下个overflow的bucket中查找
6. 如果此时map处在搬迁的过程中，则优先从oldbuckets中查找

**注：** 如果查找不到，也不会返回空值，而是返回当前类型的0值

## 插入过程

新元素的插入如下
1. 根据key计算出哈希值
2. 取哈希值的低位与hmap.B取模确定bucket的位置
3. 查找该key是否存在，存在则直接更新值，不存在则将key插入