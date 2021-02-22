package goipset

import (
	"fmt"
	"net"
	"sync"
	"syscall"

	"github.com/JiHanHuang/goipset/nl"
	"golang.org/x/sys/unix"
)

// GoIPSetEntry is used for adding, updating, retreiving and deleting entries
type GoIPSetEntry struct {
	Comment string

	Set

	Timeout *uint32
	Packets *uint64
	Bytes   *uint64

	Replace bool // replace existing entry
}

// GoIPSetResult is the result of a dump request for a set
type GoIPSetResult struct {
	Nfgenmsg *nl.Nfgenmsg
	Protocol uint8
	Revision uint8
	Family   uint8
	Flags    uint8
	SetName  string
	TypeName string

	HashSize     uint32
	NumEntries   uint32
	MaxElements  uint32
	References   uint32
	SizeInMemory uint32
	CadtFlags    uint32
	Timeout      *uint32

	Entries []GoIPSetEntry
}

// GoIpsetCreateOptions is the options struct for creating a new ipset
type GoIpsetCreateOptions struct {
	Replace  bool // replace existing ipset
	Timeout  *uint32
	Counters bool
	Comments bool
	Skbinfo  bool
}

// GoIpset using save sockets...
type GoIpset struct {
	sockets   map[int]*nl.SocketHandle
	domainSet sync.Map
}

var gipset = GoIpset{}

// Protocol returns the ipset protocol version from the kernel
func Protocol() (uint8, error) {
	return gipset.Protocol()
}

// Create creates a new ipset
func Create(setname, typename string, options GoIpsetCreateOptions) error {
	return gipset.Create(setname, typename, options)
}

// Destroy destroys an existing ipset
func Destroy(setname string) error {
	return gipset.Destroy(setname)
}

// Flush flushes an existing ipset
func Flush(setname string) error {
	return gipset.Flush(setname)
}

// List dumps an specific ipset.
func List(setname string) (*GoIPSetResult, error) {
	return gipset.List(setname)
}

// ListAll dumps all ipsets.
func ListAll() ([]GoIPSetResult, error) {
	return gipset.ListAll()
}

// Add adds an entry to an existing ipset.
func Add(setname string, entry *GoIPSetEntry) error {
	return gipset.ipsetAddDel(nl.IPSET_CMD_ADD, setname, entry)
}

// Del deletes an entry from an existing ipset.
func Del(setname string, entry *GoIPSetEntry) error {
	return gipset.ipsetAddDel(nl.IPSET_CMD_DEL, setname, entry)
}

func (g *GoIpset) Protocol() (uint8, error) {
	req := g.newIpsetRequest(nl.IPSET_CMD_PROTOCOL)
	msgs, err := req.Execute(unix.NETLINK_NETFILTER, 0)

	if err != nil {
		return 0, err
	}

	return ipsetUnserialize(msgs).Protocol, nil
}

func (g *GoIpset) ipsetType(typename string) (GoIPSetResult, error) {
	req := g.newIpsetRequest(nl.IPSET_CMD_TYPE)
	req.Flags |= unix.NLM_F_EXCL
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_TYPENAME, nl.ZeroTerminated(typename)))
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_FAMILY, nl.Uint8Attr(2)))

	debugIpsetRequest(req)

	msgs, err := req.Execute(unix.NETLINK_NETFILTER, 0)
	if err != nil {
		return GoIPSetResult{}, err
	}

	return ipsetUnserialize(msgs), nil
}

func (g *GoIpset) Create(setname, typename string, options GoIpsetCreateOptions) error {

	result, err := g.ipsetType(typename)
	if err != nil {
		return err
	}

	req := g.newIpsetRequest(nl.IPSET_CMD_CREATE)

	if !options.Replace {
		req.Flags |= unix.NLM_F_EXCL
	}

	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_SETNAME, nl.ZeroTerminated(setname)))
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_TYPENAME, nl.ZeroTerminated(typename)))
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_REVISION, nl.Uint8Attr(result.Revision)))
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_FAMILY, nl.Uint8Attr(2)))

	data := nl.NewRtAttr(nl.IPSET_ATTR_DATA|int(nl.NLA_F_NESTED), nil)

	if timeout := options.Timeout; timeout != nil {
		data.AddChild(&nl.Uint32Attribute{Type: nl.IPSET_ATTR_TIMEOUT | nl.NLA_F_NET_BYTEORDER, Value: *timeout})
	}

	var cadtFlags uint32

	if options.Comments {
		cadtFlags |= nl.IPSET_FLAG_WITH_COMMENT
	}
	if options.Counters {
		cadtFlags |= nl.IPSET_FLAG_WITH_COUNTERS
	}
	if options.Skbinfo {
		cadtFlags |= nl.IPSET_FLAG_WITH_SKBINFO
	}

	if cadtFlags != 0 {
		data.AddChild(&nl.Uint32Attribute{Type: nl.IPSET_ATTR_CADT_FLAGS | nl.NLA_F_NET_BYTEORDER, Value: cadtFlags})
	}

	req.AddData(data)

	debugIpsetRequest(req)

	_, err = ipsetExecute(req)
	return err
}

