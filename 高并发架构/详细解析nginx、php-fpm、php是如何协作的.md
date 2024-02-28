## 前言

本文解释一下cgi fastcgi php-cgi php-fpm  nginx  php 之间的关系

## 引出

很久之前的web网站都是处理一些静态的网页，当时直接使用web服务器比如nginx apapche 直接返回静态文件就行了

但是近几年web页面发展的越来越快，导致我们需要越来越多的动态内容来显示，所以需要一个脚本处理器去解析动态内容，然后返回给web服务器，然后再返回给客户端

但是后端语言比如php、java、python、golang等，所以需要一个统一的协议来让web服务器于后端语言进行交互，于是出现了cgi协议

## cgi

当web服务器如nginx接收到客户端的请求比如/index.html，nginx会直接返回给客户端这个文件，但是当收到比如/index.php请求的时候，会启动响应的cgi程序来解析这个index.php文件，然后再以cgi协议规定的格式返回给nginx，nginx再返回给客户端

简单来说就是cgi是一个协议，不是一个程序，cgi是web服务器比如nginx apache 与比如php解析器交互的一种协议，一种规定

cgi程序的工作模式为：
>每当又一个请求过来，实现cgi程序就会fork一个进程去处理这个请求，当请求完成之后又销毁进程

这样的工作模式在高并发下，频繁的创建和销毁进程是一笔很大的开销，所以就出现了fastcgi

## fastcgi

首先说明faastcgi不是一个程序，也是和cgi一样是一种协议，是cgi的升级版，由名字也能看出来，php-fpm就是一个实现了fastcgi协议的进程管理工具

fastcgi工作模式：

- fastcgi程序，会先fork一个master，解析配置文件，初始化执行环境，然后再fork多个worker，~~当一个请求过来，挑选一个空闲的进程来处理请求，请求结束之后，不销毁进程，而是放回原来的进程池中，等待接收其他的请求~~
!> worker是一直在accept()获取cgi的请求，进行处理,不是master挑选空闲的worker去处理
- master通过共享内存获取worker的信息，比如worker进程当前状态、已处理请求数等
- master通过信号来操纵worker，比如杀死worker
- 当空闲的worker太多的时候，会杀死一些worker，反之当worker不够用的时候，会临时增加一定量的worker来处理请求，但不是无限量的增加，会有一个顶峰值

## php-fpm

>php-fpm就是实现来fastcgi协议的进程管理工具

工作内容就是：

1. 等待请求: worker进程阻塞在fcgi_accept_request()等待请求
2. 解析请求： fastcgi请求到达后被worker接收，然后开始接收并解析请求数据，直到request数据完全到
3. 请求初始化： 执行php_request_startup()，此阶段会调用每个扩展的：PHP_RINIT_FUNCTION()
4. 编译、执行： 由php_execute_script()完成PHP脚本的编译、执行(这一步就是我们执行我们写的php代码了)
5. 关闭请求： 请求完成后执行php_request_shutdown()，此阶段会调用每个扩展的：PHP_RSHUTDOWN_FUNCTION()，然后进入步骤1


### php-fpm常用配置

1. pid = /usr/local/var/run/php-fpm.pid   这个文件记录的php-fpm运行时的pid
2. error_log  = /usr/local/var/log/php-fpm.log  错误日志存储的地方
3. log_level = notice   错误日志的级别
4. daemonize = yes  设置后台执行fpm
5. listen = 127.0.0.1:9000   设置fpm监听的端口号
6. user = www  group = www    fpm运行时的用户和组
7. rlimit_files = 1024    设置文件打开描述符的rlimit限制

### 重要配置 

- pm = static | dynamic | ondemand

#### pm = static 模式

表示我们创建的php-fpm子进程数量是固定的，那么就只有pm.max_children = 50这个参数生效。你启动php-fpm的时候就会一起全部启动51(1个主＋50个子)个进程，颇为壮观。

#### pm = dynamic 模式

