配置文件与目录对应关系：
系统中每个模块对应一个目录
每个目录中的config.go文件，存储该模块使用的配置格式
整体配置文件格式由，kernel/config.go文件进行串联

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
