## 参数
### log.dirs & log.dir
log.dirs:指定了 Broker 需要使用的若干个文件目录路径

log.dir:表示单个路径

只用配置log.dirs就行了，并且一定要配置多个路径，格式为CSV格式，中间用逗号隔开
/home/kafka1,/home/kafka2,/home/kafka3
- 有条件的话挂载在不同的物理盘，因为同时读写多个物理盘比单块磁盘效率高
- 实现故障转移，Failover

### zookeeper相关
#### zookeeper.connect 
这也是一个 CSV 格式的参数，比如我可以指定它的值为zk1:2181,zk2:2181,zk3:2181。
2181 是 ZooKeeper 的默认端口。

### broker相关

#### listeners
监听器，告诉外部连接者要通过什么协议访问指定主机名和端口开放的 Kafka 服务
#### advertised.listeners
Advertised 的含义表示宣称的、公布的，就是说这组监听器是 Broker 用于对外发布的

监听器的参数格式为：若干个逗号分隔的三元组，每个三元组的格式为<协议名称，主机名，端口号>，
协议名称：比如SSL使用SSL或TLS加密传输，PLAINTEXT表示明文传输，也可以自定义协议，比如 CONTROLLER: //localhost:9092。

一旦你自己定义了协议名称，你必须还要指定listener.security.protocol.map参数告诉这个协议底层使用了哪种安全协议，
比如指定listener.security.protocol.map=CONTROLLER:PLAINTEXT表示CONTROLLER这个自定义协议底层使用明文不加密传输数据

主机名：**Broker 端和 Client 端应用配置中全部填写主机名**买最好不要填ip

### topic相关
**topic相关参数会覆盖broker的参数**

#### auto.create.topics.enable
是否允许自动创建 Topic。最好设置为false，

#### unclean.leader.election.enable
是否允许 Unclean Leader 选举。false，表示不允许副本数据比较落后的节点竞争leader，因为会造成数据丢失

#### auto.leader.rebalance.enable
是否允许定期进行 Leader 选举。最好false，这个参数是**定期换leader，而不是重新选举**，

#### retention.ms
规定了该 Topic 消息被保存的时长。默认是 7 天

#### retention.bytes
规定了要为该 Topic 预留多大的磁盘空间。和全局参数作用相似，这个值通常在多租户的 Kafka 集群中会有用武之地。当前默认值是 -1，表示可以无限使用磁盘空间。


### 数据留存相关

#### log.retention
log.retention.{hour|minutes|ms}：这是个“三兄弟”，都是控制一条消息数据被保存多长时间。从优先级上来说 ms 设置
最高、minutes 次之、hour 最低。一般都是设置hour，比如log.retention.hour=168表示默认保存 7 天的数据

#### log.retention.bytes
这是指定 Broker 为消息保存的总磁盘容量大小。一般来说都是-1，表示多大的数据都可以

#### message.max.bytes
控制 Broker 能够接收的最大消息大小。默认值不到1mb，太小了，可以设置高一点

## 分区策略
所谓分区策略是决定生产者将消息发送到哪个分区的算法

**如果指定了 Key，那么默认实现按消息键保序策略；如果没有指定 Key，则使用轮询策略**

可以配置生产者端的partitioner.class，在编写生产者程序的时候，编写一个具体的类实现org.apache.kafka.clients.producer.Partitioner接口
，这个借口很简单，就定义了partiton()和close()，只需要实现partition()方法就行了

这个方法接受参数 int partition(String topic, Object key, byte[] keyBytes, Object value, byte[] valueBytes, Cluster cluster);

这里的topic、key、keyBytes、value和valueBytes都属于消息数据，cluster则是集群信息（比如当前 Kafka 集群共有多
少主题、多少 Broker 等）。Kafka 给你这么多信息，就是希望让你能够充分地利用这些信息对消息进行分区，
计算出它要被发送到哪个分区中

### 轮询策略
也称 Round-robin 策略，即顺序分配

**轮询策略有非常优秀的负载均衡表现，它总是能保证消息最大限度地被平均分配到所有分区上，故默认情况下它是最合理的分区策略，也是我们最常用的分区策略之一。**

### 随机策略
也称 Randomness 策略。所谓随机就是我们随意地将消息放置到任意一个分区上，如果要实现随机策略版的 partition 方法，很简单，只需要两行代码即可：
```java
List<PartitionInfo> partitions = cluster.partitionsForTopic(topic);
return ThreadLocalRandom.current().nextInt(partitions.size());
```
### 按消息键保序策略

Kafka 允许为每条消息定义消息键，简称为 Key，一旦消息被定义了 Key，那么你就可以保证同一个 Key 的所有消息
都进入到相同的分区里面，由于每个分区下的消息处理都是有顺序的，故这个策略被称为按消息键保序策略
![](../images/11212121.png)

```
List<PartitionInfo> partitions = cluster.partitionsForTopic(topic);
return Math.abs(key.hashCode()) % partitions.size();
```

**key+多分区可以保证消息的有序性，但是会降低吞吐量，因为只能一个消费者消费一个partition了，而通常我们的系统都是多个消费者**

## 消息压缩

**Producer 端压缩、Broker 端保持、Consumer 端解压缩**

生产者程序配置compression.type参数表示压缩算法 ，broker也有一个参数compression.type，默认值是producer表示和
producer保持一致，如果两个配置不一致，那么就会导致，生产者发送消息到broker，broker会解压，然后按照briker设置的
压缩算法重新压缩，导致线上cpu飙升，但是**broker也会进行解压缩，因为要对消息执行各种验证，**

Kafka 会将启用了哪种压缩算法封装进消息集合中，当消息到达 Consumer 端后，由 Consumer 自行解压缩还原成之前的消息。





