# 模块及代码结构
> 配置文件与目录对应关系：
系统中每个模块对应一个目录
每个目录中的config.go文件，存储该模块使用的配置格式
整体配置文件格式由，kernel/config.go文件进行串联

----
**模块约定规则**
* 每个模块被封装在一个单独的目录中
* 模块Config的结构定义在自己该目录的config.go文件中，并由kernel模块的config进行拼装，成为最终配置文件读取时使用的格式。

    type Config struct {
        Url string
    }

* 模块

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
