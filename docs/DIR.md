# 模块及代码结构
> 配置文件与目录对应关系：
系统中每个模块对应一个目录
每个目录中的config.go文件，存储该模块使用的配置格式
整体配置文件格式由，kernel/config.go文件进行串联


## 1、模块约定规则
### 1.1、独立目录
每个模块被封装在一个单独的目录中
### 1.2、模块配置
模块Config的结构定义在自己该目录的config.go文件中，并由kernel模块的config进行拼装，成为最终配置文件读取时使用的格式。


    type Config struct {
        Url string
    }

### 1.3、模块需要实现以下几个函数
模块启动及调用顺序为，首先通过Config函数对进行模块进行配置。然后通过Start函数启动模块。
当系统停止时，调用Stop函数停止模块运行。

#### 1.3.1、配置模块函数： 


    Config(kernel postoffice.IKernel,config *Config) error
返回错误
* kernel 内核模块
* config 模块的配置结构


#### 1.3.2、启动模块函数：
**注意：该函数应为非阻塞函数，如果当中有go routine来运行阻塞函数**


    Start() error
返回错误


#### 1.3.3、停止模块函数：


    Stop() error
返回错误


目前包含以下几个模块：

    auth：
        认证模块
    matrix:
        配置路径模块
    mq:
        对应kafka或其他消息队列模块
    mqtt:
        MQTT队列配置
    source:
        通道来源配置
    ssl:
        秘钥等内容配置
