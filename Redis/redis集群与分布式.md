## redis集群发展史

### 主从复制+哨兵

架构图如下：(图片来自孤独烟的博客)

![upload successful](http://blogs.xinghe.host/images/pasted-136.png)


- 这里的sentinel就是哨兵，主要负责监控、通知与故障转移
- 监控：监控主从节点是否异常，相互发送心跳包来检测
- 通知：当其中一个节点发生故障，sentinel通过api脚本通知管理员
- 自动故障转移：当主节点挂了，从节点进行投票，选举出一个节点，这个节点成为新的主节点，然后重新建立主从关系

!> 工作原理就是，当Master宕机的时候，Sentinel会选举出新的Master，并根据Sentinel中client-reconfig-script脚本配置的内容，去动态修改VIP(虚拟IP)，将VIP(虚拟IP)指向新的Master。我们的客户端就连向指定的VIP即可！**

这个模式的缺点就是：

- 主从切换的时候会丢失数据
- redis只能单点写，不能水平扩容

### Proxy+主从复制+哨兵

这种集群架构比较古老了，我也没接触过，也没有研究过，这里就不介绍了，现在的架构一般都是redis-cluster架构，也就是下面要介绍的一种，如果有兴趣的同学可以私下研究一下这种架构，也是有缺点的

### redis-cluster

redis官网出的，会一直维护下去，很稳定

大厂都在用，百度贴吧、美团等

架构如下(图片来自孤独烟的博客)

![upload successful](http://blogs.xinghe.host/images/pasted-137.png)

#### cluster的数据结构

>集群的数据结构为一个叫 `clusterNode` 的数据结构，有兴趣的同学可以看《redis设计与实现》这本书，介绍了详细的数据结构

- clusterNode结构存储了节点的一些信息：名称、状态、ip地址、端口号以及一个clusterLink结构，clusterLink保存了连接节点有关的信息

- clusterLink结构存储了连接创建时间、TCP套接字描述符、输入缓冲区、输出缓冲区以及一个clusterNode 结构体，保存了与这个连接相关的节点的信息

- 其中redisClient结构与clusterLink结构很类似，只不过redisClient的套接字和缓冲区保存的是用于连接客户端的，而clusterLink的套接字和缓冲区保存的是用于连接节点的

- 最后每个节点都保存一个clusterState的数据结构，记录的当前节点视角下，集群的状态：比如集群是处于上线还是下线状态，集群的配置纪元等
  
#### 工作原理

1. 首先cluster是根据节点数量进行分片，一共16384个槽，进行节点数量的等分

2. 当你存储一个key-value的时候，客户端首先是把key进行计算，利用公式`HASH_SLOT=CRC16(key) mod 16384` 计算出key要存在那个分片上去

3. 查询也是一样，先利用公式查询在哪个分片上，然后再去那个分片上进行查询

4. 连上redis的客户端，使用命令 `CLUSTER MEET <ip> <port> ` 可以连接两个节点建立redis-cluster，然后可以使用 `CLUSTER NODES` 查看集群的节点信息
	
   建立连接的过程就是发送ping/pong包来建立连接，在各自节点上新建一个clusterNode结构添加到clusterState.nodes里面去，如下图所示“
    
![upload successful](http://blogs.xinghe.host/images/pasted-138.png)
    
!> 所以说，集群的每个节点不是存储的全部数据，而是各个节点加起来

#### 优缺点

##### 优点

1. 无需Sentinel哨兵监控，如果Master挂了，Redis Cluster内部自动将Slave切换Master
2. 可以进行水平扩容
3. 支持自动化迁移，当出现某个Slave宕机了，那么就只有Master了，这时候的高可用性就无法很好的保证了，万一master也宕机了，咋办呢？ 针对这种情况，如果说其他Master有多余的Slave ，集群自动把多余的Slave迁移到没有Slave的Master 中。

##### 缺点

1. 批量操作很坑
2. 资源隔离性较差，容易出现相互影响

## 面试题

### redis为什么这么快

1. 单线程，避免了线程切换带来的开销

2. 非阻塞I/O多路复用机制

3. 基于内存操作

### 为什么用redis

1. 性能：参考上一个问题

2. 并发：对于数据库的并发访问，会击垮数据库，需要一个缓存层面缓解数据库的压力

### redis的过期淘汰策略

redis是定时删除+惰性删除

定时删除：每个100ms redis会随机抽取一批key检查是否过期，过期了就删除，然而是随机的，就有可能一些key永远不会删除，导致内存越来越大，所以惰性删除出现了

惰性删除：每次访问这个key的时候，检查是否过期，如果过期，此时就会删除这个key，但是还会有一些key过期了也没有被访问到，所以出现了淘汰策略

- noeviction：不淘汰，内存满了返回错误
-  volatile-random：从设置过期时间的key中随机挑选一批key淘汰
-  volatile-lru：从设置过期时间的key中挑选最近最少使用的key淘汰（lru算法）
-  allkeys-random：从所有的key中随机选择一批key淘汰
-  allkeys-lru：从所有的key中挑选最近最少使用的key淘汰（lru算法）
-  volatile-ttl：从已过期的key中淘汰

### 实现一个lru算法

hashMap+双向链表

![upload successful](http://blogs.xinghe.host/images/pasted-139.png)

- 首先预先设置 LRU 的容量，如果存储满了，可以通过 O(1) 的时间淘汰掉双向链表的尾部
- 每次新增和访问数据，都可以通过 O(1)的效率把新的节点增加到对头，或者把已经存在的节点移动到队头。

实现原理：

1. save(key, value)，首先在 HashMap 找到 Key 对应的节点，如果节点存在，更新节点的值，并把这个节点移动队头。如果不存在，需要构造新的节点，并且尝试把节点塞到队头，如果LRU空间不足，则通过 tail 淘汰掉队尾的节点，同时在 HashMap 中移除 Key。

2. get(key)，通过 HashMap 找到 LRU 链表节点，因为根据LRU 原理，这个节点是最新访问的，所以要把节点插入到队头，然后返回缓存的值。

### redis事务

1. redis事务是一系列的redis命令的集合

2. redis事务不能回滚

3. 我们的redis-cluster架构是不支持事务的，因为数据不是在同一台机器上的，如果非要用的话要使用分布式事务等，相当复杂，不如不用
    
### 多数据库机制

1. 单机默认16个数据库，每个数据库相互隔离，数据互不影响

2. cluster架构是只有db0一个数据库

### cluster如何进行批量操作

1. 能不用批量就不用，少量的数据比如mget可以使用串行化的get

2. 必须使用的话，就使用hashtag来确保所有需要批量操作的key映射到同一个节点上去
	- 对于key为{foo}.student1、{foo}.student2，{foo}student3，这类key一定是在同一个redis节点上。因为key中“{}”之间的字符串就是当前key的hash tags， 只有key中{ }中的部分才被用来做hash，因此计算出来的redis节点一定是同一个!

### redis读写分离

-  redis的瓶颈并不在I/O吞吐，它是在内存里面操作
-  即使做来读写分离，也不会有太多的性能提升，反而需要兼顾主从延迟、主从一致性等问题
-  redis在生产环境主要是容量问题，单机最多10-20G，key太多来会影响性能，cluster分片机制已经保证来我们的性能，所以读写分离不会带来更多的收益
-  redis每秒10w次读写差不多可以应对99%的场景了，实在是不行可以做redis-cluter集群，读写分离的话还要关注数据一致性问题，主从延迟问题，得不偿失

### redis双写一致性

见另一篇文章

#### redis缓存击穿、缓存雪崩、缓存穿透

见另一篇文章

#### 如何解决redis的并发竞争key问题

-  如果操作不要求有顺序的话，使用redis的setnx实现分布式锁或者其他中间件如zookeeper实现分布式锁，每次操作去抢这个锁，抢到了去操作这个key
-  如果操作要求有顺序，就在我们写入数据库的时候，保存一个时间戳，如果a的时间是3：00，而b的时间为3:05，当b获取到锁的时候，设置了时间为3:05，然后a获取到锁，去设置value的时候，发现自己的时间戳比value的时间戳早，这时候就不做set操作了，以此类推
-  使用消息队列进行串行化


### 参考

https://www.cnblogs.com/rjzheng/p/10360619.html

《Redis的设计与实现》