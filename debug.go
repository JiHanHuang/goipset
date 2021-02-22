/*
 * @Author: JiHan
 * @Date: 2021-02-20 17:23:25
 * @LastEditTime: 2021-02-22 17:17:50
 * @LastEditors: JiHan
 * @Description:
 * @Usage:
 */

package goipset

import (
	"log"
	"net"

	"github.com/JiHanHuang/goipset/nl"
	"golang.org/x/sys/unix"
)

const debug = true

var debugf = log.Printf

/* Ipset command */
var cmdStr = map[int]string{
	nl.IPSET_CMD_PROTOCOL: "IPSET_CMD_PROTOCOL",
	nl.IPSET_CMD_CREATE:   "IPSET_CMD_CREATE",
	nl.IPSET_CMD_DESTROY:  "IPSET_CMD_DESTROY",
	nl.IPSET_CMD_FLUSH:    "IPSET_CMD_FLUSH",
	nl.IPSET_CMD_RENAME:   "IPSET_CMD_RENAME",
	nl.IPSET_CMD_SWAP:     "IPSET_CMD_SWAP",
	nl.IPSET_CMD_LIST:     "IPSET_CMD_LIST",
	nl.IPSET_CMD_SAVE:     "IPSET_CMD_SAVE",
	nl.IPSET_CMD_ADD:      "IPSET_CMD_ADD",
	nl.IPSET_CMD_DEL:      "IPSET_CMD_DEL",
	nl.IPSET_CMD_TEST:     "IPSET_CMD_TEST",
	nl.IPSET_CMD_HEADER:   "IPSET_CMD_HEADER",
	nl.IPSET_CMD_TYPE:     "IPSET_CMD_TYPE",
}

/* NlMsghdr Flags */
var flagsStr = map[uint16]string{
	0: "exist",
}

/* Attributes at command level */
var attrStr = map[int]string{
	nl.IPSET_ATTR_PROTOCOL:     "IPSET_ATTR_PROTOCOL",
	nl.IPSET_ATTR_SETNAME:      "IPSET_ATTR_SETNAME",
	nl.IPSET_ATTR_TYPENAME:     "IPSET_ATTR_TYPENAME",
	nl.IPSET_ATTR_REVISION:     "IPSET_ATTR_REVISION",
	nl.IPSET_ATTR_FAMILY:       "IPSET_ATTR_FAMILY",
	nl.IPSET_ATTR_FLAGS:        "IPSET_ATTR_FLAGS",
	nl.IPSET_ATTR_DATA:         "IPSET_ATTR_DATA",
	nl.IPSET_ATTR_ADT:          "IPSET_ATTR_ADT",
	nl.IPSET_ATTR_LINENO:       "IPSET_ATTR_LINENO",
	nl.IPSET_ATTR_PROTOCOL_MIN: "IPSET_ATTR_PROTOCOL_MIN",
}

/* CADT specific attributes */
var cadtStr = map[int]string{

	nl.IPSET_ATTR_IP:          "IPSET_ATTR_IP",
	nl.IPSET_ATTR_IP_TO:       "IPSET_ATTR_IP_TO",
	nl.IPSET_ATTR_CIDR:        "IPSET_ATTR_CIDR",
	nl.IPSET_ATTR_PORT:        "IPSET_ATTR_PORT",
	nl.IPSET_ATTR_PORT_TO:     "IPSET_ATTR_PORT_TO",
	nl.IPSET_ATTR_TIMEOUT:     "IPSET_ATTR_TIMEOUT",
	nl.IPSET_ATTR_PROTO:       "IPSET_ATTR_PROTO",
	nl.IPSET_ATTR_CADT_FLAGS:  "IPSET_ATTR_CADT_FLAGS",
	nl.IPSET_ATTR_CADT_LINENO: "IPSET_ATTR_CADT_LINENO",
	nl.IPSET_ATTR_MARK:        "IPSET_ATTR_MARK",
	nl.IPSET_ATTR_MARKMASK:    "IPSET_ATTR_MARKMASK",
	nl.IPSET_ATTR_CADT_MAX:    "IPSET_ATTR_CADT_MAX",
	nl.IPSET_ATTR_HASHSIZE:    "IPSET_ATTR_HASHSIZE",
	nl.IPSET_ATTR_MAXELEM:     "IPSET_ATTR_MAXELEM",
	nl.IPSET_ATTR_NETMASK:     "IPSET_ATTR_NETMASK",
	nl.IPSET_ATTR_PROBES:      "IPSET_ATTR_PROBES",
	nl.IPSET_ATTR_RESIZE:      "IPSET_ATTR_RESIZE",
	nl.IPSET_ATTR_ELEMENTS:    "IPSET_ATTR_ELEMENTS",
	nl.IPSET_ATTR_REFERENCES:  "IPSET_ATTR_REFERENCES",
	nl.IPSET_ATTR_MEMSIZE:     "IPSET_ATTR_MEMSIZE",
	nl.SET_ATTR_CREATE_MAX:    "SET_ATTR_CREATE_MAX",
}

