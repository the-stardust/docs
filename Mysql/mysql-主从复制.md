## 原理
- 1.master将改变记录到二进制日志（binary log）。这些记录过程叫做二进制日志事件，binary log events
- 2.slave将master的binary log events拷贝到它的中继日志（relay log）
- 3.slave重做中继日志中的事件，将改变应用到自己的数据库中。 MySQL复制是异步的且串行化的

## 复制的基本原则
1.每个slave只有一个master

2.每个slave只能有一个唯一的服务器ID

3.每个master可以有多个salve

4.复制的最大问题就是数据延迟，高并发状态下，刚写入的数据会读取不到，因为slave同步数据会有延时

### 有三种同步机制
	
- 异步复制：主库在执行完客户端提交的事务后会立即将结果返给给客户端，并不关心从库是否已经接收并处理，这样就会有一个问题，主如果crash掉了，此时主上已经提交的事务可能并没有传到从上，如果此时，强行将从提升为主，可能导致新主上的数据不完整。
	
- **半同步复制：就是master提交事务后，并写到relay log中才返回给客户端 主要使用这个**<br>
		当Master上开启半同步复制功能时，至少有一个slave开启其功能。当Master向slave提交事务，且事务已写入relay-log中并刷新到磁盘上，slave才会告知Master已收到；若Master提交事务受到阻塞，出现等待超时，在一定时间内Master 没被告知已收到，此时Master自动转换为异步复制机制；
	
- 全同步复制：指当主库执行完一个事务，所有的从库都执行了该事务才返回给客户端。因为需要等待所有从库执行完该事务才能返回，所以全同步复制的性能必然会
    
## 主从配置
### 1.要求
- mysql版本一致且后台以服务运行
- 主从都配置在[mysqld]结点下，都是小写

### 2.主机修改配置

- **（1）[必须]主服务器唯一ID server-id=1**
- **（2）[必须]启用二进制日志 log-bin=自己本地的路径/data/mysqlbin**
- （3）[可选]启用错误日志 log-err=自己本地的路径/data/mysqlerr
- （4）[可选]根目录 basedir="自己本地路径"
- （5）[可选]临时目录 tmpdir="自己本地路径"
- （6）[可选]数据目录 datadir="自己本地路径/Data/"
- （7）read-only=0 主机，读写都可以
- （8）[可选]设置不要复制的数据库 binlog-ignore-db=mysql
- （9）[可选]设置需要复制的数据库 binlog-do-db=需要复制的主数据库名字 默认全部复制

### 3.从机修改配置

- **（1）[必须]从服务器唯一ID**
-  （2）[可选]启用二进制日志

### 4.关闭虚拟机linux防火墙 

   关闭虚拟机linux防火墙 service iptables stop，因修改过配置文件，请主机+从机都重启后台mysql服务
    
### 5.主机授权从机

- **（1）GRANT REPLICATION SLAVE ON *.* TO 'zhangsan'@'从机器数据库IP' IDENTIFIED BY '密码';**
- （2）flush privileges; 刷新权限
- （3）show master status; 查询master的状态 记录下File和Position的值
- （4）**执行完此步骤后不要再操作主服务器MYSQL，防止主服务器状态值变化**

### 配置从机

- CHANGE MASTER TO MASTER_HOST='192.168.124.3',MASTER_USER='zhangsan',MASTER_PASSWORD='123456',MASTER_LOG_FILE='mysqlbin.具体数字',MASTER_LOG_POS=具体值;
- start slave;
- show slave status\G\
    下面两个参数都是Yes，则说明主从配置成功！
      Slave_IO_Running: Yes
      Slave_SQL_Running: Yes
      
    
- stop slave; 停止复制