# Postoffice
Postoffice项目介绍

>为物联网云平台提供长连接通道服务，目前项目仅支持MQTT协议。提供两种MQTT协议接入方式，分别为TCP、WebSocket。

该项目采用golang设计，可同时配置多个通道，
## 1、功能概述：

Postoffice通过配置文件启动，

## 2、支持功能
### 2.1、多通道
可同时开启任意多个端口为移动端、网页端、设备端提供MQTT协议接入服务。每个端口仅支持一种接入类型，TCP接入或WebSocket接入

### 2.2、SSL支持
每个端口可配置是否使用SSL进行加密通讯，但整个系统中仅使用一套公钥及私钥。

### 2.3、集群运行
支持集群运行，整个系统部署在kubernetes中，PostOffice可多副本运行，支持集群模式。连入设备被集群中某个应用Pod所消费。

### 2.4、服务发现
当Angler部署到集群中，PostOffice将自动支持其消息转发。当Angler从集群中移除时，PostOffice将自动关闭该对用的消息支持。

## 3、配置文件格式
[配置文件格式](https://github.com/IvoryRaptor/postoffice/tree/master/docs/CONFIG.md)