/* ADT specific attributes */
var adtStr = map[int]string{
	nl.IPSET_ATTR_IPADDR_IPV4: "IPV4",
	nl.IPSET_ATTR_IPADDR_IPV6: "IPV6",
	nl.IPSET_ATTR_ETHER:       "IPSET_ATTR_ETHER",
	nl.IPSET_ATTR_NAME:        "IPSET_ATTR_NAME",
	nl.IPSET_ATTR_NAMEREF:     "IPSET_ATTR_NAMEREF",
	nl.IPSET_ATTR_IP2:         "IPSET_ATTR_IP2",
	nl.IPSET_ATTR_CIDR2:       "IPSET_ATTR_CIDR2",
	nl.IPSET_ATTR_IP2_TO:      "IPSET_ATTR_IP2_TO",
	nl.IPSET_ATTR_IFACE:       "IPSET_ATTR_IFACE",
	nl.IPSET_ATTR_BYTES:       "IPSET_ATTR_BYTES",
	nl.IPSET_ATTR_PACKETS:     "IPSET_ATTR_PACKETS",
	nl.IPSET_ATTR_COMMENT:     "IPSET_ATTR_COMMENT",
	nl.IPSET_ATTR_SKBMARK:     "IPSET_ATTR_SKBMARK",
	nl.IPSET_ATTR_SKBPRIO:     "IPSET_ATTR_SKBPRIO",
	nl.IPSET_ATTR_SKBQUEUE:    "IPSET_ATTR_SKBQUEUE",
}

func debugIpsetResult(result GoIPSetResult) {
	if !debug {
		return
	}
	debugf("IpsetResult:\n")
	debugf("%+v\n", result)
}

func debugIpsetRequest(req *nl.NetlinkRequest) {
	if !debug {
		return
	}
	date := req.Serialize()
	debugf("IpsetRequest:\n")
	debugf("Data:%# 04x", date)
	nlhdr := debugUnSerializeNlMsghdr(date[:unix.SizeofNlMsghdr])
	cmd := nlhdr.Type & 0x0f
	flag := nlhdr.Flags >> 8
	debugf("Cmd:%s    Len:%d    Flags:%s[%#04x]    Seq:%d\n",
		cmdStr[int(cmd)], nlhdr.Len, flagsStr[flag], nlhdr.Flags, nlhdr.Seq)
	next := unix.SizeofNlMsghdr
	debugUnSerializeNlData(date[next:])
}

func debugUnSerializeNlMsghdr(buf []byte) (hdr unix.NlMsghdr) {
	_ = buf[15] //bounds check
	native := nl.NativeEndian()
	var b []byte = buf[:4]
	hdr.Len = native.Uint32(b)
	b = buf[4:6]
	hdr.Type = native.Uint16(b)
	b = buf[6:8]
	hdr.Flags = native.Uint16(b)
	b = buf[8:12]
	hdr.Seq = native.Uint32(b)
	b = buf[12:16]
	hdr.Pid = native.Uint32(b)
	return hdr
}

func debugUnSerializeNlData(msg []byte) {
	nf := nl.DeserializeNfgenmsg(msg)
	debugf("NfgenFamily:%d    Version:%d    ResId:%0x\n", nf.NfgenFamily, nf.Version, nf.ResId)

	for attr := range nl.ParseAttributes(msg[4:]) {
		switch attr.Type {
		case nl.IPSET_ATTR_PROTOCOL,
			nl.IPSET_ATTR_REVISION,
			nl.IPSET_ATTR_FLAGS,
			nl.IPSET_ATTR_FAMILY:
			debugf("%s:%v\n", attrStr[int(attr.Type)], attr.Value[0])
		case nl.IPSET_ATTR_SETNAME,
			nl.IPSET_ATTR_TYPENAME:
			debugf("%s:%v\n", attrStr[int(attr.Type)], nl.BytesToString(attr.Value))
		case nl.IPSET_ATTR_DATA | nl.NLA_F_NESTED:
			debugf("%s:\n", attrStr[nl.IPSET_ATTR_DATA])
			debugParseEntry(attr.Value)
			//debugParseAttrData(attr.Value)
		case nl.IPSET_ATTR_ADT | nl.NLA_F_NESTED:
			debugf("%s:\n", attrStr[nl.IPSET_ATTR_ADT])
			//debugParseAttrADT(attr.Value)
		case nl.IPSET_ATTR_LINENO | nl.NLA_F_NESTED:
			debugf("%s:%v\n", attrStr[nl.IPSET_ATTR_LINENO], attr.Value[0])

			break
		default:
			debugf("unknown ipset attribute from kernel: T:%# 04x V:%# 04x %+v", attr.Type, attr.Value, attr)
		}
	}
}

