package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/runner-mei/snmpclient2"
)

var (
	address = flag.String("listen", ":161", "")
	file    = flag.String("file", "", "")
)

func main() {
	flag.Parse()

	if "" == *file {
		fmt.Println("file is required.")
		return
	}
	srv, e := snmpclient2.NewUdpServerFromFile("sim", *address, *file, true)
	if nil != e {
		fmt.Println(e)
		return
	}
	fmt.Println("listen at:", srv.GetPort())

	os.Stdin.Read(make([]byte, 1))
	srv.Close()
}
