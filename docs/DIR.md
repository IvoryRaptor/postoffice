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


## 2、主程序结构：
### 2.1 main
启动模块，通过配置文件，加载运行各模块，并开始运行。

### 2.1 kernel
模块在/kernel目录下
kernel实现IKernel中的各接口函数，便于各模块调用kernel功能。


#### 2.1.1 运行状态函数


    IsRun() bool
返回是否正在运行

#### 2.1.2 获取副本编号


    GetHost() int32
返回该PostOffice副本运行的编号。
**注意：此处运行时使用kubernetes的StatefulSets来创建Pod，
因此Pod中HostName为postoffice-X的格式，X为host的编号**

### 2.1.3 获取转发Topic数组
接收到MQTT的Publish消息后，通过此函数查找该消息对应转发到消息队列的topic列表。


    GetTopics(matrix string, action string) ([]string, bool)

* matrix matrix名称，每种设备一个名称
* action 动作名称，格式为resource.action
返回是否找到，以及topic的数组。


### 2.1.4 添加通道
将Source接收到的链接，添加进kernel。


    AddChannel(c net.Conn) (err error)
* c 需要网络链接
返回错误

### 2.1.5 验证连接请求
验证MQTT的ConnectMessage消息中的clientId、Username、Password信息是否合法。


    Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig
* msg 连接消息
返回验证后通道信息

### 2.1.6 发送消息
将消息发送到对应的消息队列。


    Publish(topic string,payload []byte) error
* topic 消息队列的Topic
* payload 发送数据域
返回错误


### 2.1.7 等待停止
阻塞等待停止信号。


    WaitStop()
    

## 3、主要模块说明：

### 3.1、auth
认证模块
### 3.2、matrix
配置路径模块
### 3.3、mq
对应kafka或其他消息队列模块
### 3.4、mqtt
 MQTT队列配置
### 3.5、source
通道来源配置

### 3.6、ssl
秘钥等内容配置
