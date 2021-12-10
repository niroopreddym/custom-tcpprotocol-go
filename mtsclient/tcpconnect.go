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
	"sync"

	"github.com/niroopreddym/custom-tcpprotocol-go/enum"
	helper "github.com/niroopreddym/custom-tcpprotocol-go/helpers"
	"github.com/niroopreddym/custom-tcpprotocol-go/model"
)

const (
	//MTSServer is Node ID constant for MTS Server
	MTSServer = 1
	//MTSRMSServer Node ID constant for MTS RMS Server
	MTSRMSServer = 2
	//MTSProvisioner Node ID constant for MTS Provisioner
	MTSProvisioner = 3
	//Offset is the length of byte array to indicate the msg length
	Offset = helper.Offset
	//WriteBufferSize is the size of the buffer that writes to the server
	WriteBufferSize = 1 << 12
)

var (
	//ClientCertificate has the client cert
	ClientCertificate []byte
	//JWT has the JWT data
	JWT []byte
	//KAppRMS rms key
	KAppRMS = []byte{79, 157, 102, 210, 83, 34, 156, 117, 223, 190, 187, 27, 28, 63, 94, 214, 4, 98, 123, 98, 65, 20, 143, 60, 50, 62, 162, 115, 7, 46, 119, 8}
)

//TCPConnect is the struct that implements the socket communication functionality
type TCPConnect struct {
	MTSClient        model.MTSClient
	Hostname         string
	Port             int
	DefaultTimeOutMs int
	Conn             net.Conn
	ServerBootDone   chan bool
	Wg               sync.WaitGroup
	IsAuthenticated  chan bool
	UserName         *string
	Password         *string
	ErrorChan        chan error
}

//NewTCPConnect is the ctor that instantiates the struct
func NewTCPConnect(hostname string, port int, defaultTimeOutMs int) *TCPConnect {
	return &TCPConnect{
		MTSClient: model.MTSClient{
			Connected: false,
		},
		Hostname:         hostname,
		Port:             port,
		DefaultTimeOutMs: defaultTimeOutMs,
		Conn:             &tls.Conn{},
		ServerBootDone:   make(chan bool),
		Wg:               sync.WaitGroup{},
		IsAuthenticated:  make(chan bool),
		ErrorChan:        make(chan error),
	}
}

//WithTLS connects with TLS
func (connect *TCPConnect) WithTLS(certificate []byte) {
	connect.MTSClient.UseTLS = true
	connect.MTSClient.ClientCertificate = certificate
}

//TCPServer returns the TCP server connection
func (connect *TCPConnect) TCPServer(ClientCertificate []byte, authenticationCall func()) {
	connectionString := strings.Join([]string{connect.Hostname, strconv.Itoa(connect.Port)}, ":")
	conn, err := helper.GetConnection(connectionString)
	if err != nil {
		fmt.Println("error while instantiating the connection", err)
		os.Exit(1)
	}

	connect.MTSClient.Connected = true
	connect.Conn = conn

	authenticationCall()
}

//ConnectAndLogin connects and login the user
func (connect *TCPConnect) ConnectAndLogin() {
	go connect.TCPServer(nil, connect.loginWithUsernameAndPassword)

	if <-connect.IsAuthenticated {
		connect.Conn.Close()
	} else {
		fmt.Println("UnAuthorized login creds")
		go func() { connect.ErrorChan <- fmt.Errorf("UnAuthorized login creds") }()
	}

	connect.WithTLS(ClientCertificate)
	go connect.TCPServer(nil, connect.loginWithCertificate)

	if <-connect.IsAuthenticated {
		fmt.Println("successfully booted up the server with the client certificate")
		go func() { connect.ServerBootDone <- true }()
	}
}

func (connect *TCPConnect) loginWithCertificate() {

	mtsLogin := model.MtsLogin{
		AppID:             enum.RMSServer,
		AppKey:            KAppRMS,
		ClientCertificate: ClientCertificate,
		Username:          nil,
		Password:          nil,
	}

	mtsLoginByteData, err := json.Marshal(mtsLogin)
	if err != nil {
		fmt.Println("error in marshalling the mtsLogin Data: ", err)
		go func() { connect.ErrorChan <- err }()

	}

	mtsLoginMessage := helper.CreateRequest(
		enum.Login,
		nil,
		MTSRMSServer,
		MTSServer,
		false,
		nil,
		mtsLoginByteData,
	)

	connect.Login(mtsLoginMessage, connect.IsAuthenticated)
}

