#设备连接模式分为三种，与验证模式相关：
>> 支持验证模块包括，Redis、mongoDB验证模式

mqttClientId: clientId+"|securemode=3,signmethod=hmacsha1,timestamp=132323232|"
mqttUsername: deviceName+"&"+productKey
mqttPassword: sign_hmac(deviceSecret,content)

### 1.1、clientId
表示客户端id，建议mac或sn，64字符内。

### 1.2、timestamp
表示当前时间毫秒值，可以不传递。

### 1.3、mqttClientId
格式中||内为扩展参数。

### 1.4、securemode
表示目前安全模式，可选值有：

#### 1.4.1  2-TLS直连模式
支持验证模块包括，Redis、mongoDB验证模式
#### 1.4.2  3-TCP直连模式

#### 1.4.3  98-TLS OAuth 模式
OAuth模式，为PostOffice向第三方认证。如果使用这种模式，signmethod需指定为none，密码区域填入token
deviceName为空。

#### 1.4.4  99-TLS Token 模式
主要应用于Web场景，避免重新登录时每次输入密码，使用该模式时，deviceName部分为token

### 1.5、signmethod
表示签名算法类型。
hmacsha1    HmacSHA1加密方法
hmacmd5     macmd5加密方法
none        明文传输不加密
