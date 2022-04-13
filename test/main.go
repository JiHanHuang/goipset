// +build linux

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/JiHanHuang/goipset"
	"golang.org/x/sys/unix"
)

type command struct {
	Function    func([]string)
	Description string
	ArgCount    int
}

var (
	commands = map[string]command{
		"protocol": {cmdProtocol, "prints the protocol version", 0},
		"create":   {cmdCreate, "creates a new ipset", 2},
		"destroy":  {cmdDestroy, "creates a new ipset", 1},
		"list":     {cmdList, "list specific ipset", 1},
		"listall":  {cmdListAll, "list all ipsets", 0},
		"flush":    {cmdFlush, "list all ipsets", 1},
		"add":      {cmdAddDel(goipset.Add), "add entry", 2},
		"del":      {cmdAddDel(goipset.Del), "delete entry", 2},
	}

	timeoutVal   *uint32
	timeout      = flag.Int("timeout", -1, "timeout, negative means omit the argument")
	comment      = flag.String("comment", "", "comment")
	family       = flag.String("family", "inet", "inet or inet6")
	withComments = flag.Bool("with-comments", false, "create set with comment support")
	withCounters = flag.Bool("with-counters", false, "create set with counters support")
	withSkbinfo  = flag.Bool("with-skbinfo", false, "create set with skbinfo support")
	replace      = flag.Bool("replace", false, "replace existing set/entry")
	debug        = flag.Bool("debug", false, "set debug mode")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	if *timeout >= 0 {
		v := uint32(*timeout)
		timeoutVal = &v
	}

	log.SetFlags(log.Lshortfile)

	cmdName := args[0]
	args = args[1:]

	cmd, exist := commands[cmdName]
	if !exist {
		fmt.Printf("Unknown command '%s'\n\n", cmdName)
		printUsage()
		os.Exit(1)
	}

	if cmd.ArgCount != len(args) {
		fmt.Printf("Invalid number of arguments. expected=%d given=%d\n", cmd.ArgCount, len(args))
		os.Exit(1)
	}

	goipset.Debug = *debug

	cmd.Function(args)
}

func printUsage() {
	fmt.Printf("Usage: %s COMMAND [args] [-flags]\n\n", os.Args[0])
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Println("Available commands:")
	for _, name := range names {
		fmt.Printf("  %-15v %s\n", name, commands[name].Description)
	}
	fmt.Println("\nAvailable flags:")
	flag.PrintDefaults()
}

func cmdProtocol(_ []string) {
	protocol, err := goipset.Protocol()
	check(err)
	log.Println("Protocol:", protocol)
}

func cmdCreate(args []string) {
	f := unix.AF_INET
	if *family == "inet6" {
		f = unix.AF_INET6
	}
	err := goipset.Create(args[0], args[1], goipset.GoIpsetCreateOptions{
		Replace:  *replace,
		Timeout:  timeoutVal,
		Comments: *withComments,
		Counters: *withCounters,
		Skbinfo:  *withSkbinfo,
		Family:   f,
	})
	check(err)
}

func cmdDestroy(args []string) {
	check(goipset.Destroy(args[0]))
}

func cmdFlush(args []string) {
	check(goipset.Flush(args[0]))
}

func cmdList(args []string) {
	result, err := goipset.List(args[0])
	check(err)
	log.Printf("%+v", result)
}

func cmdListAll(args []string) {
	result, err := goipset.ListAll()
	check(err)
	for _, ipset := range result {
		log.Printf("%+v", ipset)
	}
}

func cmdAddDel(f func(string, *goipset.GoIPSetEntry) error) func([]string) {
	return func(args []string) {
		setName := args[0]
		element := args[1]

		entry := goipset.GoIPSetEntry{
			Timeout: timeoutVal,
			Set:     parseIPSetSet(element),
			Comment: *comment,
			Replace: *replace,
		}

		check(f(setName, &entry))
	}
}

func parseIPSetSet(element string) goipset.Set {
	var set goipset.Set
	if strings.Contains(element, "/") {
		if strings.Contains(element, ",") { //net,port
			netPort := goipset.SetNetPort{}
			en := strings.Split(element, ",")
			_ = en[1]
			netPort.IP, netPort.CIDR = parseNet(en[0])
			netPort.Port, netPort.PortTo, netPort.Proto = parsePort(en[1])
			set = &netPort
		} else { //net
			net := goipset.SetNet{}
			net.IP, net.CIDR = parseNet(element)
			set = &net
		}
	} else {
		if strings.Contains(element, ",") { //ip,port
			ipPort := goipset.SetIPPort{}
			en := strings.Split(element, ",")
			_ = en[1]
			ipPort.IP, ipPort.IPTO = parseIP(en[0])
			ipPort.Port, ipPort.PortTo, ipPort.Proto = parsePort(en[1])
			set = &ipPort
		} else { //ip
			ip := goipset.SetIP{}
			ip.IP, ip.IPTO = parseIP(element)
			set = &ip
		}

	}
	return set
}

func parsePort(element string) (p, pTo uint16, pro uint8) {
	portEntry := strings.Split(element, "-")
	var strPort string
	if i := strings.Index(portEntry[0], ":"); i > 0 {
		proto := portEntry[0][:i]
		proto = strings.ToLower(proto)
		switch proto {
		case "tcp":
			pro = unix.IPPROTO_TCP
		case "udp":
			pro = unix.IPPROTO_UDP
		default:
			fmt.Printf("ipset: invalid proto '%s'\n", proto)
			os.Exit(1)
		}
		strPort = portEntry[0][i+1:]
	} else {
		pro = unix.IPPROTO_TCP
		strPort = portEntry[0]
	}
	port, _ := strconv.Atoi(strPort)
	p = uint16(port)
	if len(portEntry) == 2 {
		portto, _ := strconv.Atoi(portEntry[1])
		p = uint16(portto)
	}
	return
}

func parseIP(element string) (ip, ipto net.IP) {
	en := strings.Split(element, "-")
	if len(en) == 2 {
		ip = net.ParseIP(en[0])
		ipto = net.ParseIP(en[1])
	} else {
		ip = net.ParseIP(element)
	}
	return
}

func parseNet(element string) (ip net.IP, cidr uint8) {
	ip, cidro, err := net.ParseCIDR(element)
	if err != nil {
		ip = net.ParseIP(element)
	}
	cidrPrefix, _ := cidro.Mask.Size()
	cidr = uint8(cidrPrefix)
	return
}

// panic on error
func check(err error) {
	if err != nil {
		panic(err)
	}
}
