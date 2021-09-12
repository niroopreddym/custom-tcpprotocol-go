package main

import (
	"fmt"

	"github.com/niroopreddym/custom-tcpprotocol-go/mtsclient"
)

func main() {
	tcpConnect := mtsclient.NewTCPConnect("127.0.0.1", 10001, 10000)
	tcpConnect.ConnectAndLogin()
	fmt.Println(tcpConnect)

	// ln, err := net.Listen("tcp", "127.0.0.1:2345")

	// if err != nil {
	// 	panic(err)
	// }

	// defer ln.Close()

	// go func() {

	// 	for {
	// 		conn, err := ln.Accept()
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		scanner := bufio.NewScanner(conn)
	// 		for scanner.Scan() {
	// 			io.WriteString(conn, scanner.Text())
	// 		}

	// 		// conn.Close()
	// 	}
	// }()

	// for {
	// 	conn, err := ln.Accept()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	io.WriteString(conn, fmt.Sprintf("hii"))
	// 	// conn.Close()
	// }
}
