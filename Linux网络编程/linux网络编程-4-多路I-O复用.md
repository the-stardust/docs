## 前言
多路IO转接服务器也叫做多任务IO服务器。该类服务器实现的主旨思想是，不再由应用程序自己监视客户端连接，取而代之由内核替应用程序监视文件。

主要使用的方法有三种：select()函数 poll()函数 epoll()函数

## select
- 1.select能监听的文件描述符个数受限于FD_SETSIZE,一般为1024，单纯改变进程打开的文件描述符个数并不能改变select监听文件个数（需要重新编译linux内核生效）

- 2.解决1024以下客户端时使用select是很合适的，但如果链接客户端过多，**select采用的是轮询模型**，会大大降低服务器响应效率，不应在select上投入更多精力


      #include <sys/select.h>
      /* According to earlier standards */
      #include <sys/time.h>
      #include <sys/types.h>
      #include <unistd.h>
      int select(int nfds, fd_set *readfds, fd_set *writefds,
	  fd_set *exceptfds, struct timeval *timeout);
          nfds: 		监控的文件描述符集里最大文件描述符加1，因为此参数会告诉内核检测前多少个文件描述符的状态
          readfds：	监控有读数据到达文件描述符集合，传入传出参数
          writefds：	监控写数据到达文件描述符集合，传入传出参数
          exceptfds：	监控异常发生达文件描述符集合,如带外数据到达异常，传入传出参数
          timeout：	定时阻塞监控时间，3种情况
                      1.NULL，永远等下去
                      2.设置timeval，等待固定时间
                      3.设置timeval里时间均为0，检查描述字后立即返回，轮询
          struct timeval {
              long tv_sec; /* seconds */
              long tv_usec; /* microseconds */
          };
          void FD_CLR(int fd, fd_set *set); 	//把文件描述符集合里fd清0
          int FD_ISSET(int fd, fd_set *set); 	//测试文件描述符集合里fd是否置1
          void FD_SET(int fd, fd_set *set); 	//把文件描述符集合里fd位置1
          void FD_ZERO(fd_set *set); 			//把文件描述符集合里所有位清0
          
>select函数的逻辑就是，服务端使用select扮演一个代理的角色，去监听最多1024个socket的文件描述符，当哪个socket文件描述符有事件发生，select使用轮询的方式去看看到底是哪个socket发生了事件，然后去交给服务端处理，

### 缺点

- 1.最多1024个socket
- 2.当有事件发生，在文件描述符表中，比如监听的句柄是3，4，1020，这三个，但是selecvt就要循环1021次，去看看是哪个socket发生事件，所以太麻烦了
- 3.select的参数是传入传出参数，传入的参数会被覆盖，所有要自己声明一个数组去存储

## poll
poll函数就是select函数的升级版，并没有解决循环的问题，只是把参数的传入传出属性给调整了，不用自己去声明数组了

    include <poll.h>
    int poll(struct pollfd *fds, nfds_t nfds, int timeout);
        struct pollfd {
            int fd; /* 文件描述符 */
            short events; /* 监控的事件 */
            short revents; /* 监控事件中满足条件返回的事件 */
        };
        POLLIN			普通或带外优先数据可读,即POLLRDNORM | POLLRDBAND
        POLLRDNORM		数据可读
        POLLRDBAND		优先级带数据可读
        POLLPRI 		高优先级可读数据
        POLLOUT		普通或带外数据可写
        POLLWRNORM		数据可写
        POLLWRBAND		优先级带数据可写
        POLLERR 		发生错误
        POLLHUP 		发生挂起
        POLLNVAL 		描述字不是一个打开的文件

        nfds 			监控数组中有多少文件描述符需要被监控

        timeout 		毫秒级等待
            -1：阻塞等，#define INFTIM -1 				Linux中没有定义此宏
            0：立即返回，不阻塞进程
            >0：等待指定毫秒数，如当前系统时间精度不够毫秒，向上取值



如果不再监控某个文件描述符时，可以把pollfd中，fd设置为-1，poll不再监控此pollfd，下次返回时，把revents设置为0。