func (g *GoIpset) Destroy(setname string) error {
	req := g.newIpsetRequest(nl.IPSET_CMD_DESTROY)
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_SETNAME, nl.ZeroTerminated(setname)))
	_, err := ipsetExecute(req)
	return err
}

func (g *GoIpset) Flush(setname string) error {
	req := g.newIpsetRequest(nl.IPSET_CMD_FLUSH)
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_SETNAME, nl.ZeroTerminated(setname)))
	_, err := ipsetExecute(req)
	return err
}

func (g *GoIpset) List(name string) (*GoIPSetResult, error) {
	req := g.newIpsetRequest(nl.IPSET_CMD_LIST)
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_SETNAME, nl.ZeroTerminated(name)))

	msgs, err := ipsetExecute(req)
	if err != nil {
		return nil, err
	}

	result := ipsetUnserialize(msgs)
	return &result, nil
}

func (g *GoIpset) ListAll() ([]GoIPSetResult, error) {
	req := g.newIpsetRequest(nl.IPSET_CMD_LIST)

	msgs, err := ipsetExecute(req)
	if err != nil {
		return nil, err
	}

	result := make([]GoIPSetResult, len(msgs))
	for i, msg := range msgs {
		result[i].unserialize(msg)
	}

	return result, nil
}

func (g *GoIpset) ipsetAddDel(nlCmd int, setname string, entry *GoIPSetEntry) error {
	if entry.Set == nil {
		return fmt.Errorf("Set is nil in GoIPSetEntry")
	}
	req := g.newIpsetRequest(nlCmd)

	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_SETNAME, nl.ZeroTerminated(setname)))
	data := nl.NewRtAttr(nl.IPSET_ATTR_DATA|int(nl.NLA_F_NESTED), nil)

	if !entry.Replace {
		req.Flags |= unix.NLM_F_EXCL
	}

	if entry.Timeout != nil {
		data.AddChild(&nl.Uint32Attribute{Type: nl.IPSET_ATTR_TIMEOUT | nl.NLA_F_NET_BYTEORDER, Value: *entry.Timeout})
	}
	entry.Set.serializeAttr(data)

	data.AddChild(&nl.Uint32Attribute{Type: nl.IPSET_ATTR_LINENO | nl.NLA_F_NET_BYTEORDER, Value: 0})
	req.AddData(data)

	debugIpsetRequest(req)

	_, err := ipsetExecute(req)
	return err
}

func (g *GoIpset) newIpsetRequest(cmd int) *nl.NetlinkRequest {
	req := nl.NewNetlinkRequest(cmd|(unix.NFNL_SUBSYS_IPSET<<8), nl.GetIpsetFlags(cmd))

	// Add the netfilter header
	msg := &nl.Nfgenmsg{
		NfgenFamily: uint8(unix.AF_INET),
		Version:     nl.NFNETLINK_V0,
		ResId:       0,
	}
	req.AddData(msg)
	req.AddData(nl.NewRtAttr(nl.IPSET_ATTR_PROTOCOL, nl.Uint8Attr(nl.IPSET_PROTOCOL)))

	return req
}

func ipsetExecute(req *nl.NetlinkRequest) (msgs [][]byte, err error) {
	msgs, err = req.Execute(unix.NETLINK_NETFILTER, 0)

	if err != nil {
		if errno := int(err.(syscall.Errno)); errno >= nl.IPSET_ERR_PRIVATE {
			err = nl.IPSetError(uintptr(errno))
		}
	}
	return
}

func ipsetUnserialize(msgs [][]byte) (result GoIPSetResult) {
	for _, msg := range msgs {
		result.unserialize(msg)
	}
	debugIpsetResult(result)
	return result
}

