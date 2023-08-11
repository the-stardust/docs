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

## 消息丢失

- kafka是能保证消息不丢失的，只要是已提交的消息，什么是已提交呢，就是**若干个**broker接收到消息并写入到日志中，这
个**若干个**是需要配置，可能是一个broker，也可能是所有broker

### 生产者丢失消息

- 不要使用producer.send(msg),要是有producer.send(msg,callback)，使用回调函数来确认消息是否真正的被接收并写入日志
- 根据配置可以是有一个broker写入日志成功就认为消息保存成功，也可以是所有broker写入日志才认为保存成功

### 消费者丢失消息

- 保持先消费消息，再提交offset，
- 避免消息不丢失简单，但是就造成了极有可能消息重复消费，所以消费者端需要再业务上保证幂等性

## 参数设置
- acks = all :所以broker都保存了消息，才认为消息被接收了
- retries = MAX; MAX为一个很大的值，表示当消息发送失败或者写入失败，就重试MAX次，直到写入成功
- unclean.leader.election.enable = false；如果一个broker落后leader很多，就不允许它进行leader的竞争
- replication.factor >= 3；保证消息有多个副本，即使其中有机器宕机，也有副本顶上去
- min.insync.replicas > 1；控制消息至少被写入到多少个副本才算成功提交，大于1可以提高消息的持久性和安全性
- replication.factor > min.insync.replicas；如果两者相等，只要有一个机器挂了，那整个分区就不能工作了，
推荐设置成 replication.factor = min.insync.replicas + 1
- enable.auto.commit = false；消费者端的一个参数，把消息自动提交关闭了，让程序进行消息消费后手动提交，
保证消息不丢失，但是会造成重复消费

## 消息过滤(拦截器)

kafka提供消息过滤机制，类似laravel中间件，在消息发送之前和消息发送成功之后
- 生产者端可以实现在消息发送前对消息进行过滤+"美容"，
    - onSend:方法在消息发送之前调用
    - onAcknowledgement:这个会在消息提交成功或者发送失败后调用，前面提到的发送消息的callback，onAcknowledgement
    要早于callback调用，但是onAcknowledgement和onSend不在一个线程中调用
- 消费者端可以实现ConsumerInterceptor接口
    - onConsume：改方法在消息返回给consumer程序之前调用，也就是正式处理消息之前
    - onCommit：消费者端提交offset后调用，可以做一些日志以及统计的操作
        
!> Kafka 拦截器可以应用于包括客户端监控、端到端系统性能检测、消息审计等多种功能在内的场景


## kafka机制
### Kafka的消息传输模型是什么？请描述消息发布和订阅的过程。

#### Kafka是一个分布式流处理平台，主要用于高吞吐量的实时数据管道和发布/订阅系统。其核心概念包括以下几点：

- 消息：Kafka通过消息进行数据传输，是可发布和可订阅的元素。
- 主题（Topic）：消息在Kafka中以主题的形式组织和分类存储，每个主题可以有多个生产者和消费者。
- 分区（Partition）：主题可以分为多个分区，每个分区是一个有序、不可变的消息序列。
- 生产者（Producer）：负责将消息发布到Kafka集群中的主题。
- 消费者（Consumer）：从Kafka集群中的主题订阅并消费消息。
- 副本（Replication）：Kafka使用副本机制来提供容错和可靠性，每个分区可以有多个副本。

#### Kafka如何处理消息的持久化和可靠性传递？

1. Kafka的消息传输模型是基于发布/订阅模式的。消息发布过程包括：

- 生产者将消息发送到指定主题的分区中。
- Kafka通过写入磁盘的方式将消息持久化存储，以确保消息的可靠性。
- 消息按照分区和时间顺序被追加到日志文件（称为日志分段）中。
2. 消息订阅过程包括：

- 消费者通过指定主题和分区进行订阅。
- 每个消费者维护一个消息偏移量来跟踪其消费位置。
- Kafka将消息以批量的方式传递给消费者，消费者可以按照自己的需求进行处理。


