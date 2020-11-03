- daemonize yes redis默认不是守护进程no，修改为守护进程;

- pidfile /var/run/redis.pid 当redis以守护进程运行时，redis会默认把pid写进/var/run/redis.pid文件，如需修改配置可以指定

- redis端口：port 6379

- bind 127.0.0.1 绑定的主机地址，docker容器要使用docker inspect 查看IPAddress 加入绑定

- timeout 300 当客户端闲置时间后关闭 0为关闭该功能

- loglevel verbose 指定日志级别 redis有四个 debug verbose notice warning 默认为loglevel verbose 生产环境推荐notice

- logfile stdout 日志记录方式，默认为标准输出，如果配置redis为守护进程，并且为标准输出，日志会记录在/dev/null

- database 16 设置redis的数据库个数 默认16 用select <下标>切换

- save <seconds> <changes> 表示redis在多长时间内有多少次更新操作，就将数据同步到数据文件。可以多个条件配合使用 Redis默认配置文件中提供了三个条件：

        save 900 1
        
        save 300 10
        
        save 60 10000

        分别表示900秒内有一个更改，300秒内有10次更改以及60秒内有10000次更改

- rdbcompression yes 指定存储值本地数据库时是否压缩数据，默认为yes，redis采用LZF压缩，如果为了节省cpu时间，可以关闭，但会导致数据库文件变的巨大

- dbfilename dump.rdb 指定本地数据库文件名

-  dir ./ 指定本地数据库存放目录 默认当前目录

- slaveof <materip><masterport> 当设置本机redis为slave的时候，设置主机master的ip和port 当redis启动时，自动同步数据

- masterauth <master-password> 当master设置密码的时候，配置slave服务连接密码

- requirepass foobared 设置redis密码的时候，如果配置了连接密码，客户端在连接的时候，必须输入 auth <password> 才能进行操作

- maxclients 128 设置默认客户端最大连接数 0为不做限制 当客户端连接数到达限制时，Redis会关闭新的连接并向客户端返回max number of clients reached错误信息

- maxmomory <bytes> 指定Redis最大内存限制，Redis在启动时会把数据加载到内存中，达到最大内存后，Redis会先尝试清除已到期或即将到期的Key，当此方法处理 后，仍然到达最大内存设置，将无法再进行写入操作，但仍然可以进行读取操作。Redis新的vm机制，会把Key存放内存，Value会存放在swap区

-  appendonly no 指定是否在每次更新操作后进行日志记录，reids默认为异步吧数据同步到磁盘中，如果为no，可能断电后会导致部分数据丢失，因为save配置的那一段时间，数据实在内存中

- appendfilename appendonly.aof制定更新日志文件名

- appendfsync everysec 指定日更新日志条件，3个选项

        no：表示等操作系统进行数据缓存同步到磁盘（快）
        
        always：表示每次更新后手动调用fsync()函数进行同步（慢，安全）
        
        everysec：表示每秒同步（折中，默认值）

- vm-enable no 指定是否启用虚拟机内存机制，默认no，VM机制将数据分页存放，由redis将冷数据也就是不经常访问的数据，swap到磁盘中去，将热数据也就是经常访问的数据放在内存中

- vm-swap-file /tmp/redis.swap 虚拟内存文件路径不可多个redis实例共享

- vm-max-memory 0 将所有大于vm-max-memory的数据存入内存中去，无论vm-max-memory设置多小，所有索引数据都是放在内存存储的（redis的索引数据，也就是keys），也就是说，当vm-max-memory为0时，所有value都在磁盘中 默认0；

- vm-page-size 32bytes reids把swap分为很多page，一个page上的数据不能被多个对象共享，vm-page-size是要根据存储的数据大小决定的，如果是存储很多个小对象，可以设置为32或者64bytes 如果是很多大对象，可以增加page

- 设置swap中page的数量，由于页表（一种表示页面空闲或使用的bitmap）存在内存中的，在磁盘中8个pages将消耗1byte的内存

- vm-max-threads 4 默认值4，设置swap文件的线程数量，如果设置为0，则表示串行化，执行时间会变长，设置的值最好不要超过机器核数

- glueoutputbuf yes 这只客户端在应答时，是否把较小的包合并为一个包发送，默认开启

- hash-map-zipmap-entries 64 hash-map-zipmap-value 512 指定超过一定数量或者最大元素超过某个临界值，采用一个特殊的哈希算法

- activerehashing yes 指定会否激活重置哈希， 默认开启，

- include /path/to/local.conf hiding包含其他的配置文件，可以在同一主机行多个redis实例之间会用同一份配置文件，而同时各个实例之间又拥有自己的特定配置

- auto-aof-rewrite-percentage 100 auto-aof-rewrite-min-size 64mb 默认值 大生产环境会设置到G级别，aof文件大小超过这个值，会触发重写机制