select()服务端代码实现：

      int main(int argc, char *argv[])
      {
          int i, maxi, maxfd, listenfd, connfd, sockfd;
          int nready, client[FD_SETSIZE]; 	/* FD_SETSIZE 默认为 1024 */
          ssize_t n;
          fd_set rset, allset;
          char buf[MAXLINE];
          char str[INET_ADDRSTRLEN]; 			/* #define INET_ADDRSTRLEN 16 */
          socklen_t cliaddr_len;
          struct sockaddr_in cliaddr, servaddr;

          listenfd = Socket(AF_INET, SOCK_STREAM, 0);

          bzero(&servaddr, sizeof(servaddr));
          servaddr.sin_family = AF_INET;
          servaddr.sin_addr.s_addr = htonl(INADDR_ANY);
          servaddr.sin_port = htons(SERV_PORT);

          Bind(listenfd, (struct sockaddr *)&servaddr, sizeof(servaddr));

          Listen(listenfd, 20); 		/* 默认最大128 */

          maxfd = listenfd; 			/* 初始化 */
          maxi = -1;					/* client[]的下标 */

          for (i = 0; i < FD_SETSIZE; i++)
              client[i] = -1; 		/* 用-1初始化client[] 这个数组只存需要监听的文件描述符*/

          FD_ZERO(&allset); /*清空set*/
          FD_SET(listenfd, &allset); /* 构造select监控文件描述符集 */

          for ( ; ; ) {
              rset = allset; 			/* 每次循环时都从新设置select监控信号集 rset是每次需要监听的集合，由于是传入传出参数或覆盖，所有需要allset辅助记录*/
              nready = select(maxfd+1, &rset, NULL, NULL, NULL);
			/*nready返回的是所有需要监听的socket文件描述符中，所符合要求的socket文件描述符的总数*/
              if (nready < 0)
                  perr_exit("select error");
              if (FD_ISSET(listenfd, &rset)) { /* new client connection */
                  cliaddr_len = sizeof(cliaddr);
                  connfd = Accept(listenfd, (struct sockaddr *)&cliaddr, &cliaddr_len);
                  printf("received from %s at PORT %d\n",
                          inet_ntop(AF_INET, &cliaddr.sin_addr, str, sizeof(str)),
                          ntohs(cliaddr.sin_port));
                          /*输出哪个客户端ip和端口*/
                  for (i = 0; i < FD_SETSIZE; i++) {
                      if (client[i] < 0) {
                          client[i] = connfd; /* 保存accept返回的文件描述符到client[]里 */
                          break;
                      }
                  }
                  /* 达到select能监控的文件个数上限 1024 */
                  if (i == FD_SETSIZE) {
                      fputs("too many clients\n", stderr);
                      exit(1);
                  }

                  FD_SET(connfd, &allset); 	/* 添加一个新的文件描述符到监控信号集里 */
                  if (connfd > maxfd)
                      maxfd = connfd; 		/* select第一个参数需要 */
                  if (i > maxi)
                      maxi = i; 				/* 更新client[]最大下标值 */

                  if (--nready == 0)
                      continue; 				/* 如果没有更多的就绪文件描述符继续回到上面select阻塞监听,负责处理未处理完的就绪文件描述符 */
              		}
                  for (i = 0; i <= maxi; i++) { 	/* 检测哪个clients 有数据写入到socket中 */
                      if ( (sockfd = client[i]) < 0)
                          continue;
                      if (FD_ISSET(sockfd, &rset)) {
                          if ( (n = Read(sockfd, buf, MAXLINE)) == 0) {
                              Close(sockfd);		/* 当client关闭链接时，服务器端也关闭对应链接 */
                              FD_CLR(sockfd, &allset); /* 解除select监控此文件描述符 */
                              client[i] = -1;
                          } else {
                              int j;
                              for (j = 0; j < n; j++)
                                  buf[j] = toupper(buf[j]);
                              Write(sockfd, buf, n);
                          }
                          if (--nready == 0)
                              break;
                      }
                  }
              }
              close(listenfd);
              return 0;
      }
      
## epoll
epoll是Linux下多路复用IO接口select/poll的增强版本，
**它能显著提高程序在大量并发连接中只有少量活跃的情况下的系统CPU利用率**

- 因为它会复用文件描述符集合来传递结果而不用迫使开发者每次等待事件之前都必须重新准备要被侦听的文件描述符集合，
- 另一点原因就是获取事件的时候，它无须遍历整个被侦听的描述符集，只要遍历那些被内核IO事件异步唤醒而加入Ready队列的描述符集合就行了。

!>目前epell是linux大规模并发网络程序中的热门首选模型**

epoll 流程为：

    有监听fd事件发送--->返回监听满足数组--->判断返回数组元素--->
    lfd满足accept--->返回cfd---->read()读数据--->write()给客户端回应。


epoll除了提供select/poll那种IO事件的**水平触发（Level Triggered）**外，还提供了**边沿触发（Edge Triggered）**，这就使得用户空间程序有可能缓存IO状态，减少epoll_wait/epoll_pwait的调用，提高应用程序效率


### 基础 API

- int epoll_create(int size)		size：监听数目（是一个建议值）
- int epoll_ctl(int epfd, int op, int fd, struct epoll_event *event)
  - epfd：	为epoll_creat的句柄
  - op：表示动作，用3个宏来表示：
		EPOLL_CTL_ADD (注册新的fd到epfd)，
		EPOLL_CTL_MOD (修改已经注册的fd的监听事件)，
		EPOLL_CTL_DEL (从epfd删除一个fd)；
  - event：告诉内核需要监听的事件
  
  	  epoll_event.data类型：
  	  struct epoll_event {
			epoll_data_t data; /* User data variable */
            __uint32_t events; /* Epoll events */
		};
		typedef union epoll_data {
			void *ptr; /*泛型指针，epoll反应堆的时候会重点研究*/
			int fd; /*与epoll_ctl函数里面第三个参数一致就行*/
			uint32_t u32;
			uint64_t u64;
		} epoll_data_t;
        epoll_event.events可选参数：
		EPOLLIN ：	表示对应的文件描述符可以读（包括对端SOCKET正常关闭）
		EPOLLOUT：	表示对应的文件描述符可以写
		EPOLLPRI：	表示对应的文件描述符有紧急的数据可读（这里应该表示有带外数据到来）
		EPOLLERR：	表示对应的文件描述符发生错误
		EPOLLHUP：	表示对应的文件描述符被挂断；
		EPOLLET： 	将EPOLL设为边缘触发(Edge Triggered)模式，这是相对于水平触发(Level Triggered)而言的
		EPOLLONESHOT：只监听一次事件，当监听完这次事件之后，如果还需要继续监听这个socket的话，需要再次把这个socket加入到EPOLL队列里
        