### Kafka处理消息的持久化和可靠性传递是通过以下机制实现的：

1. 消息被追加到日志文件（日志分段）中，并通过操作系统的页缓存提供磁盘持久化能力。
2. 生产者可以选择使用acks配置来指示Kafka在接收到消息后发送确认响应，确保消息的可靠性传递。
3. Kafka使用复制机制来提供副本级别的容错能力，每个分区可以有多个副本分布在不同的服务器上。

### Kafka的分区和副本是什么？它们的作用是什么？
分区和副本是Kafka中的重要概念：

- 分区将主题划分为多个有序的数据片段，每个分区都有一个唯一的标识符。分区使得Kafka能够在集群中并行处理消息。
- 副本是分区的副本，用于提供高可用性和故障恢复能力。每个分区可以有多个副本，其中一个被称为领导者（Leader），其余的副本称为追随者（Follower）。

### 什么是消费者组（Consumer Group）？它在Kafka中的作用是什么？
消费者组是一组消费者的集合，共同消费一个或多个主题的消息。消费者组在Kafka中起到以下作用：

- 提供负载均衡：主题中的每个分区只能由一个消费者组内的一个消费者进行消费，其他消费者处于空闲状态。
- 提供容错性：如果某个消费者失败，消费者组中的其他消费者会接管其分区并继续消费。

### Kafka中的消息偏移量（Offset）是什么？为什么重要？

消息偏移量是消费者在每个主题分区上的唯一标识，用于指示消费者在分区中所消费的位置。偏移量是一个64位整数，用于跟踪消费进度。它对于实现精确的消费位置和失败恢复非常重要。

### Kafka如何进行水平伸缩？请描述一下分区再均衡的过程。

Kafka通过水平伸缩来提高吞吐量和处理能力。水平伸缩的过程主要涉及到分区的再均衡，包括以下几个步骤：

1. 增加或删除主题的分区数量。
2. 消费者组重新分配分区：消费者组中的消费者重新分配新的分区以实现负载均衡。
3. 分区迁移：Kafka将分区从一个Broker节点迁移到其他Broker节点，以平衡数据分布和负载。

### 在Kafka中，什么是Producer的acks配置？它有哪些可选的值？

Producer的acks配置用于控制消息的可靠性。它指定了生产者在发送消息后等待多少个副本的确认响应才认为消息被成功发送。常用的可选值包括：

- 0：生产者不会等待副本的确认响应，将消息立即发送。这种配置具有最低的延迟，但没有任何保证消息是否成功送达。
- 1：生产者会等待分区的领导者副本确认消息，但不需要等待其他副本的确认。这种配置提供了一定程度的消息传递可靠性。
- -1 或 all：生产者会等待所有副本都确认消息后才认为消息被成功发送。这种配置提供了最高的消息传递可靠性。

### 你能解释一下Kafka的零拷贝机制吗？

Kafka的零拷贝机制是指在数据传输过程中，避免了数据从内核空间到用户空间的多次拷贝操作。通过使用"sendfile"系统调用、mmap和"scatter/gather"技术，
Kafka能够直接在内核空间和网络堆栈之间传输数据，而无需将数据复制到用户空间。

### 监控和调优Kafka集群的性能可以使用以下方法：

- 使用监控工具：例如Kafka自带的JMX监控或第三方监控工具，用于实时监控集群的状况、性能指标和健康状态。
- 配置参数优化：调整Kafka的配置参数，根据集群规模和使用场景进行优化，包括副本数量、内存分配、磁盘IO等。
- 分区和副本配置：根据负载和容错需求，合理设置主题的分区和副本数量。
- 合理的网络拓扑：将Kafka Broker部署在不同的机架和数据中心，以提供高可用性和容灾能力。
- 数据存储优化：选择适当的存储介质和文件系统，并进行定期的磁盘清理和压缩操作。
- 定期监视和故障排除：通过日志分析、错误日志和警报来监视Kafka集群的运行状况，并定期进行故障排查和问题解决。




