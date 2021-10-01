package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/niroopreddym/custom-tcpprotocol-go/enum"
	"github.com/niroopreddym/custom-tcpprotocol-go/mtsclient"
)

//TestElapsedSeconds intial time to trigger the fan out test events
var TestElapsedSeconds = 1

func main() {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println(err)
		}
	}()

	tcpConnect := mtsclient.NewTCPConnect("127.0.0.1", 10001, 10000)

	defer tcpConnect.Conn.Close()
	//do all operations on top of TLS
	tcpConnect.WithTLS(nil)

	isAuthenticatedChan := make(chan bool)
	errorChan := make(chan error)

	tcpConnect.Wg.Add(1)
	tcpConnect.ConnectAndLogin(isAuthenticatedChan, errorChan)
	// err := <-errorChan
	// if err != nil {
	// 	fmt.Println(err)
	// }

	if <-tcpConnect.ServerBootDone {
		tcpConnect.Wg.Add(1)
		go sendOPLTestMessages(tcpConnect, &tcpConnect.Wg)
	}

	log.Println("Ending MtsClientExample go")
	tcpConnect.Wg.Wait()
}

func sendOPLTestMessages(tcpConnect *mtsclient.TCPConnect, wg *sync.WaitGroup) {

	testEventsMap := map[int]enum.TestEvent{}
	// Create some test events
	testEventsMap[3] = enum.SendOPLPayload
	testEventsMap[6] = enum.SendOPLPayload
	testEventsMap[9] = enum.SendOPLPayload
	testEventsMap[12] = enum.Exit

	// Begin example
	log.Println("Starting MtsClientExample in go")

	exitTest := false
	for !exitTest {

		// Send messages within this program loop
		if event, ok := testEventsMap[TestElapsedSeconds]; ok {
			//do something here
			switch event {
			case enum.SendOPLPayload:
				tcpConnect.SendTestOPLPayload(nil)
				break
			case enum.Exit:
				exitTest = true
				break
			}
		}

		if !exitTest {
			time.Sleep(1 * time.Second)
			TestElapsedSeconds++
		}
	}

	wg.Done()
}
