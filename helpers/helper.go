package helper

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/niroopreddym/custom-tcpprotocol-go/enum"
	"github.com/niroopreddym/custom-tcpprotocol-go/model"
)

//UintSize max unsigned int size
const UintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

//MaxMessageLength is the length limitation for the message to be sent to the server
const MaxMessageLength = 1 << 22 // 2^2 * 2^10 * 2^10 = 4 MiB

//Offset is the length of byte array to indicate the msg length coming from server
const Offset int = 4

const (
	//MaxInt max signed integer value
	MaxInt = 1<<(UintSize-1) - 1 // 1<<31 - 1 or 1<<63 - 1
	//MinInt min signed int value
	MinInt = -MaxInt - 1 // -1 << 31 or -1 << 63
	//MaxUint is max unsigned int value
	MaxUint = 1<<UintSize - 1 // 1<<32 - 1 or 1<<64 - 1
)

//PrepareData prepares the data to be sent
func PrepareData(msg []byte) []byte {
	// Expected length should be uint (TA8319)
	msgLength := len(msg)
	log.Println("Send message length ", msgLength)

	if msgLength > MaxMessageLength {
		var errorMsg = fmt.Sprintf("Messages longer than %d are not supported.", MaxMessageLength)
		log.Println(errorMsg)
		log.Panic("msglength exception")
	}

	//get bytes fro msgLength for now hard code to 4 i.e., BitConverter.GetBytes(msgLength)
	byteArrayLen := ConvertIntToByte(int32(msgLength))
	data := make([]byte, len(byteArrayLen)+msgLength)

	copy(data, byteArrayLen)
	copy(data[len(byteArrayLen):], msg)

	return data
}

//ConvertByteToInt converts the byte data to 4 bit int
func ConvertByteToInt(bytes []byte) int {
	i := int32(binary.LittleEndian.Uint32(bytes))
	return int(i)
}

//ConvertIntToByte converts the int to 4 bit byte data
func ConvertIntToByte(msgLength int32) []byte {
	buff := make([]byte, Offset)
	binary.LittleEndian.PutUint32(buff, uint32(msgLength))
	return buff
}

//StrToPointer converts the string to pointer
func StrToPointer(str string) *string {
	return &str
}

//CreateRequest creates a MTSMessage payload structure used to send to the server
func CreateRequest(requestType enum.MTSRequest, attrRoute *string, srcID int, dstID int, isError bool, jwt *string, data []byte) model.MTSMessage {
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
	}

	conn, err := tls.Dial("tcp", connectionString, tlsConfig)
	if err != nil {
		fmt.Println("error occured while establishing the connection: ", err)
		return nil, err
	}

	return conn, err
}

//CreateResponse creates a MTSMessage as Response
func CreateResponse(requestMessage *model.MTSMessage, responseType enum.MTSRequest, attrRoute *string, isError bool, jwt *string, data []byte) model.MTSMessage {
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

//CreateErrorResponse error response
func CreateErrorResponse(errorID enum.MtsErrorID, errorMsg string, requestMessage *model.MTSMessage, responseType enum.MTSRequest, attrRoute *string, jwt *string) model.MTSMessage {
	var errorResponseData = model.MtsErrorResponse{
		MtsError:        errorID,
		MtsErrorMessage: errorMsg,
	}

	errorResponseByteArray, err := json.Marshal(errorResponseData)
	if err != nil {
		fmt.Println("Error getting the client cert", err)
	}

	return CreateResponse(requestMessage, responseType, attrRoute, true, jwt, errorResponseByteArray)
}
