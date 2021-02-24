/*
 * @Author: JiHan
 * @Date: 2021-02-22 12:24:19
 * @LastEditTime: 2021-02-24 11:32:03
 * @LastEditors: JiHan
 * @Description:
 * @Usage:
 */

package goipset

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/JiHanHuang/goipset/nl"
	"golang.org/x/sys/unix"
)

type Set interface {
	serializeAttr(parent *nl.RtAttr)
	String() string
	//update(date interface{})
}

var protoStr = map[uint8]string{
	unix.IPPROTO_TCP: "TCP",
	unix.IPPROTO_UDP: "UDP",
}

//SetResult is ipset list result set
type SetResult struct {
	MAC   net.HardwareAddr
	IP    net.IP
	CIDR  uint8
	Port  uint16
	Proto uint8
}

func (set *SetResult) serializeAttr(*nl.RtAttr) {
	return
}
func (set *SetResult) String() string {
	if set.MAC != nil {
		return set.MAC.String()
	}
	retStr := fmt.Sprintf("%s", set.IP.String())
	if set.CIDR > 0 {
		retStr = fmt.Sprintf("%s/%d", retStr, set.CIDR)
	}
	if set.Port > 0 && set.Proto > 0 {
		retStr = fmt.Sprintf("%s,%s:%d", retStr, protoStr[set.Proto], set.Port)
	}
	return retStr
}

//SetIP surpport signal ip and ip-ipto
type SetIP struct {
	IP   net.IP
	IPTO net.IP
}

func (set *SetIP) serializeAttr(parent *nl.RtAttr) {
	if set.IP != nil {
		attrIP := nl.NewRtAttr(nl.IPSET_ATTR_IP|int(nl.NLA_F_NESTED), nil)
		attrIP.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_IPADDR_IPV4|int(nl.NLA_F_NET_BYTEORDER), set.IP.To4()))
		parent.AddChild(attrIP)
		if set.IPTO != nil {
			attrIPTO := nl.NewRtAttr(nl.IPSET_ATTR_IP_TO|int(nl.NLA_F_NESTED), nil)
			attrIPTO.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_IPADDR_IPV4|int(nl.NLA_F_NET_BYTEORDER), set.IPTO.To4()))
			parent.AddChild(attrIPTO)
		}
	}
}
func (set *SetIP) String() string {
	if set.IPTO == nil {
		return fmt.Sprintf("%s", set.IP.String())
	}
	return fmt.Sprintf("%s-%s", set.IP.String(), set.IPTO.String())
}

/*SetIPPort support ip,port format
ip,port:
	entry type: ip,<proto:>port
	ip type:x.x.x.x or x.x.x.x-x.x.x.x
	port type: xx or xx-xx
	proto type: udp or tcp or null
*/
type SetIPPort struct {
	Name   string
	IP     net.IP
	IPTO   net.IP
	Port   uint16
	PortTo uint16
	Proto  uint8
}

func (set *SetIPPort) serializeAttr(parent *nl.RtAttr) {
	if set.IP != nil && set.Port > 0 {
		attrIP := nl.NewRtAttr(nl.IPSET_ATTR_IP|int(nl.NLA_F_NESTED), nil)
		attrIP.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_IPADDR_IPV4|int(nl.NLA_F_NET_BYTEORDER), set.IP.To4()))
		parent.AddChild(attrIP)
		if set.IPTO != nil {
			attrIPTO := nl.NewRtAttr(nl.IPSET_ATTR_IP_TO|int(nl.NLA_F_NESTED), nil)
			attrIPTO.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_IPADDR_IPV4|int(nl.NLA_F_NET_BYTEORDER), set.IPTO.To4()))
			parent.AddChild(attrIPTO)
		}
		bytesPort := make([]byte, 2)
		binary.BigEndian.PutUint16(bytesPort, uint16(set.Port))
		parent.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_PORT|int(nl.NLA_F_NET_BYTEORDER), bytesPort))
		if set.PortTo > 0 {
			bytesPortTo := make([]byte, 2)
			binary.BigEndian.PutUint16(bytesPortTo, uint16(set.PortTo))
			parent.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_PORT_TO|int(nl.NLA_F_NET_BYTEORDER), bytesPortTo))
		}
		parent.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_PROTO, nl.Uint8Attr(set.Proto)))
	}
}
func (set *SetIPPort) String() string {
	ipStr := fmt.Sprintf("%s", set.IP.String())
	if set.IPTO != nil {
		ipStr = fmt.Sprintf("%s-%s", set.IP.String(), set.IPTO.String())
	}
	portStr := fmt.Sprintf("%s:%d", protoStr[set.Proto], set.Port)
	if set.PortTo > 0 {
		portStr = fmt.Sprintf("%s:%d-%d", protoStr[set.Proto], set.Port, set.PortTo)
	}
	return fmt.Sprintf("%s,%s", ipStr, portStr)
}

