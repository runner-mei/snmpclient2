package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/runner-mei/snmpclient2"
	//"web"
)

var (
	laddr       = flag.String("laddr", "0.0.0.0:0", "the address of bind, default: '0.0.0.0:0'")
	network     = flag.String("network", "udp4", "the family of address, default: 'udp4'")
	timeout     = flag.Int("timeout", 5, "the second of timeout, default: '5'")
	port        = flag.String("port", "161", "the port of address, default: '161'")
	communities = flag.String("communities", "public;public1", "the community of snmp")
	version     = flag.String("version", "v2c", "the version of snmp")
	username    = flag.String("username", "", "the username of snmp v3")
)

func main() {
	flag.Parse()

	targets := flag.Args()
	if nil == targets || 1 != len(targets) {
		flag.Usage()
		return
	}

	scanner := snmpclient2.NewPingers(256)

	version, err := snmpclient2.ParseVersion(*version)
	if err != nil {
		fmt.Println(err)
		return
	}

	if version == snmpclient2.V3 {
		e := scanner.ListenV3(*network, *laddr, *username)
		if nil != e {
			fmt.Println(e)
			return
		}
	} else {
		for _, community := range strings.Split(*communities, ";") {
			e := scanner.Listen(*network, *laddr, snmpclient2.V2c, community)
			if nil != e {
				fmt.Println(e)
				return
			}
		}
	}

	defer scanner.Close()

	ip_range, err := ParseIPRange(targets[0])
	if nil != err {
		fmt.Println(err)
		return
	}
	var wait sync.WaitGroup
	is_stopped := int32(0)
	go func() {
		for i := 0; i < scanner.Length(); i++ {
			ip_range.Reset()

			if i != 0 {
				time.Sleep(500 * time.Millisecond)
			}

			for ip_range.HasNext() {
				err = scanner.Send(i, net.JoinHostPort(ip_range.Current().String(), *port))
				if nil != err {
					fmt.Println(err)
					goto end
				}
			}
		}
	end:
		atomic.StoreInt32(&is_stopped, 1)
		wait.Done()
	}()
	wait.Add(1)

	for {
		ra, t, err := scanner.Recv(time.Duration(*timeout) * time.Second)
		if nil != err {
			if err == snmpclient2.TimeoutError {
				fmt.Println(err)
			} else if 0 == atomic.LoadInt32(&is_stopped) {
				continue
			}
			break
		}
		fmt.Println(ra, t)
	}
	wait.Wait()
}