func (connect *TCPConnect) loginWithUsernameAndPassword() {
	mtsLogin := model.MtsLogin{
		AppID:    enum.RMSServer,
		AppKey:   KAppRMS,
		Username: connect.UserName,
		Password: connect.Password,
	}

	mtsLoginByteData, err := json.Marshal(mtsLogin)
	if err != nil {
		fmt.Println("error in marshalling the mtsLogin Data: ", err)
		go func() { connect.ErrorChan <- err }()

	}

	mtsLoginMessage := helper.CreateRequest(
		enum.Login,
		nil,
		MTSRMSServer,
		MTSServer,
		false,
		nil,
		mtsLoginByteData,
	)

	connect.Login(mtsLoginMessage, connect.IsAuthenticated)
}

//Login login the user and returns the MtsLoginResponse
func (connect *TCPConnect) Login(mtsLoginMessage model.MTSMessage, certificateReceived chan bool) {
	err := connect.SendLoginPayload(mtsLoginMessage, 1000000)

	if err != nil {
		fmt.Println("Error getting the client cert", err)
		connect.ErrorChan <- err
	}

	isDone, err := connect.Receieve()
	if err != nil {
		fmt.Println("error reading the data")
		connect.ErrorChan <- err
	}

	if isDone {
		certificateReceived <- true
	}
}

//SendAcknowledgmentToServer sends the ack back to the server
func (connect *TCPConnect) SendAcknowledgmentToServer(mtsMessage *model.MTSMessage) error {
	strJWT := string(JWT)
	mtsMessage.JWT = &strJWT

	switch mtsMessage.Route {
	case enum.OPL:
		log.Println("OPL response")
		strOplResponse, _ := json.Marshal(mtsMessage)
		log.Println("received OPL response: ", strOplResponse)
		return nil
	case enum.RMSPing:
		log.Print("Received RMS Ping request. Sending RMS Ping response.")
		fmt.Println("JWT Token is : ", mtsMessage.JWT)
		msgResponse := helper.CreateResponse(mtsMessage, enum.RMSPingResponse, nil, false, mtsMessage.JWT, make([]byte, 4))
		return connect.SendDataToServer(msgResponse)
	default:
		log.Print("Unknown message : ", mtsMessage.Route)
		var responseMsg = helper.CreateErrorResponse(enum.InvalidRequest, enum.MtsErrorID.String(enum.InvalidRequest), mtsMessage, mtsMessage.Route, nil, mtsMessage.JWT)
		return connect.SendDataToServer(responseMsg)
	}
}

//SendDataToServer sends the data to MTS Server
func (connect *TCPConnect) SendDataToServer(mtsMessage model.MTSMessage) error {

	mtsMessageByteData, err := json.Marshal(mtsMessage)
	if err != nil {
		fmt.Println("error in marshalling the MTSMessage Data: ", err)
	}

	err = connect.send(mtsMessageByteData)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (connect *TCPConnect) send(msg []byte) error {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println(err)
		}
	}()

	data := helper.PrepareData(msg)
	//sending message
	fmt.Println("sender payload json:", string(msg))

	fmt.Println("byte array : ", data)
	num, err := connect.WriteToConn(data)
	if err != nil {
		return fmt.Errorf("Sender: Write Error: %w", err)
	}

	log.Printf("Sender: Wrote %d byte(s)\n", num)
	return nil
}

//WriteToConn writes to connection
func (connect *TCPConnect) WriteToConn(content []byte) (int, error) {
	writer := bufio.NewWriterSize(connect.Conn, WriteBufferSize)
	number, err := writer.Write(content)
	if err == nil {
		err = writer.Flush()
	}
	return number, err
}

// ReadFromConn reads from conn
func (connect *TCPConnect) ReadFromConn() (bool, error) {
	size := 100
	buff := make([]byte, size)
	reader := bufio.NewReader(connect.Conn)

	finalstring := ""
	for {
		_, err := io.ReadFull(reader, buff[:size])
		if err != nil {
			fmt.Println("error occured while reading form the connection")
			return false, err
		}

		finalstring = finalstring + string(buff)
		finalstring = connect.validateServerResponseLength([]byte(finalstring))
	}
}