- int epoll_wait(int epfd, struct epoll_event *events, int maxevents, int timeout) 类似select函数，等待所监控文件描述符上有事件的产生
	
    	events：		用来存内核得到事件的集合，
		maxevents：	告之内核这个events有多大，这个maxevents的值不能大于创建epoll_create()时的size，
		timeout：	是超时时间
			-1：	阻塞
			0：	立即返回，非阻塞
			>0：	指定毫秒
		返回值：	成功返回有多少文件描述符就绪，时间到时返回0，出错返回-
 



## epoll进阶

- epoll_create函数创建的其实是一个红黑树，返回的句柄是socket文件描述符，在socket文件描述符表中，取出来是一个指针，指针指向一个红黑树的树根，
- 使用epoll_ctl就是往红黑树上添加节点，监听socket句柄，或者删除节点，
- epoll_wait函数返回红黑树上的事件，相比select和poll，是事件通知模型函数，有事件响应返回对应的socket句柄，不用再遍历所有的socket，在监听大量socket却只有少量socket活跃的情况下，有很大的性能提升

### ET/LT模式

也就是边沿触发与水平触发，首先，现在基本上我们所使用的模型都是**非阻塞I/O的边沿触发**模型

![upload successful](../images/pasted-43.png)

使用电路上的知识，上升沿与下降沿与边沿触发相似，水平沿与水平触发类似

简单点讲就是，当客户端发送数据到socket中1000k的数据，这时候socket触发了读事件，epoll_wait返回socket句柄，然后这是服务端从**缓冲区**读取数据,但是**服务端只能读取500k的数据**，此时还有500k的数据还存留在缓冲区，但是服务端无法继续读取了

**因为只有触发事件后，服务端才会调用read函数读取缓冲区的数据**，所有只能等到下次客户端发送数据，服务端才能读取，但是第二次读取到的数据是**上一次遗留的500k**，此时，缓冲区会留下1000k的数据，循环产生这种事件的话，会造成**缓冲区爆满**，客户端无法发送数据，服务端也没有事件通知读取数据

所以epoll工作在ET模式的时候，必须使用**非阻塞套接口**，以避免由于一个文件句柄的阻塞读/阻塞写操作把处理多个文件描述符的任务饿死。最好以下面的方式调用ET模式的epoll接口，在后面会介绍避免可能的缺陷。

- 基于非阻塞文件句柄， fcntl 函数可以将一个socket 句柄设置成非阻塞模式: 
	flags = fcntl(sockfd, F_GETFL, 0);  //获取文件的flags值。
    - fcntl(sockfd, F_SETFL, flags | O_NONBLOCK);   //设置成非阻塞模式；
    - flags  = fcntl(sockfd,F_GETFL,0);
    - fcntl(sockfd,F_SETFL,flags&~O_NONBLOCK);    //设置成阻塞模式；
    - 并在接收和发送数据时，将recv, send 函数的最后有一个flag 参数设置成MSG_DONTWAIT
    	- recv(sockfd, buff, buff_size,MSG_DONTWAIT);     //非阻塞模式的消息发送
    	- send(scokfd, buff, buff_size, MSG_DONTWAIT);   //非阻塞模式的消息接受
- 只有当read或者write返回EAGAIN(非阻塞读，暂时无数据)时才需要挂起、等待。
	- 这并不是说每次read时都需要循环读，直到读到产生一个EAGAIN才认为此次事件处理完成，当read返回的读到的数据长度小于请求的数据长度时，就可以确定此时缓冲中已没有数据了，也就可以认为此事读事件已处理完成
    



### LT模式即Level Triggered工作模式。

与ET模式不同的是，以LT方式调用epoll接口的时候，它就相当于一个速度比较快的poll，无论后面的数据是否被使用。

LT(level triggered)：LT是缺省的工作方式，并且同时支持block和no-block socket。
- 在这种做法中，内核告诉你一个文件描述符是否就绪了，然后你可以对这个就绪的fd进行IO操作。
- 如果你不作任何操作，内核还是会继续通知你的，所以，这种模式编程出错误可能性要小一点。传统的select/poll都是这种模型的代表。


### ET(edge-triggered)：ET是高速工作方式

只支持no-block socket。在这种模式下，当描述符从未就绪变为就绪时，内核通过epoll告诉你。然后它会假设你知道文件描述符已经就绪，并且不会再为那个文件描述符发送更多的就绪通知。请注意，**如果一直不对这个fd作IO操作(从而导致它再次变成未就绪)，内核不会发送更多的通知(only once).**      