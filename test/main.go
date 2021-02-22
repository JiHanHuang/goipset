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
		"add":      {cmdAddDel(goipset.Add), "add entry", 2},
		"del":      {cmdAddDel(goipset.Del), "delete entry", 2},
	}

	timeoutVal   *uint32
	timeout      = flag.Int("timeout", -1, "timeout, negative means omit the argument")
	comment      = flag.String("comment", "", "comment")
	withComments = flag.Bool("with-comments", false, "create set with comment support")
	withCounters = flag.Bool("with-counters", false, "create set with counters support")
	withSkbinfo  = flag.Bool("with-skbinfo", false, "create set with skbinfo support")
	replace      = flag.Bool("replace", false, "replace existing set/entry")
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
	err := goipset.Create(args[0], args[1], goipset.GoIpsetCreateOptions{
		Replace:  *replace,
		Timeout:  timeoutVal,
		Comments: *withComments,
		Counters: *withCounters,
		Skbinfo:  *withSkbinfo,
	})
	check(err)
}

func cmdDestroy(args []string) {
	check(goipset.Destroy(args[0]))
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

		var set goipset.Set

		//ip,port
		if strings.Contains(element, ",") {
			en := strings.Split(element, ",")
			if len(en) != 2 {
				fmt.Printf("Invalid entry format '%s'\n", element)
				os.Exit(1)
			}
			ipPort := goipset.SetIPPort{}
			//get ips
			ipEntry := strings.Split(en[0], "-")
			if len(ipEntry) == 2 {
				ipPort.IP = net.ParseIP(ipEntry[0])
				ipPort.IPTO = net.ParseIP(ipEntry[1])
			} else {
				ipPort.IP = net.ParseIP(en[0])
			}

			//get ports
			portEntry := strings.Split(en[1], "-")
			var strPort string
			if i := strings.Index(portEntry[0], ":"); i > 0 {
				proto := portEntry[0][:i]
				switch proto {
				case "tcp":
					ipPort.Proto = unix.IPPROTO_TCP
				case "udp":
					ipPort.Proto = unix.IPPROTO_UDP
				default:
					fmt.Printf("ipset: invalid proto '%s'\n", proto)
					os.Exit(1)
				}
				strPort = portEntry[0][i+1:]
			} else {
				strPort = portEntry[0]
			}
			port, _ := strconv.Atoi(strPort)
			ipPort.Port = uint16(port)
			if len(portEntry) == 2 {
				portto, _ := strconv.Atoi(portEntry[1])
				ipPort.PortTo = uint16(portto)
			}
			set = &ipPort
			//ip
		} else {
			en := strings.Split(element, "-")
			if len(en) == 2 {
				set = &goipset.SetIP{IP: net.ParseIP(en[0]), IPTO: net.ParseIP(en[1])}
			} else {
				set = &goipset.SetIP{IP: net.ParseIP(element)}
			}
		}
		entry := goipset.GoIPSetEntry{
			Timeout: timeoutVal,
			Set:     set,
			Comment: *comment,
			Replace: *replace,
		}

		check(f(setName, &entry))
	}
}

// panic on error
func check(err error) {
	if err != nil {
		panic(err)
	}
}