表示启动进程是动态分配的，随着请求量动态变化的。他由 pm.max_children，pm.start_servers，pm.min_spare_servers，pm.max_spare_servers 这几个参数共同决定。

- pm.max_children ＝ 50 	  是最大可创建的子进程的数量。必须设置。这里表示最多只能50个子进程。

- pm.start_servers = 20    随着php-fpm一起启动时创建的子进程数目。默认值：min_spare_servers + (max_spare_servers - min_spare_servers) / 2。这里表示，一起启动会有20个子进程。

- pm.min_spare_servers = 10    设置服务器空闲时最小php-fpm进程数量。必须设置。如果空闲的时候，会检查如果少于10个，就会启动几个来补上。

- pm.max_spare_servers = 30    设置服务器空闲时最大php-fpm进程数量。必须设置。如果空闲时，会检查进程数，多于30个了，就会关闭几个，达到30个的状态。

#### 到底选择static还数dynamic?

一般原则是：dynamic动态适合小内存机器，灵活分配进程，省内存。static静态适用于大内存机器，动态创建回收进程对服务器资源也是一种消耗。

如果你的内存很大，有8-20G，按照一个php-fpm进程20M算，100个就2G内存了，那就可以开启static模式。如果你的内存很小，比如才256M，那就要小心设置了，因为你的机器里面的其他的进程也算需要占用内存的，所以设置成dynamic是最好的，比如：pm.max_chindren = 8, 占用内存160M左右，而且可以随时变化，对于一半访问量的网站足够了。

### 慢日志查询

我们有时候会经常饱受500,502问题困扰。当nginx收到如上错误码时，可以确定后端php-fpm解析php出了某种问题，比如，执行错误，执行超时。这个时候，我们是可以开启慢日志功能的。

>slowlog = /usr/local/var/log/php-fpm.log.slow

>request_slowlog_timeout = 15s

当一个请求该设置的超时时间15秒后，就会将对应的PHP调用堆栈信息完整写入到慢日志中。

php-fpm慢日志会记录下进程号，脚本名称，具体哪个文件哪行代码的哪个函数执行时间过长：

	1. [21-Nov-2013 14:30:38] [pool www] pid 11877
    2. script_filename = /usr/local/lnmp/nginx/html/www.quancha.cn/www/fyzb.php
    3. [0xb70fb88c] file_get_contents() /usr/local/lnmp/nginx/html/www.quancha.cn/www/fyzb.php:2

通过日志，我们就可以知道第2行的file_get_contents 函数有点问题，这样我们就能追踪问题了。

## nginx 

nginx是在此处作为反向代理服务器，转发请求的，此处就不再介绍nginx来，这个东西可以写十篇文章也写不完，有兴趣的同学可以自己学习一下，博主最近也在看极客时间上陶辉大佬分析的《nginx核心100讲》

我会写一篇关于nginx的博客，有兴趣的同学到时候可以看看

## 流程介绍

这里简单介绍下他们工作的流程

1. 客户端发起请求，此时请求会到nginx这里
2. nginx也是master worker模式，worker会accept请求，这里使用了accept_mutex的东西来解决惊群效应，worker利用了异步I/O多路复用的epoll实现，
3. nginx的worker看请求的是静态文件还是动态内容，可以在nginx配置里面设置转发请求，静态文件就直接返回给客户端了，如果是动态内容，比如index.php，此时，使用socket通信，通过fastcgi协议将请求转发给php-fpm
4. php-fpm的worker进程accept到了请求,就走下面的流程；如何没有空闲的worker，并且配置的最大值 pm.max_children = 50 ，这里是50个worker，就会阻塞等待其他worker处理完
5. worker进程把请求交给php解析器，php解析器进行解析，比如从redis、mysql取出数据等，然后返回给worker
6. worker将请求返回给nginx，然后回到进程池里面等待新请求
7. nginx的worker将结果返回给客户端，然后回到nginx的worker进程池等待新的请求
8. 客户端展示请求结果

## 参考

https://www.zybuluo.com/phper/note/89081
https://www.zhihu.com/question/30672017