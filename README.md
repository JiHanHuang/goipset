# goipset
A golang ipset client uses netlink to communicate with the kernel.
用golang写的一个ipset客户端，使用netlink与内核进行通信。

## 基础环境
内核必须包含ipset的ko, 可以通过以下方式确认：
```
lsmod | grep ip_set
```
否则你需要通过以下方式安装ipset：
```
yum install ipset
```
或者
```
insmod <your_path>/ip_set.ko
```
更多ipset信息请参考[官网](http://ipset.netfilter.org/ipset.man.html)

## 使用指南

**作为单纯客户端使用**
你可以尝试编译`test`目录下的`main.go`:
```
go build -o goipset main.go
```
你会得到一个可执行文件`goipset`，你可以像使用标准(c)版本的ipset
一样使用这个客户端(可能部分命令还没有支持，待完善)。比如:
```
# ./goipset add hash_ip 1.1.1.1
```
**作为三方库调用**
比如：
```go

```

## 更多
后续将持续补齐ipset的相关功能，如果有需求可以提pr。
提供一些ipset和iptables的配合使用的相关文档。