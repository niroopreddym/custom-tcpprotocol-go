package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/niroopreddym/custom-tcpprotocol-go/enum"
	helper "github.com/niroopreddym/custom-tcpprotocol-go/helpers"
	"github.com/niroopreddym/custom-tcpprotocol-go/mtsclient"
)

//TestElapsedSeconds intial time to trigger the fan out test events
var TestElapsedSeconds = 1

//get these values securely from the env
const username = "mtstest"
const password = "Test123"

func main() {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println(err)
		}
	}()

	fmt.Printf("Please press ctrl + c to stop the connection: ")
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	tcpConnect := mtsclient.NewTCPConnect("127.0.0.1", 10001, 10000)
	tcpConnect.UserName = helper.StrToPointer(username)
	tcpConnect.Password = helper.StrToPointer(password)

	defer tcpConnect.Conn.Close()
	//do all operations on top of TLS
	tcpConnect.WithTLS(nil)

	tcpConnect.ConnectAndLogin()

	for {
		select {
		case isDone := <-tcpConnect.ServerBootDone:
			if !isDone {
				fmt.Println("Un-Authenticated")
				os.Exit(1)
			}

			tcpConnect.Wg.Add(1)
			go sendOPLTestMessages(tcpConnect, &tcpConnect.Wg)
			tcpConnect.Wg.Wait()
		case msg := <-tcpConnect.ErrorChan:
			fmt.Println("BOOM!", msg.Error())
			if strings.Contains(msg.Error(), "use of closed network connection") {
				time.Sleep(500 * time.Millisecond) // has to be exponential backoff mechanism
			} else {
				fmt.Println(msg)
				os.Exit(1)
			}
		case <-done:
			return
		default:
			fmt.Println("    .")
			time.Sleep(500 * time.Millisecond)
		}
	}
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
