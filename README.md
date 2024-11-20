## 原理
申请SSL，通过API设置DNS TXT记录校验域名所有权后获得SSL证书。

## 支持的DNS供应商
- 阿里云DNS
- 腾讯云DNSPOD

## 案例
- 1、自动申请test.com和itgod.org的SSL通配符证书，保存在./cert_storage目录中，证书到期前10天重新申请证书覆盖即将过期证书
- 2、test.com证书更新后，需要调用/opt/update.sh这个外部脚本执行一些外部的自定义逻辑，比如替换nginx的cert和CDN证书等

### 案例配置文件
```
global:
  certStoragePath: ./cert_storage  # 证书保存路径
providers:
  # 阿里云DNS配置示例
  - providerName: aliDNS  # DNS提供商名称，支持aliDNS|tencentCloud
  enable: true        # 是否启用该证书托管,默认为true启用
  domains: # 只能写一个子域，多个子域需拆分多个provider， 例外： *.xx.com和二级主域可以同时颁发
    - "test.com"   # 设置通配符域名的ssl和二级主域
    - "*.test.com"
  saveSSLName:         # 默认留空，指定SSL证书保存的名字，留空则默认使用domains[0].key的名称
  renewBeforeDay: 10  # 到期前10天重新颁发
  email: "your@mail.com" # 申请证书时设置的邮箱
  hook: "sh /opt/update.sh" # 默认留空, 填写脚本路径后，颁发完证书会主动调用外部脚本
  accessKey: "your access token"   # alidns有key和secret，dnspod只需要access token
  secretKey: "your secret key"

  # 腾讯云dns配置实例
  - providerName: tencentCloud
  enable: false        # 禁用该证书托管,默认为true启用
  domains:
    - "itgod.org"
    - "*.itgod.org"
  saveSSLName:         # 默认留空，指定SSL证书保存的名字，留空则默认使用domains[0].key的名称
  renewBeforeDay: 10  # 到期前10天重新颁发
  email: "your@mail.com" # 申请证书时设置的邮箱
  hook: "" # 默认留空, 填写脚本路径后，颁发完证书会主动调用此处设置的外部脚本或命令
  accessKey: "457y51,your token"  # 填写腾讯云secret id
  secretKey: ""  # 填写腾讯云secret key
```
