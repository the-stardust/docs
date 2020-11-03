## 单进程的生命周期

CLI/CGI模式的PHP属于单进程的SAPI模式。这类的请求在处理一次请求后就关闭。也就是只会经过如下几个环节： 开始 - 请求开始 - 请求关闭 - 结束 SAPI接口实现就完成了其生命周期。如图所示：

![pic](http://blogs.xinghe.host/images/pasted-141.png)


### 启动

在调用每个模块的模块初始化之前，会有一个初始化的过程：

- 初始化若干全局变量
		这里的初始化大多数是将其设置为NULL
- 初始化若干常量
		这里的常量是PHP自己的常量比如：PHP_VERSION PEAR_EXTENSION_DIR等，是卸载congig.w32.h里面的
- 初始化Zend引擎和核心组件
		zend_startup函数就是初始化Zend引擎，这里的初始化操作包括内存管理初始化、全局使用的函数指针初始化、，对PHP源文件进行词法分析、语法分析、中间代码执行的函数指针的赋值，初始化若干hashTable(比如函数表、常量表)，为ini文件解析做准备、为PHP源文件解析做准备，注册内置函数(strlen、define等)，注册标准常量(E_ALL、TRUE、FALSE等)、注册GLOBALS全局变量
- 解析php.ini
- 全局操作函数初始化
		$_GET $_POST $_FILES等  这里的初始化只是把这些变量名添加到CG(auto_globals)这个变量表中 
- 初始化静态构建的模块和共享模块(MINIT)
		模块初始化会执行两个操作:
        	1.将这些模块注册到已注册模块列表
        	2.将每个模块中包含的函数注册到函数表(CG(function_table))
        注册了静态构建的模块后，PHP会注册附加模块，CLI模式下是没有附加模块的
        在内置模块和附加模块后，接下来是注册通过共享对象(比如DLL)和php.ini文件灵活配置的扩展
        在所有模块都注册后，PHP会马上遍历每个模块，执行每个模块的初始化函数，就是PHP_MINIT_FUNCTION包含的内容
- 禁用函数和类
		php_disable_functions 禁用一些PHP的函数 就是将这些函数从CG(function_table)里面删除 php_disable_classes 就是从CG(class_table)类表中删除。

### ACTIVATION

在处理了文件相关的内容，PHP会调用php_request_startup做请求初始化操作，他处理调用每个模块的初始化函数外，还做了其他的工作：

- 激活Zend引擎
		gc_reset函数用来重置垃圾收集机制
        init_compiler函数用来初始化编译器
        init_executor函数用来初始化中间代码执行过程
        通过php.ini配置的zend_extensions也是在这里被遍历调用activate函数。
- 激活SAPI
		如果当前模式下有设置activate函数，则运行此函数，激活SAPI，在CLI模式下此函数指针被设置为NULL。        
- 环境初始化
		这里的环境初始化指的是用户需要用到的一些环境变量初始化，包括服务器环境、请求数据环境等、实际用到的变量就是$_POST $_GET $_COOKIE、$_SERVER、$_ENV、$_FILES  
        以$_COOKIE为例，php_default_treat_data函数会对依据分隔符，将所有的cookie拆分并赋值给对应的变量
- 模块请求初始化
		PHP通过zend_activate_modules函数实现模块的请求初始化，也就是图中的 Call each extension's RINIT 遍历module_registry变量中的所有模块，调用其RINIT方法实现模块请求的初始化

### 运行

1.php_execute_script函数包含了运行PHP脚本的全部过程。

2.当一个PHP文件需要解析执行时，它可能会需要执行三个文件，其中包括一个前置执行文件、当前需要执行的主文件和一个后置执行文件

3.对于需要解析执行的文件，通过zend_compile_file（compile_file函数）做词法分析、语法分析和中间代码生成操作，返回此文件的所有中间代码。 如果解析的文件有生成有效的中间代码，则调用zend_execute（execute函数）执行中间代码。 如果在执行过程中出现异常并且用户有定义对这些异常的处理，则调用这些异常处理函数。 在所有的操作都处理完后，PHP通过EG(return_value_ptr_ptr)返回结果

### DEACTIVATION

PHP关闭请求的过程是一个若干个关闭操作的集合，这个集合存在于php_request_shutdown函数中。 这个集合包括如下内容：

1. 调用所有通过register_shutdown_function()注册的函数，这些关闭时调用的函数是在用户空间添加进来的，一个简单的例子，我们可以在脚本出错时调用一个统一的函数，给用户一个友好一些的页面，这个有点类似于网页中的404页面。
2. 执行所有的__destruct函数
3. 将所有的输出刷出去
4. 发送http应答头，这也是一个输出字符串的过程
5. 遍历每个模块的关闭请求方法，执行模块的请求关闭操作，这就是我们在图中看到的Call each extension's RSHUTDOWN。
6. 销毁全局变量表(PG(http_globals))的变量
7. 通过zend_deactivate函数，关闭词法分析器、语法分析器和中间代码执行器
8. 调用每个扩展的post-RSHUTDOWN函数
9. 关闭SAPI，通过sapi_deactivate销毁SG(sapi_headers)、SG(request_info)等的内容。
10. 关闭流的包装器、关闭流的过滤器
11. 关闭内存管理
12. 重新设置最大执行时间

### 结束

- flush
		sapi_flush将最后的内容刷新出去。其调用的是sapi_module.flush，在CLI模式下等价于fflush函数。
- 关闭Zend引擎
		zend_shutdown将关闭Zend引擎
        此时对应图中的流程，我们应该是执行每个模块的关闭模块操作
        关闭所有的模块后，PHP继续销毁全局函数表、全局类表、全局变量表等，遍历每个扩展的shutdown函数
        
## 多进程SAPI生命周期

通常PHP是编译为apache的一个模块来处理PHP请求，apache会fork出多个子进程，每个进程的内存空间独立，每个进程会经过开始和结束环节，不过每个进程的开始阶段只在进程fork出来以后执行，在整个进程的生命周期内可能处理多个请求，只有在apache关闭或者进程结束之后才会进行关闭阶段，在这两个阶段之间会随着每个请求的重复请求开始-请求关闭的环节，如图所示


![upload successful](http://blogs.xinghe.host/images/pasted-142.png)

### 多线程的生命周期

多线程模式和多进程中的某个进程类似，不同的是在整个进程的生命周期内会并行的重复着 请求开始-请求关闭的环节

![upload successful](http://blogs.xinghe.host/images/pasted-143.png)


### Zend引擎

Zend引擎是PHP实现的核心，提供了语言实现上的基础设施 例如：PHP的语法实现，脚本的编译运行环境，扩展机制预计内存管理

目前PHP的实现和Zend引擎之间的关系非常紧密，甚至有些过于紧密了，例如很多PHP扩展都是使用的Zend API， 而Zend正是PHP语言本身的实现，PHP只是使用Zend这个内核来构建PHP语言的，而PHP扩展大都使用Zend API， 这就导致PHP的很多扩展和Zend引擎耦合在一起了

PHP中的扩展通常是通过Pear库或者原生扩展