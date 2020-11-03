## 前言

电商或者大数据，肯定会接触搜索这一块，这几年比较火的就是elasticSearch-ES，是基于java的lucene的分布式搜索引擎，前几年大家一般都用solr，这两年都转型到了es上

面试官上来肯定会问你有没有接触过es？es的分布式架构设计能介绍下么？了解lunece么？倒排索引的原理解释下

这些最简单最基础的知识点肯定是要了解的


## es分布式架构原理

### 基础概念

> index -> type -> mapping -> document -> field。
        
但是在es 7.0 版本以后，type的概念基本移除，一个index只有一个type

- index ： 类比于mysql的数据表table，
- type： mysql里面没有这种概念，意思是同一个index里面，两个type少数字段不一样
- mapping： mysql里面的建表语句，代表idnex的数据结构设计
- document：类比mysql里面的一条数据
- field：字面意思，类比mysql里面的字段

### 分布式

每个index可以有多个shard，每个shard存储部分数据，每个shard可以有多个副本(replica shard)；

多个shard是支持**横向扩展，提高性能**，多个副本是**高可用**

每个shard分布在不同的机器上，每个shard的副本分布在不同于当前shard的机器上，这样一来，就算有一台机器宕机，其他机器上也有副本数据

![upload successful](http://blogs.xinghe.host/images/pasted-151.png)

每个shard有一个peimary shard，负责写入数据然后同步给其他的replica shard

es集群有多个节点，会选举出来一个master节点，master节点主要是管理工作，负责切换primary shard和replica shard的身份、维护索引的元数据等

如果master节点挂了，会重新选举出来一个master节点

如果非master节点挂了，master节点会把宕机的那台机器的primary shard 身份转义到没有宕机的replica shard 上，这个replica shard 就会成为新的primary shard，**如果宕机的机器恢复了的话，修复后的节点自动成为replica shard**

## 工作原理

### 写数据过程

- 客户端选择一个node发送请求，这个node就成为了**coordinating node(协调节点)**
- coordinating node 对 document 进行路由，路由到对应的node上去
- 实际的node上的primary node处理请求，然后把数据同步给replica node
- coordinating node 如果发现primary node和所有的replica node都处理完之后，返回给客户端响应结果

![upload successful](http://blogs.xinghe.host/images/pasted-152.png)

### 读数据过程

可以通过doc id 来查询，会根据doc id 进行hash，判断当前doc id 分配到哪个shard上去，从那个shard查询

- 客户端发送一个请求到**任意**一个node，成为coordinating node
- coordinating node 对doc id 路由，将请求发送给对应的node，此时会使用round-robin**随机轮训算法**，在primary shard以及其所有的replica中随机选择一个，让读请求负载均衡
- 接受请求的node返回document给coordinating node
- coordinating node 返回 document 给客户端

### 搜索数据过程

- 客户端发送请求到任意一个node，此node成为coordinating node
- coordinating node 转发请求到**所有**shard对应的primary shard 或者 replica shard上(随便一个都可以)
- query phase:每个shard将自己的搜索结果(其实就是一些doc id)返回给coordinating node 由coordinating node进行数据的合并、分页、排序，产出最终结果
- fetch phase:然后coordinating node 根据结果的 doc id 从各个节点上拉取**实际的document数据**，返回给客户端

### 写数据底层原理

!> 写请求是写入 primary shard，然后同步给所有的 replica shard；读请求可以从 primary shard 或 replica shard 读取，采用的是随机轮询算法

![upload successful](http://blogs.xinghe.host/images/pasted-153.png)

- buffer：先写入内存buffer，在buffer里的时候，数据是搜索不到的，同时将数据写入translog

- refresh：如果buffer快满了，或者到一定时间(默认每隔一秒钟)，就会将buffer数据refresh到一个新的segment file 中，但是此时数据不是直接进入segment file ，而是先进入到 os cache，这个过程就是refresh

- os cache：操作系统有一个 os cache 的东西，即系统缓存，数据写入磁盘之前，会先进入操作系统级别的一个内存缓存中去，只要**buffer的数据写入到 os cache 中，数据就会被搜索到**

重复上面的步骤，新的数据会不断的进入buffer和translog，不断的又将buffer写入到一个又一个segment file 中，每次refresh 完 buffer 清空，translog 保留，随着这个过程推进，translog 会变得越来越大，当translog 到达一定大小的时候，或者默认30分钟，会执行一次commit

- commit：操作发生的第一步，就是将 buffer 中的数据刷到 os cache 中去，清空 buffer， 然后将一个 commit point 写入磁盘文件，里面标识着这个 commit point 对应的所有 segment file ，同时强行将 os cache 里面的数据 fsync 到磁盘上，最后**清空**现有的 translog 日志文件，重启一个 translog，此时 commit 操作完成(也叫flush)

- translog：translog 日志文件的作用是：你执行 commit 操作之前，数据要么在 buffer 中，要么在 os cache 中，他们都是内存，断电后数据就消失了，当机器恢复的时候，会读取 translog 的内容恢复数据

!> commit 操作是默认5秒一次，也就是说es可能会有5秒的数据丢失几率，可以设置每次写操作都必须 fsync 到磁盘上，但是性能会差很多**

#### 总结

数据先写入内存 buffer ，然后每个1s，将数据 refresh 到 os cache 中，到了 os cache，数据就能被搜索到(所以说es是**准实时**，有1s延迟)，每个5s，将数据写入 translog 中(此时宕机，丢失5s的数据)，translog 打到一定程度或者每隔30mins，会触发 commit 操作，将缓冲区数据 flush 到 segment file 磁盘文件中去

**数据写入 segment file 之后，同时就建立好了倒排索引。**

### 删除/更新底层原理

如果是删除操作，commit 的时候会生成一个 .del 文件，里面将某个 doc 标识成 deleted 状态，那么搜索的时候根据 .del 文件就知道这个 doc 是否被删除了

如果是更新操作，就是将原来 doc 标识成 deleted ，然后重新写入一条数据

buffer 每 refresh 一次，就会产生一个segment file，所以默认情况下是1s一个 segment file

这样下来 segment file 会越来越多，此时会定期执行 merge，每次 merge 的时候，将多个 segment file 合并成一个，同时这里会将标识为 deleted 的doc给**物理删除掉**，然后将新的 segment file 写入磁盘，这里会写一个 commit point，标识所有新的 segment file，然后打开 segment file功搜索使用，同时删除旧的 segment file

## 倒排索引

倒排索引就是**关键词到文档 ID 的映射**，每个关键词都对应着一系列的文件，这些文件中都出现了关键词。

![upload successful](http://blogs.xinghe.host/images/pasted-154.png)

其实倒排索引还记录了更多的信息，比如文档频率信息，标识在文档集合中有多少个文档包含某个单词

要注意倒排索引的两个重要细节：

- 倒排索引中的所有词项对应一个或多个文档；
- 倒排索引中的词项根据字典顺序升序排列

**上面只是一个简单的栗子，并没有严格按照字典顺序升序排列。**

## 搜索优化

说实话，es 性能优化是没有什么银弹的，啥意思呢？就是**不要期待着随手调一个参数，就可以万能的应对所有的性能慢的场景。**也许有的场景是你换个参数，或者调整一下语法，就可以搞定，但是绝对不是所有场景都可以这样。

所有es是根据自己的业务场景进行优化的，没有固定优化的策略


### 性能优化的杀手锏——filesystem cache

![upload successful](http://blogs.xinghe.host/images/pasted-155.png)

es的搜索引擎严重依赖底层的 filesystem cache，你如果给 filesystem cache 更多的内存，尽量让内存足以容纳所有的 idx segment file 索引数据文件，那么你的搜索就全部基于内存，性能非常高

磁盘搜索和内存搜索，性能相差一个数量级 秒级和毫秒级的差别

归根结底，你要让 es 性能要好，最佳的情况下，就是你的机器的内存，至少可以容纳你的总数据量的一半。

### 数据预热

把一些重要的、常用的数据进行预热

就是写一个脚本，每隔几秒就访问一下热点数据，将热点数据刷到 filesystem cache 中去，后面用户来看这些热点数据的时候，走的就是内存搜索，非常快

### 冷热分离

类似于mysql的水平拆分，将大量的访问很少、频率很低的数据，单独写入一个索引文件，然后热数据写入另一个索引文件，这样可以确保热数据被预热之后，尽量都保留在 filesystem cache 中，**别让冷数据给冲刷掉**

你看，假设你有 6 台机器，2 个索引，一个放冷数据，一个放热数据，每个索引 3 个 shard。3 台机器放热数据 index，另外 3 台机器放冷数据 index。然后这样的话，你大量的时间是在访问热数据 index，热数据可能就占总数据量的 10%，此时数据量很少，几乎全都保留在 filesystem cache 里面了，就可以确保热数据的访问性能是很高的。

但是对于冷数据而言，是在别的 index 里的，跟热数据 index 不在相同的机器上，大家互相之间都没什么联系了。如果有人访问冷数据，可能大量数据是在磁盘上的，此时性能差点，就 10% 的人去访问冷数据，90% 的人在访问热数据，也无所谓了。

### document 模型设计

对于 MySQL，我们经常有一些复杂的关联查询。在 es 里该怎么玩儿，es 里面的复杂的关联查询尽量别用，一旦用了性能一般都不太好。

**秉承一个观点，不要再es里面进行复杂查询，尽量在程序里面协调**

因为性能真的很差

### 分页

es的深度分页很坑，性能差到5-10秒，因为es是分布式的，你查询第100页这种，每个shard都会查询1000条数据，然后推给coordinating node ，coordinating node 再根据需求进行合并、排序、再分页，性能差到极致

#### 不允许深度分页

跟mysql类似，分布式的mysql也是不允许分页的，性能很差

#### 类似微博的下拉翻页

不让用户跳转翻页，可以一直下拉翻页这种

es有scroll api，scroll 会一次性给你生成 **所有数据的一个快照**，通过游标scroll_id移动获取下一页，基本上是毫秒级的性能，这样子你初始化的时候就必须指定scroll 参数，高速es保存此次搜索的上下文多长时间

除了scroll api 也可以用 search_after来做，search_after 的思想是利用前一页的结果来帮助检索下一页，这样也不允许跳转页面，也是一页一页往下翻，就是需要一个字段来作为sort字段

## es部署架构

面试官肯定问你们的es是怎么部署的，看看你是否实际操作过es部署，因为理论和实践是不一样的

可以简单说一下就行了，具体的深入问题你可以说不太清楚，你们公司是专业运维搞的这些，你只是辅助工作

- es生产集群我们部署了5台机器，每台机器是6核64G,一共是320G内存
- 我们es机器的日增长量大概是2000万条，大概是500MB，月增长大概6亿数据，15G，目前系统稳定运行了几个月了，数据总量大概在100G左右
- 目前线上5个索引(结合自己的业务来说),每个索引20G，每个索引分配8个shard，比默认的5个多了3个shard