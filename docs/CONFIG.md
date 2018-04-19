# Postoffice 配置文件格式
> 所有系统启动参数依赖于配置文件，配置文件采用YAML格式，路径为config/postoffice/config.yaml
----
*样式如下：*

    auth:
      type: mongodb
      url: mongodb://192.168.41.170:30707
    matrix:
      zookeeper:
        host: zookeeper.default
        port: 2181
    mq:
      type: kafka
      host: kafka.default
      port: 9092
    mqtt:
      keepAlive: 300
      connectTimeout: 2
      ackTimeout: 20
      timeoutRetries: 3
    source:
      - type: tcp
        ssl: false
        port: 1883
      - type: websocket
        ssl: false
        port: 8080
    ssl:
      crt: ./cert/postoffice.crt
      key: ./cert/postoffice.key

## 1、auth 认证方式：
用于服务连接设备用户名密码的验证。
### 1.1、type
验证方式配置，设置使用哪种验证方式。mongodb、mysql、redis、mock等几种模式
### 1.2、url
连接存储使用的url地址。

## 2、matrix 服务发现：
当Angler部署到集群中，PostOffice将自动支持其消息转发。当Angler从集群中移除时，PostOffice将自动关闭该对用的消息支持。
### 2.1 zookeeper配置
系统中将应用信息配置在zookeeper中，此处配置zookeeper主机地址及端口号。
（由于系统默认将Zookeeper安装在Kubernetes的default namespace下，因此使用zookeeper.default）
[Zookeeper中Postoffice转发规则结构](https://github.com/IvoryRaptor/InvoryRaptor/blob/master/zookeeper/POSTOFFICE.md)

## 3、mq 消息队列
消息转发的目标消息队列
### 3.1 type 消息队列类型
消息队列类型，kafka、activemq、rabbitmq等
### 3.2 host及port 连接消息队列的地址
目标消息队列的地址及端口

## 4、mqtt MQTT连接配置
mqtt服务连接信息，保持会话时间、连接超时等信息

## 5、source 通道配置
服务提供的MQTT端口信息，此处为数组。支持多端口接入。每个端口有自己独立的配置。
### 5.1 type
通道类型，目前MQTT支持的通道类型有两种分别是直接TCP通讯，及基于WebSocket的通讯。因此该参数可以为tcp或weboskcet。
### 5.2 ssl
是否使用ssl加密通道。TCP客户端连接时使用ssl，WebSocket连接时使用wss。此处仅设置是否使用加密通讯。
### 5.3 port
端口设置，设置监听端口。使用时注意打开防火墙该端口设置。

## 6、ssl SSL秘钥配置
设置ssl加密时所使用的公钥及私钥。
