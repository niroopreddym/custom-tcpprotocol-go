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
)

//Offset is the length of byte array to indicate the msg length
const Offset int = helper.Offset

var (
	//ClientCertificate has the client cert
	ClientCertificate []byte
	//JWT has the JWT data
	JWT []byte
	//KAppRMS rms key
	KAppRMS = []byte{79, 157, 102, 210, 83, 34, 156, 117, 223, 190, 187, 27, 28, 63, 94, 214, 4, 98, 123, 98, 65, 20, 143, 60, 50, 62, 162, 115, 7, 46, 119, 8}
)

//TCPConnect logins
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
}

//NewTCPConnect ctor
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
	}
}

//WithTLS connects with TLS
func (connect *TCPConnect) WithTLS(certificate []byte) {
	connect.MTSClient.UseTLS = true
	connect.MTSClient.ClientCertificate = certificate
}

//TCPServer returns the TCP server connection
func (connect *TCPConnect) TCPServer(ClientCertificate []byte, errorChan chan error, authenticationCall func(errorChan chan error)) {
	connectionString := strings.Join([]string{connect.Hostname, strconv.Itoa(connect.Port)}, ":")
	conn, err := helper.GetConnection(connectionString)
	if err != nil {
		fmt.Println("error while instantiating the connection", err)
		os.Exit(1)
	}

	connect.MTSClient.Connected = true
	connect.Conn = conn

	authenticationCall(errorChan)
}

//ConnectAndLogin connects and login the user
func (connect *TCPConnect) ConnectAndLogin(isAuthenticated chan bool, errorChan chan error) {
	go connect.TCPServer(nil, errorChan, connect.loginWithUsernameAndPassword)

	if <-connect.IsAuthenticated {
		connect.Conn.Close()
	} else {
		fmt.Println("UnAuthorized login creds")
		go func() { errorChan <- fmt.Errorf("UnAuthorized login creds") }()
	}

	connect.WithTLS(ClientCertificate)
	go connect.TCPServer(nil, errorChan, connect.loginWithCertificate)

	if <-connect.IsAuthenticated {
		fmt.Println("successfully booted up the server with the client certificate")
		go func() { connect.ServerBootDone <- true }()
	}
}

func (connect *TCPConnect) loginWithCertificate(errorChan chan error) {

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
		go func() { errorChan <- err }()

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

	connect.Login(mtsLoginMessage, connect.IsAuthenticated, errorChan)
}

func (connect *TCPConnect) loginWithUsernameAndPassword(errorChan chan error) {
	mtsLogin := model.MtsLogin{
		AppID:    enum.RMSServer,
		AppKey:   KAppRMS,
		Username: connect.UserName,
		Password: connect.Password,
	}

	mtsLoginByteData, err := json.Marshal(mtsLogin)
	if err != nil {
		fmt.Println("error in marshalling the mtsLogin Data: ", err)
		go func() { errorChan <- err }()

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

	connect.Login(mtsLoginMessage, connect.IsAuthenticated, errorChan)
}

//SendAcknowledgmentToServer sends the ack back to the server
func (connect *TCPConnect) SendAcknowledgmentToServer(mtsMessage *model.MTSMessage) error {
	switch mtsMessage.Route {
	case enum.OPL:
		// TODO: Route OPL Messages to nodes
		//break
		log.Println("OPL response")
		strOplResponse, _ := json.Marshal(mtsMessage)
		log.Println("received OPL response: ", strOplResponse)
		return nil
	case enum.RMSPing:
		log.Print("Received RMS Ping request. Sending RMS Ping response.")
		fmt.Println("JWT Token is : ", mtsMessage.JWT)
		msgResponse := helper.CreateResponse(mtsMessage, enum.RMSPingResponse, nil, false, mtsMessage.JWT, make([]byte, 4))
		return connect.Send(msgResponse)
	default:
		log.Print("Unknown message : ", mtsMessage.Route)
		var responseMsg = helper.CreateErrorResponse(enum.InvalidRequest, enum.MtsErrorID.String(enum.InvalidRequest), mtsMessage, mtsMessage.Route, nil, mtsMessage.JWT)
		return connect.Send(responseMsg)
	}
}

//Login login the user and returns the MtsLoginResponse
func (connect *TCPConnect) Login(mtsLoginMessage model.MTSMessage, certificateReceived chan bool, errorChan chan error) {
	err := connect.SendLoginPayload(mtsLoginMessage, 1000000)

	if err != nil {
		fmt.Println("Error getting the client cert", err)
		go func() { errorChan <- err }()

	}

	isDone, err := connect.Receieve()
	if err != nil {
		fmt.Println("error reading the data")
		go func() { errorChan <- err }()
	}

	if isDone {
		go func() { certificateReceived <- true }()
	}
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

//Send sends the data to MTS Client
func (connect *TCPConnect) Send(mtsMessage model.MTSMessage) error {

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

	fmt.Println("rquired response size: ", responseLengthInt)

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

	// data := <-dataSegmentString

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
		go func() { connect.IsAuthenticated <- false }()
	}

	if mtsMessage.IsError == true {
		go func() { connect.IsAuthenticated <- false }()
	}

	ClientCertificate = mtsResponse.ClientCertificate
	if mtsMessage.JWT != nil {
		JWT = []byte(*mtsMessage.JWT)
	} else {
		JWT = nil
	}

	go func() { connect.IsAuthenticated <- true }()
}

//WriteToConn writes to connection
func (connect *TCPConnect) WriteToConn(conn net.Conn, content []byte) (int, error) {
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

	connect.SendOPLPayload(mtsLoginMessage)
}

//SendOPLPayload Send an OPL Message
func (connect *TCPConnect) SendOPLPayload(mtsOPLMessage model.MTSMessage) {
	connect.Send(mtsOPLMessage)
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
	num, err := connect.WriteToConn(connect.Conn, data)
	if err != nil {
		return fmt.Errorf("Sender: Write Error: %w", err)
	}

	log.Printf("Sender: Wrote %d byte(s)\n", num)
	return nil
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
