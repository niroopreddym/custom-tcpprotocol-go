package mtsclient

import (
	"bufio"
	"crypto/tls"
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
	conn, err := GetConnection(connectionString)
	if err != nil {
		os.Exit(1)
	}

	return &TCPConnect{
		MTSClient: MTSClient{
			Connected: true,
		},
		Hostname:         hostname,
		Port:             port,
		DefaultTimeOutMs: defaultTimeOutMs,
		Conn:             conn,
	}
}

//GetConnection instantiates and gets the connection
func GetConnection(connectionString string) (net.Conn, error) {
	tlsConfig := &tls.Config{}

	tcp, err := net.Dial("tcp", connectionString)
	if err != nil {
		log.Fatal("error whwn dialing the connection ", err)
	}

	serverCert, err := tlsConfig.GetCertificate(&tls.ClientHelloInfo{
		ServerName: "onity.net",
		Conn:       tcp,
	})

	if err != nil {
		log.Fatal("cannot retrive the server cert ", err)
	}

	fmt.Println(serverCert)

	conn, err := tls.Dial("tcp", connectionString, tlsConfig)
	if err != nil {
		fmt.Println("error occured while establishing the connection: ", err)
		return nil, err
	}

	return conn, err
}

//ConnectComplete runs the stream logic after the connection is successfull
func (connect *TCPConnect) ConnectComplete() {
	connect.StartReader()
}

//StartReader starts reading the stream via tcp connection
func (connect *TCPConnect) StartReader() {
	if connect.MTSClient.UseTLS {
		StartTLS()
	}
}

//StartTLS validate the certificate with client hostname
func StartTLS() {

}

//WithTLS connects with TLS
func (connect *TCPConnect) WithTLS(certificate []byte) {
	connect.MTSClient.UseTLS = true
	connect.MTSClient.ClientCertificate = certificate
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
)

// //ReadFromConn reads from conn
// func ReadFromConn(conn net.Conn, delim byte) (string, error) {
// 	reader := bufio.NewReader(conn)
// 	var buffer bytes.Buffer
// 	for {
// 		ba, isPrefix, err := reader.ReadLine()
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			return "", err
// 		}
// 		buffer.Write(ba)
// 		if !isPrefix {
// 			break
// 		}
// 	}
// 	return buffer.String(), nil
// }

//ReadFromConn reads from conn
func ReadFromConn(conn net.Conn, delim byte) (string, error) {

	// useful block
	buff := make([]byte, 50)
	c := bufio.NewReader(conn)

	for {
		// read a single byte which contains the message length
		size, err := c.ReadByte()
		if err != nil {
			return "", err
		}

		// read the full message, or return an error
		_, err = io.ReadFull(c, buff[:int(size)])
		if err != nil {
			return "", err
		}

		fmt.Printf("received %x\n", buff[:int(size)])
	}

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

//SendTestOPLPayload sends the test payload to the server
func (connect *TCPConnect) SendTestOPLPayload(jwt *string) {
	mtsOPLPayload := model.MtsOplPayload{
		RoomID:          "101",
		ProxyMACAddress: "",
		Data:            make([]byte, 16),
	}

	strMtsOPLPayload, _ := json.Marshal(mtsOPLPayload)

	mtsLoginMessage := connect.CreateRequest(
		enum.OPL,
		nil,
		MTSRMSServer,
		MTSServer,
		false,
		nil,
		strMtsOPLPayload,
	)

	connect.SendOPLPayload(mtsLoginMessage)
}

//SendOPLPayload Send an OPL Message
func (connect *TCPConnect) SendOPLPayload(mtsOPLMessage model.MTSMessage) {
	connect.Send(mtsOPLMessage, 1000000)
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
