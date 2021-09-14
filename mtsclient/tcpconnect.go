package mtsclient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/niroopreddym/custom-tcpprotocol-go/enum"
	"github.com/niroopreddym/custom-tcpprotocol-go/model"
)

//UintSize max unsigned int size
const UintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

const (
	//MTSServer is Node ID constant for MTS Server
	MTSServer = 1
	//MTSRMSServer Node ID constant for MTS RMS Server
	MTSRMSServer = 2
	//MTSProvisioner Node ID constant for MTS Provisioner
	MTSProvisioner = 3
)

const (
	//MaxInt max signed integer value
	MaxInt = 1<<(UintSize-1) - 1 // 1<<31 - 1 or 1<<63 - 1
	//MinInt min signed int value
	MinInt = -MaxInt - 1 // -1 << 31 or -1 << 63
	//MaxUint is max unsigned int value
	MaxUint = 1<<UintSize - 1 // 1<<32 - 1 or 1<<64 - 1
)

//MaxMessageLength max chunk size that can be written to the server
const MaxMessageLength = 1 << 22 //  1 x 2 ^ 22 = 4 MB

//TCPConnect logins
type TCPConnect struct {
	MTSClient        MTSClient
	Hostname         string
	Port             int
	DefaultTimeOutMs int
	Conn             net.Conn
}

//NewTCPConnect ctor
func NewTCPConnect(hostname string, port int, defaultTimeOutMs int) *TCPConnect {
	connectionString := strings.Join([]string{hostname, strconv.Itoa(port)}, ":")
	conn := GetConnection(connectionString)

	return &TCPConnect{
		MTSClient:        MTSClient{},
		Hostname:         hostname,
		Port:             port,
		DefaultTimeOutMs: defaultTimeOutMs,
		Conn:             conn,
	}
}

//GetConnection instantiates and gets the connection
func GetConnection(connectionString string) net.Conn {
	// conn, err := net.Dial("tcp", "localhost:10001")
	// ctx, cancel := context.WithTimeout(context.Background(), 1000000*time.Millisecond)
	// defer cancel()
	conn, err := net.DialTimeout("tcp", connectionString, 10000*time.Millisecond)
	if err != nil {
		fmt.Println("error occured while establishing the connection: ", err)
		os.Exit(1)
	}

	return conn
}

//ConnectAndLogin connects and login the user
func (connect *TCPConnect) ConnectAndLogin() error {

	isAuthenticated := connect.loginWithUsernameAndPassword()
	if !isAuthenticated {
		fmt.Printf("UnAuthorized login creds")
	}

	return nil
}

func (connect *TCPConnect) loginWithUsernameAndPassword() bool {

	kAppRMS := []byte{79, 157, 102, 210, 83, 34, 156, 117, 223, 190, 187, 27, 28, 63, 94, 214, 4, 98, 123, 98, 65, 20, 143, 60, 50, 62, 162, 115, 7, 46, 119, 8}
	username := "mtstest"
	password := "Test1234"

	mtsLogin := MtsLogin{
		AppID:    enum.RMSServer,
		AppKey:   kAppRMS,
		Username: StrToPointer(username),
		Password: StrToPointer(password),
	}

	mtsLoginByteData, err := json.Marshal(mtsLogin)
	if err != nil {
		fmt.Println("error in marshalling the mtsLogin Data: ", err)
	}

	mtsLoginMessage := connect.CreateRequest(
		enum.Login,
		nil,
		MTSRMSServer,
		MTSServer,
		false,
		nil,
		mtsLoginByteData,
	)

	var mtsResult = connect.Login(mtsLoginMessage)
	fmt.Printf("Result: %v", mtsResult)
	// ClientCertificate = mtsResult.Data.ClientCertificate
	// JWT = mtsResult.Jwt
	return true
}

//Login login the user and returns the MTSResult of type MtsLoginResponse
func (connect *TCPConnect) Login(mtsLoginMessage model.MTSMessage) model.MTSResult {
	var mtsLoginResponse = connect.Send(mtsLoginMessage, 1000000)

	if mtsLoginResponse.HasError {
		fmt.Println("Invalid format exception: ", mtsLoginResponse.ErrorData)
	}

	return mtsLoginResponse
}

//Send sends the data to MTS Client
func (connect *TCPConnect) Send(mtsMessage model.MTSMessage, timeOutMs int) model.MTSResult {

	mtsMessageByteData, err := json.Marshal(mtsMessage)
	fmt.Printf("MTS msg before sending: %v", string(mtsMessageByteData))
	if err != nil {
		fmt.Println("error in marshalling the MTSMessage Data: ", err)
	}

	response := connect.send(mtsMessageByteData, timeOutMs)
	return response
}

const (
	//DELIMITER representing the end of message
	DELIMITER byte = '\n'
	//QUIT_SIGN when to close the client interaction
	QUIT_SIGN = "quit!"
)

//ReadFromConn reads from conn
func ReadFromConn(conn net.Conn, delim byte) (string, error) {
	reader := bufio.NewReader(conn)
	var buffer bytes.Buffer
	for {
		ba, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		buffer.Write(ba)
		if !isPrefix {
			break
		}
	}
	return buffer.String(), nil
}

//WriteToConn writes to connection
func WriteToConn(conn net.Conn, content []byte) (int, error) {
	writer := bufio.NewWriterSize(conn, 1<<12)
	number, err := writer.Write(content)
	if err == nil {
		err = writer.Flush()
	}
	return number, err
}

func (connect *TCPConnect) send(msg []byte, timeOutMs int) model.MTSResult {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println(err)
		}
	}()

	for {
		//sending message
		num, err := WriteToConn(connect.Conn, msg)
		if err != nil {
			log.Printf("Sender: Write Error: %s\n", err)
			break
		}
		log.Printf("Sender: Wrote %d byte(s)\n", num)

		fmt.Println("Reading the response")

		respContent, err := ReadFromConn(connect.Conn, DELIMITER)
		if err != nil {
			log.Printf("Sender: Read error: %s", err)
			break
		}
		log.Printf("Sender: Received content: %s\n", respContent)
	}

	return model.MTSResult{}
}

//CreateRequest creates a MTSMessage for user login
func (connect *TCPConnect) CreateRequest(requestType enum.MTSRequest, attrRoute *string, srcID int, dstID int, isError bool, jwt *string, data []byte) model.MTSMessage {
	var rpcID int
	var lastRPCID int
	//if OPL request, set RpcId to 0 and don't increment LastRpcId
	if requestType == enum.OPL {
		rpcID = 0
	} else {
		if lastRPCID == MaxInt {
			lastRPCID = 1
		} else {
			lastRPCID = lastRPCID + 1
		}

		rpcID = lastRPCID
	}

	mtsMessage := model.MTSMessage{
		Route:   requestType,
		SrcID:   srcID,
		DstID:   dstID,
		RPCID:   rpcID,
		Reply:   false,
		IsError: isError,
		Data:    data,
	}

	if attrRoute != nil {
		mtsMessage.AttributeRoute = *attrRoute
	}

	if attrRoute != nil {
		mtsMessage.JWT = *jwt
	}

	return mtsMessage
}

//StrToPointer converts the string to pointer
func StrToPointer(str string) *string {
	return &str
}