//SetMac
type SetMac struct {
	MAC net.HardwareAddr
}

func (set *SetMac) serializeAttr(parent *nl.RtAttr) {
	if set.MAC != nil {
		parent.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_ETHER, set.MAC))
	}
}
func (set *SetMac) String() string {
	return set.MAC.String()
}

//SetNet
type SetNet struct {
	IP   net.IP
	CIDR uint8
}

func (set *SetNet) serializeAttr(parent *nl.RtAttr) {
	if set.IP != nil {
		attrIP := nl.NewRtAttr(nl.IPSET_ATTR_IP|int(nl.NLA_F_NESTED), nil)
		attrIP.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_IPADDR_IPV4|int(nl.NLA_F_NET_BYTEORDER), set.IP.To4()))
		parent.AddChild(attrIP)
		if set.CIDR > 0 {
			parent.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_CIDR, nl.Uint8Attr(set.CIDR)))
		}
	}
}
func (set *SetNet) String() string {
	if set.CIDR > 0 {
		return fmt.Sprintf("%s/%d", set.IP.String(), set.CIDR)
	}
	return fmt.Sprintf("%s", set.IP.String())
}

//SetNetPort
type SetNetPort struct {
	IP     net.IP
	CIDR   uint8
	Port   uint16
	PortTo uint16
	Proto  uint8
}

func (set *SetNetPort) serializeAttr(parent *nl.RtAttr) {
	if set.IP != nil {
		attrIP := nl.NewRtAttr(nl.IPSET_ATTR_IP|int(nl.NLA_F_NESTED), nil)
		attrIP.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_IPADDR_IPV4|int(nl.NLA_F_NET_BYTEORDER), set.IP.To4()))
		parent.AddChild(attrIP)
		if set.CIDR > 0 {
			parent.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_CIDR, nl.Uint8Attr(set.CIDR)))
		}
		bytesPort := make([]byte, 2)
		binary.BigEndian.PutUint16(bytesPort, uint16(set.Port))
		parent.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_PORT|int(nl.NLA_F_NET_BYTEORDER), bytesPort))
		if set.PortTo > 0 {
			bytesPortTo := make([]byte, 2)
			binary.BigEndian.PutUint16(bytesPortTo, uint16(set.PortTo))
			parent.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_PORT_TO|int(nl.NLA_F_NET_BYTEORDER), bytesPortTo))
		}
		parent.AddChild(nl.NewRtAttr(nl.IPSET_ATTR_PROTO, nl.Uint8Attr(set.Proto)))
	}
}
func (set *SetNetPort) String() string {
	ipStr := fmt.Sprintf("%s", set.IP.String())
	if set.CIDR > 0 {
		ipStr = fmt.Sprintf("%s/%d", set.IP.String(), set.CIDR)
	}
	portStr := fmt.Sprintf("%s:%d", protoStr[set.Proto], set.Port)
	if set.PortTo > 0 {
		portStr = fmt.Sprintf("%s:%d-%d", protoStr[set.Proto], set.Port, set.PortTo)
	}
	return fmt.Sprintf("%s,%s", ipStr, portStr)
}
