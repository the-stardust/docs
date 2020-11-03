
#### 一主二从
配从(库)不配主(库)

每次与master断开之后，都需要重新连接，除非你配置进redis.conf文件

info replication 查看主从关系
<!--more-->
1.配置步骤

（1）拷贝多个redis.conf文件

（2）开启daemonize yes

（3）pid文件名字

（4）指定端口

（5）log文件名字

（6）dump.rdb名字

###### 一主二从问题汇总

1 切入点问题,slave1、slave2是从头开始复制还是从切入点开始复制?刚连上的slave会全量复制一次，接下来都是增量复制

2 从机不可以写

3 主机shutdown后情况如何？从机是原地待命

4 主机又回来了后，主机新增记录，从机会自动连接，然后复制数据

5 其中一台从机down后,重启后变成master，和原来的体系会脱离，需要重新手动配置，除非写进redis.conf配置文件

#### 薪火相传
1.上一个Slave可以是下一个slave的Master，Slave同样可以接收其他slaves的连接和同步请求，那么该slave作为了链条中下一个的master,可以有效减轻master的写压力

2.中途变更转向:会清除之前的数据，重新建立拷贝最新的

4.slaveof 新主库IP 新主库端口

#### 反客为主
1.SLAVEOF no one ，使一个从机变成主机 变成master，然后另一个slaveof ip port后，进行全量复制

2.主机恢复后，还是master，但是与另一个体系脱离

#### 哨兵模式

1.就是自动化的反客为主，主机挂了之后，从机投票出来一个主机，形成新体系

2.配置步骤：

1.自定义的/myredis目录下新建sentinel.conf文件，名字绝不能错

2. sentinel monitor 被监控数据库名字(自己起名字) 127.0.0.1 6379 1 （上面最后一个数字1，表示主机挂掉后salve投票看让谁接替成为主机，得票数多少后成为主机）

3.启动哨兵：redis-sentinel /myredis/sentinel.conf

4.哨兵流程：

（1）原有的master挂了（2）投票新选（大概一分钟不到）（3）重新主从继续开工,info replication查查看（4）如果之前的master重启回来，会成为slave

5.一组sentinel能同时监控多个Master

复制的原理：
1.slave启动成功连接到master后会发送一个sync命令

2.Master接到命令启动后台的存盘进程，同时收集所有接收到的用于修改数据集命令，在后台进程执行完毕之后，master将传送整个数据文件到slave,以完成一次完全同步

3.全量复制：而slave服务在接收到数据库文件数据后，将其存盘并加载到内存中。

4.增量复制：Master继续将新的所有收集到的修改命令依次传给slave,完成同步

5.但是只要是重新连接master,一次完全同步（全量复制)将被自动执行