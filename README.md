# goipset
A golang ipset client uses netlink to communicate with the kernel.
用golang写的一个ipset客户端，使用netlink与内核进行通信。

其中netlink通信部分是基于：[netlink](https://github.com/vishvananda/netlink)
ipset相关功能也参考了其中的写法。

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
package main

import (
	"log"
	"net"

	"github.com/JiHanHuang/goipset"
)

func main() {
	ipsetName := "test"
	protocol, err := goipset.Protocol()
	check(err)
	log.Println("Protocol:", protocol)

	err = goipset.Create(ipsetName, "hash:ip", goipset.GoIpsetCreateOptions{})
	check(err)

	entry := goipset.GoIPSetEntry{
		Set: &goipset.SetIP{IP: net.ParseIP("1.1.1.1")},
	}
	err = goipset.Add(ipsetName, &entry)
	check(err)

	result, err := goipset.List(ipsetName)
	check(err)
	log.Printf("List:%v", result.Entries)
}

func check(err error) {
	if err != nil {
		log.Fatalf("Err: %v", err)
	}
}
```

## 更多
后续将持续补齐ipset的相关功能，欢迎随时交流。
提供一些ipset和iptables的配合使用的相关文档。