func (result *GoIPSetResult) unserialize(msg []byte) error {
	result.Nfgenmsg = nl.DeserializeNfgenmsg(msg)

	for attr := range nl.ParseAttributes(msg[4:]) {
		switch attr.Type {
		case nl.IPSET_ATTR_PROTOCOL:
			result.Protocol = attr.Value[0]
		case nl.IPSET_ATTR_SETNAME:
			result.SetName = nl.BytesToString(attr.Value)
		case nl.IPSET_ATTR_TYPENAME:
			result.TypeName = nl.BytesToString(attr.Value)
		case nl.IPSET_ATTR_REVISION:
			result.Revision = attr.Value[0]
		case nl.IPSET_ATTR_FAMILY:
			result.Family = attr.Value[0]
		case nl.IPSET_ATTR_FLAGS:
			result.Flags = attr.Value[0]
		case nl.IPSET_ATTR_DATA | nl.NLA_F_NESTED:
			if err := result.parseAttrData(attr.Value); err != nil {
				return err
			}
		case nl.IPSET_ATTR_ADT | nl.NLA_F_NESTED:
			if err := result.parseAttrADT(attr.Value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown ipset attribute from kernel: %+v %v", attr, attr.Type&nl.NLA_TYPE_MASK)
		}
	}
	return nil
}

func (result *GoIPSetResult) parseAttrData(data []byte) error {
	for attr := range nl.ParseAttributes(data) {
		switch attr.Type {
		case nl.IPSET_ATTR_HASHSIZE | nl.NLA_F_NET_BYTEORDER:
			result.HashSize = attr.Uint32()
		case nl.IPSET_ATTR_MAXELEM | nl.NLA_F_NET_BYTEORDER:
			result.MaxElements = attr.Uint32()
		case nl.IPSET_ATTR_TIMEOUT | nl.NLA_F_NET_BYTEORDER:
			val := attr.Uint32()
			result.Timeout = &val
		case nl.IPSET_ATTR_ELEMENTS | nl.NLA_F_NET_BYTEORDER:
			result.NumEntries = attr.Uint32()
		case nl.IPSET_ATTR_REFERENCES | nl.NLA_F_NET_BYTEORDER:
			result.References = attr.Uint32()
		case nl.IPSET_ATTR_MEMSIZE | nl.NLA_F_NET_BYTEORDER:
			result.SizeInMemory = attr.Uint32()
		case nl.IPSET_ATTR_CADT_FLAGS | nl.NLA_F_NET_BYTEORDER:
			result.CadtFlags = attr.Uint32()
		default:
			return fmt.Errorf("unknown ipset data attribute from kernel: %+v %v", attr, attr.Type&nl.NLA_TYPE_MASK)
		}
	}
	return nil
}

func (result *GoIPSetResult) parseAttrADT(data []byte) error {
	for attr := range nl.ParseAttributes(data) {
		switch attr.Type {
		case nl.IPSET_ATTR_DATA | nl.NLA_F_NESTED:
			entry, err := parseIPSetEntry(attr.Value)
			if err != nil {
				return err
			}
			result.Entries = append(result.Entries, entry)
		default:
			return fmt.Errorf("unknown ADT attribute from kernel: %+v %v", attr, attr.Type&nl.NLA_TYPE_MASK)
		}
	}
	return nil
}

func parseIPSetEntry(data []byte) (entry GoIPSetEntry, err error) {
	for attr := range nl.ParseAttributes(data) {
		switch attr.Type {
		case nl.IPSET_ATTR_TIMEOUT | nl.NLA_F_NET_BYTEORDER:
			val := attr.Uint32()
			entry.Timeout = &val
		case nl.IPSET_ATTR_BYTES | nl.NLA_F_NET_BYTEORDER:
			val := attr.Uint64()
			entry.Bytes = &val
		case nl.IPSET_ATTR_PACKETS | nl.NLA_F_NET_BYTEORDER:
			val := attr.Uint64()
			entry.Packets = &val
		case nl.IPSET_ATTR_ETHER:
			entry.Set = &SetMac{net.HardwareAddr(attr.Value)}
		case nl.IPSET_ATTR_COMMENT:
			entry.Comment = nl.BytesToString(attr.Value)
		case nl.IPSET_ATTR_IP | nl.NLA_F_NESTED:
			set := SetIP{}
			for attr := range nl.ParseAttributes(attr.Value) {
				switch attr.Type {
				case nl.IPSET_ATTR_IP:
					set.IP = net.IP(attr.Value)
				default:
					err = fmt.Errorf("unknown nested ADT attribute from kernel: %+v", attr)
				}
			}
			entry.Set = &set
		default:
			err = fmt.Errorf("unknown ADT attribute from kernel: %+v", attr)
		}
	}
	return
}
