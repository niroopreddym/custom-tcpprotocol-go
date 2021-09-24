package main

import (
	"fmt"
	"log"

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

	// testEventsMap := map[int]enum.TestEvent{}
	tcpConnect := mtsclient.NewTCPConnect("localhost", 10001, 10000)
	defer tcpConnect.Conn.Close()
	//do all operations on top of TLS
	tcpConnect.WithTLS(nil)

	tcpConnect.ConnectAndLogin()

	// // Create some test events
	// testEventsMap[3] = enum.SendOPLPayload
	// testEventsMap[6] = enum.SendOPLPayload
	// testEventsMap[9] = enum.SendOPLPayload
	// testEventsMap[12] = enum.Exit

	// // Begin example
	// log.Println("Starting MtsClientExample in go")

	// exitTest := false
	// for !exitTest {
	// 	for !tcpConnect.MTSClient.Connected {
	// 		tcpConnect.ConnectAndLogin()
	// 		if !tcpConnect.MTSClient.Connected {
	// 			log.Println("Trying to connect")
	// 			time.Sleep(3000)
	// 		} else {
	// 			log.Println("Connected")
	// 		}
	// 	}

	// 	// Send messages within this program loop
	// 	if event, ok := testEventsMap[TestElapsedSeconds]; ok {
	// 		//do something here
	// 		switch event {
	// 		case enum.SendOPLPayload:
	// 			tcpConnect.SendTestOPLPayload(nil)
	// 			break
	// 		case enum.Exit:
	// 			exitTest = true
	// 			break
	// 		}
	// 	}

	// 	if !exitTest {
	// 		time.Sleep(1000)
	// 		TestElapsedSeconds++
	// 	}
	// }

	// fmt.Println(tcpConnect)
	log.Println("Ending MtsClientExample go")
}

// func main() {
// 	data := `	 �{"version":1,"attributeRoute":"A","route":9,"srcId":1,"dstId":2,"rpcId":1,"reply":false,"error":false,"jwt":null,"data":"AAAAAA=="}`
// 	byteArr, err := json.Marshal(data)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	mtsResponseMessage := model.MTSMessage{}
// 	err = json.Unmarshal(byteArr, &mtsResponseMessage)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }
