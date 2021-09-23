package mtsclient

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
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
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println(err)
		}
	}()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		// RootCAs:            x509.NewCertPool(),
		// Renegotiation:      tls.RenegotiateFreelyAsClient,
	}

	conn, err := tls.Dial("tcp", connectionString, tlsConfig)
	if err != nil {
		fmt.Println("error occured while establishing the connection: ", err)
		return nil, err
	}

	return conn, err
}

//GetConnectionClientCert instantiates and gets the connection with certificate
func GetConnectionClientCert(connectionString string, cert []byte) (net.Conn, error) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println(err)
		}
	}()

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(cert) {
		return nil, fmt.Errorf("failed to add client cert to cert pool")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            certPool,
		Renegotiation:      tls.RenegotiateFreelyAsClient,
	}

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

	// conn, err := tls.Dial("tcp", "onity.net:443", nil)
	// if err != nil {
	// 	fmt.Println("Server doesn't support SSL certificate err: " + err.Error())
	// }

	// err = conn.VerifyHostname("onity.net")
	// if err != nil {
	// 	panic("Hostname doesn't match with certificate: " + err.Error())
	// }
}

//WithTLS connects with TLS
func (connect *TCPConnect) WithTLS(certificate []byte) {
	connect.MTSClient.UseTLS = true
	connect.MTSClient.ClientCertificate = certificate
}

//ConnectAndLogin connects and login the user
func (connect *TCPConnect) ConnectAndLogin() error {

	mtsCertResponse, isAuthenticated := connect.loginWithUsernameAndPassword()
	if !isAuthenticated {
		fmt.Println("UnAuthorized login creds")
		return fmt.Errorf("UnAuthorized login creds")
	}

	connect.Conn.Close()

	connect.WithTLS(mtsCertResponse.Data)

	connectionString := strings.Join([]string{connect.Hostname, strconv.Itoa(connect.Port)}, ":")
	conn, err := GetConnectionClientCert(connectionString, connect.MTSClient.ClientCertificate)
	if err != nil {
		os.Exit(1)
	}

	connect.Conn = conn

	mtsJWTResponse, isAuthenticated := connect.loginWithCertificate()
	if !isAuthenticated {
		return fmt.Errorf("UnAuthorized client cert")
	}

	fmt.Println(mtsJWTResponse)
	return nil
}

func (connect *TCPConnect) loginWithCertificate() (*model.MTSMessage, bool) {
	kAppRMS := []byte{79, 157, 102, 210, 83, 34, 156, 117, 223, 190, 187, 27, 28, 63, 94, 214, 4, 98, 123, 98, 65, 20, 143, 60, 50, 62, 162, 115, 7, 46, 119, 8}

	mtsLogin := MtsLogin{
		AppID:             enum.RMSServer,
		AppKey:            kAppRMS,
		ClientCertificate: connect.MTSClient.ClientCertificate,
		Username:          nil,
		Password:          nil,
	}

	mtsLoginByteData, err := json.Marshal(mtsLogin)
	if err != nil {
		fmt.Println("error in marshalling the mtsLogin Data: ", err)
		return nil, false
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

	mtsIntialResult, err := connect.LoginGetCert(mtsLoginMessage)
	if err != nil {
		return nil, false
	}

	// ClientCertificate = mtsResult.Data.ClientCertificate
	// JWT = mtsResult.Jwt

	return mtsIntialResult, true
}

func (connect *TCPConnect) loginWithUsernameAndPassword() (*model.MTSMessage, bool) {

	kAppRMS := []byte{79, 157, 102, 210, 83, 34, 156, 117, 223, 190, 187, 27, 28, 63, 94, 214, 4, 98, 123, 98, 65, 20, 143, 60, 50, 62, 162, 115, 7, 46, 119, 8}
	username := "mtstest"
	password := "Test123"

	mtsLogin := MtsLogin{
		AppID:    enum.RMSServer,
		AppKey:   kAppRMS,
		Username: &username,
		Password: &password,
	}

	mtsLoginByteData, err := json.Marshal(mtsLogin)
	if err != nil {
		fmt.Println("error in marshalling the mtsLogin Data: ", err)
		return nil, false
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

	mtsIntialResult, err := connect.LoginGetCert(mtsLoginMessage)
	if err != nil {
		fmt.Println(err)
		// return nil, false
	}

	trimmedResponse := `{"version":1,"attributeRoute":"A","route":9,"srcId":1,"dstId":2,"rpcId":1,"reply":false,"error":false,"jwt":null,"data":"AAAAAA=="}`

	log.Printf("Sender: Received content: %s\n", trimmedResponse)
	mtsResponseMessage := model.MTSMessage{}

	err = json.Unmarshal([]byte(trimmedResponse), &mtsResponseMessage)
	if err != nil {
		fmt.Println("unmarshal resposne error : ", err)
		return nil, false
	}

	fmt.Printf("Result: %v", mtsIntialResult)

	ackResponse, err := connect.SendAcknowledgmentToServer(&mtsResponseMessage)
	// ackResponse, err := connect.SendAcknowledgmentToServer(mtsIntialResult)
	fmt.Println(err)
	fmt.Println(ackResponse)
	// ClientCertificate = mtsResult.Data.ClientCertificate
	// JWT = mtsResult.Jwt

	// sendthePongResponse to the server

	return mtsIntialResult, true
}

//SendAcknowledgmentToServer sends the ack back to the server
func (connect *TCPConnect) SendAcknowledgmentToServer(mtsMessage *model.MTSMessage) (*model.MTSMessage, error) {
	// var value string
	// if mtsMessage.jwt == "jwt" {
	// 	value = mtsMessage.jwt
	// } else {
	// 	mtsMessage.jwt = nil
	// }

	// var JWT = mtsMessage.jwt == null || mtsMessage.jwt == JWT || value
	// JWT := nil
	switch mtsMessage.Route {
	case enum.OPL:
		// TODO: Route OPL Messages to nodes
		//break
		return nil, nil
	case enum.RMSPing:
		log.Print("Received RMS Ping request. Sending RMS Ping response.")
		msgResponse := connect.CreateResponse(mtsMessage, enum.RMSPingResponse, nil, false, mtsMessage.JWT, make([]byte, 4))
		return connect.Send(msgResponse, 1000000)
	default:
		log.Print("Unknown message : ", mtsMessage.Route)
		var responseMsg = connect.CreateErrorResponse(enum.InvalidRequest, enum.MtsErrorID.String(enum.InvalidRequest), mtsMessage, mtsMessage.Route, nil, mtsMessage.JWT)
		return connect.Send(responseMsg, 1000000)
	}
}

//CreateErrorResponse error response
func (connect *TCPConnect) CreateErrorResponse(errorID enum.MtsErrorID, errorMsg string, requestMessage *model.MTSMessage, responseType enum.MTSRequest, attrRoute *string, jwt *string) model.MTSMessage {
	var errorResponseData = model.MtsErrorResponse{
		MtsError:        errorID,
		MtsErrorMessage: errorMsg,
	}

	errorResponseByteArray, err := json.Marshal(errorResponseData)
	if err != nil {
		fmt.Println("Error getting the client cert", err)
	}

	return connect.CreateResponse(requestMessage, responseType, attrRoute, true, jwt, errorResponseByteArray)
}

//CreateResponse creates a MTSMessage as Response
func (connect *TCPConnect) CreateResponse(requestMessage *model.MTSMessage, responseType enum.MTSRequest, attrRoute *string, isError bool, jwt *string, data []byte) model.MTSMessage {
	mtsMessage := model.MTSMessage{
		Version:        1,
		Route:          responseType,
		SrcID:          requestMessage.SrcID,
		DstID:          requestMessage.DstID,
		RPCID:          requestMessage.RPCID,
		Reply:          true,
		IsError:        isError,
		Data:           data,
		AttributeRoute: attrRoute,
		JWT:            jwt,
	}

	return mtsMessage
}

//  private MTSMessage CreateResponse(MTSMessage requestMessage, MTSRequest responseType, string attrRoute, bool error, string jwt, byte[] data)
//         {
//             return new MTSMessage(responseType, attrRoute, requestMessage.dstId, requestMessage.srcId, requestMessage.rpcId, true, error,
//                 jwt, data);
//         }

//LoginGetCert login the user and returns the MTSResult of type MtsLoginResponse
func (connect *TCPConnect) LoginGetCert(mtsLoginMessage model.MTSMessage) (*model.MTSMessage, error) {
	mtsLoginResponse, err := connect.GetCertificate(mtsLoginMessage, 1000000)

	if err != nil {
		fmt.Println("Error getting the client cert", err)
		return nil, fmt.Errorf("Error getting the client cert")
	}

	return mtsLoginResponse, nil
}

//GetCertificate sends the data to MTS Client to get intial cert
func (connect *TCPConnect) GetCertificate(mtsMessage model.MTSMessage, timeOutMs int) (*model.MTSMessage, error) {
	mtsMessageByteData, err := json.Marshal(mtsMessage)
	if err != nil {
		fmt.Println("error in marshalling the MTSMessage Data: ", err)
		return nil, err
	}

	err = connect.send(mtsMessageByteData, timeOutMs)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	response, err := connect.Receieve()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return response, nil
}

//Send sends the data to MTS Client
func (connect *TCPConnect) Send(mtsMessage model.MTSMessage, timeOutMs int) (*model.MTSMessage, error) {

	mtsMessageByteData, err := json.Marshal(mtsMessage)
	if err != nil {
		fmt.Println("error in marshalling the MTSMessage Data: ", err)
	}

	err = connect.send(mtsMessageByteData, timeOutMs)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	response, err := connect.Receieve()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	//first resposne will give you cert data
	return response, nil
}

const (
	//DELIMITER representing the end of message
	DELIMITER byte = '\n'
)

//ReadFromConn reads from conn
func ReadFromConn(conn net.Conn, delim byte) ([]byte, error) {
	reader := bufio.NewReader(conn)
	var buffer bytes.Buffer

	for {
		ba, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				if buffer.Len() == 0 {
					fmt.Println("conn closed....", err)
					return nil, err
				}

				break
			}

			return nil, err
		}

		buffer.Write(ba)

		fmt.Println(buffer.String())
		if !isPrefix {
			break
		}
	}

	return buffer.Bytes(), nil
	// return standardizeSpaces(buffer.String()), nil
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
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

func (connect *TCPConnect) send(msg []byte, timeOutMs int) error {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println(err)
		}
	}()

	data := PrepareData(msg)
	//sending message
	fmt.Println("json:", string(msg))

	fmt.Println("byte array : ", data)
	num, err := WriteToConn(connect.Conn, data)
	if err != nil {
		return fmt.Errorf("Sender: Write Error: %w", err)
	}

	log.Printf("Sender: Wrote %d byte(s)\n", num)
	return nil
}

