## 概述
1. 操作系统直接管理着内存，所以操作系统也需要内存管理，计算机中通常都有内存管理单元(MMU)用于处理cpu对内存的访问
2. 应用程序无法直接调用物理内存，只能向OS申请，向OS申请内存，会引发系统调用，系统调用会把cpu从用户态切换到内核态
3. 为了减少系统调用开销,通常在用户态就对内存进行管理，使用完内存不立即返回给OS，而是复用，避免每次内存释放和申请带来开销
4. PHP不需要显示的内存管理，由Zend引擎来管理
5. PHP内存限制
    - php.ini中的默认32MB
            memory_limit = 32M
    - 动态修改内存
            ini_set ("memory_limit", "128M")
    - 获取目前内存占用
            memory_get_usage() : 获取PHP脚本所用的内存大小
            memory_get_peak_usage() ：返回当前脚本到目前位置所占用的内存峰值。


    
## PHP的内存管理
![pic](../images/263175-116fa0f4acf9111a.png)
### 接口层
是一些宏定义

### 堆层
_zend_mm_heap

初始化内存，调用_zend_mm_startup,PHP内存管理维护三个列表
- 小块内存列表 free_buckets
- 大块内存列表 large_free_buckets
- 剩余内存列表 rest_buckets

### 存储层
- 内存分配的方式对堆层透明化，实现存储层和堆层的分离
- 不同的内存分配方案，有对应的处理函数

### 内存的申请
PHP对于内存的申请，围绕着小块内存列表(free_buckets)、大块内存列表(large_free_buckets)、剩余内存列表(rest_buckets) 三个列表来分层进行的

ZendMM向OS进行内存申请，首先ZendMM的最底层heap层先向OS申请一块大的内存，通过对上述三个列表的填充，建立一个类似内存池的管理机制

在程序需要内存的时候，ZendMM会在内存池中分配相应的内存供程序使用，这样的好处是避免了PHP频繁的向OS申请内存和释放内存


### ZendMM对内存分配的处理步骤

1. 内存检查
2. 命中缓存，找到内存块，调至步骤5
3. 没有命中缓存，在ZendMM管理的heap层存储中搜索大小适合的内存块，是在三个列表中小到大进行的，找到block后，调到步骤5
4. 步骤3没有找到内存，则使用ZEND_MM_STORAGE_ALLOC申请新内存块(至少为ZEND_MM_SEG_SIZE)，进行步骤6
5. 使用zend_mm_remove_from_free_list函数将依据使用block节点在zend_mm_free_block中移除
6. 内存分配完毕，对zend_mm_heap结构中的各种标识型变量进行维护，包括large_free_buckets、peak、size等
7. 返回分配的内存地址

![](../images/263175-6e8b09aa1b85d27b.png)

## 内存的销毁

ZendMM在内存销毁的处理上采用与内存申请相同的策略，当程序unset一个变量或者其他的释放行为的时候，ZendMM不会立刻将
内存交回给系统，而是只在自身维护的内存池中将其标志成可用，按照内存大小整理到上述三个列表(large、free、small)之中，
以备下次内存申请时使用

ZendMM将内存块以整理收回到zend_mm_heap的方式回收到内存池中

程序使用的所有内存，将在进程结束的时候，统一交回给OS
