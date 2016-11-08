# Install & Test

1. `cd comet/bin`
2. `go build -o server .`
3. `./server`
4. open a new session, `mosquitto_sub -c -i jason -t test -u UserJK -P sdf -d -q 0`
4. open another session, `mosquitto_pub -t test -m "ggg" -i Jason -P Pass -u UserName -d -q 0`


# Roadmap

- [x] comet
- [ ] router

# IM Server 架构

## 组件
Comet (客户端连接), Logic (业务处理), Router (消息分发)

## Comet
- 负责连接客户端，需要实现 mqtt 协议 (sub/pub)
- pub 的消息 通过 rpc 交给 logic 做修改 (敏感词过滤, 等)，最终交给 router 分发消息。
- 上报 username 和 topic 信息给 router
- 每个client订阅各自的message queue, 客户端收到的消息都来自这个queue
- comet 只维护本机连接的client, 所以支持水平扩展。给客户端提供接口查询可用ip, 客户端根据ping值选择连接
- 当客户订阅topic变化时，需要 上报 router。
- comet 故障时，及时通知 router， 可以借助 etcd

cleanSession=0 时，如何保存 user session? 历史消息如何保存; 下次登陆到其他comet上，session 如何转移
//- user 退出登陆后，如果cleanSession=0, 则在本地新建一个user_offline_file, 记录离线消息, 向 Logic 注册离线时间，cometID。定时清理 user 和 offline_file。
//- user 再次登陆时，向 Logic 询问 上次登陆的 cometID, ip。向之前的 comet 请求转移 user session.
- 如果queue使用 kafka的话，离线记录是不会丢的，只要设置适当的topic消息保留长度就可以保证一定时间内消息不丢。
- session 可以集中管理，因为是根据 username 读写的，所以性能不会差。

## Logic
- RPC server, 业务逻辑，修改 message (from user, message qos, message body, to topic)
- RPC 鉴权

## Router
- 接受 comet发送的 (username, topic, action), 记录 topic -> usernames 的映射
- 接受 comet.publish(topic, payload)的调用，找到topic, 找到所有相关的usernames, 
把消息放到queue的每个client的topic中。 queue 的 topic 与 username 一一对应。 

## Roadmap

### Step I
- comet 实现 mqtt 最简部分, qos=0, topic 不支持通配符, cleanSession=true
- router 单机部署

### Following Steps
- comet 实现 mqtt 全部协议
- router 应用 raft


> 参考b站 [goim 项目](https://github.com/Terry-Mao/goim), [文章](https://mp.weixin.qq.com/s?__biz=MjM5NzAwNDI4Mg==&mid=2652190998&idx=1&sn=5023e23660ede074c9eb48e166a8faf3&scene=4&uin=NjUwNjA2Njgx&key=305bc10ec50ec19b3be7fc02fb1e31e65fe4d6d9316142020c045362c0e1eafd1a78d57247eb2bc8dd1c23751e1e79de&devicetype=iMac+MacBookAir7%2C1+OSX+OSX+10.11.5+build(15F34)&version=11020201&lang=zh_CN&pass_ticket=jg6FcNlzsgy4ssoPzjNwvKBBxc05AQJUdzQP2P5PCP6XzqP%2FkXdem%2BzEuy7wuzDg)