//MAXMESSAGELENGTH is the length limitation for the message to be sent to the server
const MAXMESSAGELENGTH = 1 << 22 // 2^2 * 2^10 * 2^10 = 4 MiB

//PrepareData prepares the data to be sent
func PrepareData(msg []byte) []byte {
	// Expected length should be uint (TA8319)
	msgLength := len(msg)
	log.Println("Send message length ", msgLength)

	if msgLength > MAXMESSAGELENGTH {
		var errorMsg = fmt.Sprintf("Messages longer than %d are not supported.", MAXMESSAGELENGTH)
		log.Println(errorMsg)
		log.Panic("msglength exception")
	}

	//get bytes fro msgLength for now hard code to 4 i.e., BitConverter.GetBytes(msgLength)
	byteArrayLen := convertIntToByte(int32(msgLength))
	data := make([]byte, len(byteArrayLen)+msgLength)

	copy(data, byteArrayLen)
	copy(data[len(byteArrayLen):], msg)

	return data
}

func convertIntToByte(msgLength int32) []byte {
	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(msgLength))
	return buff
}

//Receieve receives the response
func (connect *TCPConnect) Receieve() (*model.MTSMessage, error) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			fmt.Println(err)
		}
	}()

	mtsResponseMessage := model.MTSMessage{}

	for {
		fmt.Println("Reading the response")

		respContent, err := ReadFromConn(connect.Conn, DELIMITER)
		if err != nil {
			// connect.Conn.Close()
			fmt.Println(err)
			return nil, fmt.Errorf("Sender: Read error: %w", err)
		}

		trimmedResponse := string(respContent[4:])
		fmt.Println(trimmedResponse)

		trimmedResponse = `{"version":1,"attributeRoute":"A","route":9,"srcId":1,"dstId":2,"rpcId":1,"reply":false,"error":false,"jwt":null,"data":"AAAAAA=="}`

		log.Printf("Sender: Received content: %s\n", trimmedResponse)
		mtsResponseMessage := model.MTSMessage{}

		err = json.Unmarshal([]byte(trimmedResponse), &mtsResponseMessage)
		if err != nil {
			fmt.Println("unmarshal resposne error : ", err)
			return nil, err
		}

		if mtsResponseMessage.Route == 3 {
			break
		}

		go connect.SendAcknowledgmentToServer(&mtsResponseMessage)
	}

	return &mtsResponseMessage, nil
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
		Version:        1,
		Route:          requestType,
		SrcID:          srcID,
		DstID:          dstID,
		RPCID:          rpcID,
		Reply:          false,
		IsError:        isError,
		Data:           data,
		AttributeRoute: attrRoute,
		JWT:            jwt,
	}

	return mtsMessage
}

//StrToPointer converts the string to pointer
func StrToPointer(str string) *string {
	return &str
}
