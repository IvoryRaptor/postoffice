# Postoffice 配置文件格式
> 所有系统启动参数依赖于配置文件，配置文件采用YAML格式，路径为config/postoffice/config.yaml

`
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
`