func (connect *TCPConnect) validateServerResponseLength(buff []byte) string {
	var lengthOfResponseBa []byte
	lengthOfResponseBa = buff[:4]

	// convert This length Of Response To data length
	responseLengthInt := helper.ConvertByteToInt(lengthOfResponseBa)

	if len(buff) < responseLengthInt+Offset {
		return string(buff)
	}

	dataSegment := buff[Offset : responseLengthInt+Offset]
	fmt.Println("required segement : ", string(dataSegment))
	connect.ProcessDataSegment(string(dataSegment))

	remainingData := buff[responseLengthInt+Offset:]
	fmt.Println("remaining segment : ", string(remainingData))
	return string(remainingData)
}

//ProcessDataSegment process this segment asynchronously
func (connect *TCPConnect) ProcessDataSegment(dataSegmentString string) {
	mtsResponseMessage := model.MTSMessage{}
	err := json.Unmarshal([]byte(dataSegmentString), &mtsResponseMessage)
	if err != nil {
		fmt.Println("error occured while unmarshalling datasegment: ", err)
	}

	switch mtsResponseMessage.Route {
	case enum.OPL:
		fmt.Println("OPL response: ")
		connect.SendAcknowledgmentToServer(&mtsResponseMessage)
	case enum.LoginResponse:
		connect.ExtractCertData(mtsResponseMessage)
	case enum.RMSPing:
		connect.SendAcknowledgmentToServer(&mtsResponseMessage)
	}
}

//ExtractCertData extracts the cert information out of the response
func (connect *TCPConnect) ExtractCertData(mtsMessage model.MTSMessage) {
	mtsResponse := model.MtsLoginResponse{}
	responseData := mtsMessage.Data
	err := json.Unmarshal(responseData, &mtsResponse)
	if err != nil {
		fmt.Println("error occuredwhen unmarshalling the response data")
		connect.IsAuthenticated <- false
	}

	if mtsMessage.IsError == true {
		connect.IsAuthenticated <- false
	}

	ClientCertificate = mtsResponse.ClientCertificate
	if mtsMessage.JWT != nil {
		JWT = []byte(*mtsMessage.JWT)
	} else {
		JWT = nil
	}

	connect.IsAuthenticated <- true
}

//Receieve receives the response
func (connect *TCPConnect) Receieve() (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println(err)
		}
	}()

	isDone, err := connect.ReadFromConn()
	return isDone, err
}

//SendLoginPayload sends the data to MTS Client and gets the appropriate response
func (connect *TCPConnect) SendLoginPayload(mtsMessage model.MTSMessage, timeOutMs int) error {
	mtsMessageByteData, err := json.Marshal(mtsMessage)
	if err != nil {
		fmt.Println("error in marshalling the MTSMessage Data: ", err)
		return err
	}

	err = connect.send(mtsMessageByteData)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

//SendTestOPLPayload sends the test payload to the server
func (connect *TCPConnect) SendTestOPLPayload() {
	//101 is the operating RoomID
	mtsOPLPayload := model.MtsOplPayload{
		RoomID:          "101",
		ProxyMACAddress: nil,
		Data:            make([]byte, 16),
	}

	strMtsOPLPayload, err := json.Marshal(mtsOPLPayload)
	if err != nil {
		fmt.Println("error occured in marshalling the OPLpayload data")
		fmt.Println(err)
	}

	mtsLoginMessage := helper.CreateRequest(
		enum.OPL,
		nil,
		MTSRMSServer,
		MTSServer,
		false,
		helper.StrToPointer(string(JWT)),
		strMtsOPLPayload,
	)

	connect.SendDataToServer(mtsLoginMessage)
}

//SendMTSOPLPayload sends the OPL payload to the server
func (connect *TCPConnect) SendMTSOPLPayload(mtsOPLPayload *model.MtsOplPayload) {
	strMtsOPLPayload, err := json.Marshal(mtsOPLPayload)
	if err != nil {
		fmt.Println("error occured in marshalling the OPLpayload data")
		fmt.Println(err)
	}

	mtsLoginMessage := helper.CreateRequest(
		enum.OPL,
		nil,
		MTSRMSServer,
		MTSServer,
		false,
		helper.StrToPointer(string(JWT)),
		strMtsOPLPayload,
	)

	connect.SendDataToServer(mtsLoginMessage)
	connect.Wg.Done()
}
