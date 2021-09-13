package mtsclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

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

	conn, err := net.Dial("tcp", connectionString)
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
	password := "Test123"

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
	if err != nil {
		fmt.Println("error in marshalling the MTSMessage Data: ", err)
	}

	response := connect.send(mtsMessageByteData, timeOutMs)
	return response
}

func (connect *TCPConnect) send(msg []byte, timeOutMs int) model.MTSResult {
	//find a way to prepare the data
	// var data = PrepareData(msg)

	//open the connection and send the data here

	for {
		_, err := io.Writer.Write(connect.Conn, msg)
		if err != nil {
			fmt.Println("some error while writing the data to the connection : ", err)
			// panic(err)
			break
		}

		log.Printf("Send: %s", msg)

		buff := make([]byte, 1024)
		n, err := connect.Conn.Read(buff)
		if err != nil {
			fmt.Println("some error while writing the data to the connection : ", err)
			// panic(err)
			break
		}

		log.Printf("Receive: %s", buff[:n])
	}

	return model.MTSResult{}
}

//MaxMessageLength max chunk size that can be written to the server
const MaxMessageLength = 1 << 22 //  1 x 2 ^ 22 = 4 MB

// func PrepareData() {

// }

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

		// LastRpcId = LastRpcId == int.MaxValue ? 1 : LastRpcId + 1;
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

// //Send senfs the message to client
// func (login *TCPConnect) Send(mtsMessage model.MTSMessage) {
// 	fmt.Println(mtsMessage)
// 	mutex := sync.Mutex{}
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	data, err := json.Marshal(mtsMessage)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	Send(data)
// }

// func send(byte[] msg){
// 	var data = PrepareData(msg)
// 	fmt.Printf("Sending message, data size %d", len(data))

// 	if (Stream.CanWrite)
// 		Stream.BeginWrite(data, 0, data.Length, WriteComplete, this)
// 	else
// 	{
// 		Log?.LogWarning($"{LoggerMods.CallerInfo()}Cannot write to stream")
// 		throw new MTSConnectionClosedException("Cannot write to stream")
// 	}
// }