func debugParseAttrData(data []byte) {
	for attr := range nl.ParseAttributes(data) {
		switch attr.Type {
		case nl.IPSET_ATTR_HASHSIZE | nl.NLA_F_NET_BYTEORDER,
			nl.IPSET_ATTR_MAXELEM | nl.NLA_F_NET_BYTEORDER,
			nl.IPSET_ATTR_TIMEOUT | nl.NLA_F_NET_BYTEORDER,
			nl.IPSET_ATTR_ELEMENTS | nl.NLA_F_NET_BYTEORDER,
			nl.IPSET_ATTR_REFERENCES | nl.NLA_F_NET_BYTEORDER,
			nl.IPSET_ATTR_MEMSIZE | nl.NLA_F_NET_BYTEORDER,
			nl.IPSET_ATTR_CADT_FLAGS | nl.NLA_F_NET_BYTEORDER:
			debugf("%s:%d\n", cadtStr[int(attr.Type&nl.NLA_TYPE_MASK)], attr.Uint32())
		default:
			debugf("unknown ipset data attribute from kernel: T:%# 04x V:%# 04x %+v", attr.Type, attr.Value, attr)
		}
	}
}

func debugParseAttrADT(data []byte) {
	for attr := range nl.ParseAttributes(data) {
		switch attr.Type {
		case nl.IPSET_ATTR_DATA | nl.NLA_F_NESTED:
			debugf("%s:\n", attrStr[nl.IPSET_ATTR_DATA])
			debugParseEntry(attr.Value)
		default:
			debugf("unknown ADT attribute from kernel: T:%# 04x V:%# 04x %+v", attr.Type, attr.Value, attr)
		}
	}
}

func debugParseEntry(data []byte) {
	for attr := range nl.ParseAttributes(data) {
		switch attr.Type {
		case nl.IPSET_ATTR_TIMEOUT | nl.NLA_F_NET_BYTEORDER:
			debugf("    %s:%d\n", adtStr[nl.IPSET_ATTR_TIMEOUT], attr.Uint32())
		case nl.IPSET_ATTR_BYTES | nl.NLA_F_NET_BYTEORDER:
		case nl.IPSET_ATTR_PACKETS | nl.NLA_F_NET_BYTEORDER:
			debugf("    %s:%d\n", adtStr[int(attr.Type&nl.NLA_TYPE_MASK)], attr.Uint64())

		case nl.IPSET_ATTR_ETHER:
			debugf("    %s:%x\n", adtStr[nl.IPSET_ATTR_ETHER], net.HardwareAddr(attr.Value))
		case nl.IPSET_ATTR_COMMENT:
			debugf("    %s:%s\n", adtStr[nl.IPSET_ATTR_COMMENT], nl.BytesToString(attr.Value))
		case nl.IPSET_ATTR_IP | nl.NLA_F_NESTED,
			nl.IPSET_ATTR_IP_TO | nl.NLA_F_NESTED:
			debugf("    %s:\n", cadtStr[int(attr.Type&nl.NLA_TYPE_MASK)])
			for attr := range nl.ParseAttributes(attr.Value) {
				switch attr.Type {
				case nl.IPSET_ATTR_IPADDR_IPV4 | nl.NLA_F_NET_BYTEORDER,
					nl.IPSET_ATTR_IPADDR_IPV6 | nl.NLA_F_NET_BYTEORDER:
					debugf("        %s:%s\n", adtStr[int(attr.Type&nl.NLA_TYPE_MASK)], net.IP(attr.Value).String())
				default:
					debugf("unknown nested ADT attribute from kernel: T:%# 04x V:%# 04x %+v", attr.Type, attr.Value, attr)
				}
			}
		case nl.IPSET_ATTR_PORT | nl.NLA_F_NET_BYTEORDER,
			nl.IPSET_ATTR_PORT_TO | nl.NLA_F_NET_BYTEORDER:
			debugf("    %s:%d\n", cadtStr[int(attr.Type&nl.NLA_TYPE_MASK)], attr.Uint16())
		case nl.IPSET_ATTR_PROTO:
			debugf("    %s:%d\n", cadtStr[nl.IPSET_ATTR_PROTO], attr.Uint8())
		case nl.IPSET_ATTR_CADT_LINENO | nl.NLA_F_NET_BYTEORDER:
			debugf("    %s:%d\n", cadtStr[nl.IPSET_ATTR_CADT_LINENO], attr.Uint32())
			break
		default:
			debugf("unknown ADT attribute from kernel: T:%# 04x V:%# 04x %+v", attr.Type, attr.Value, attr)
		}
	}
	return
}
