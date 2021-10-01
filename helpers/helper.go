package helper

import (
	"encoding/binary"
	"fmt"
	"log"
)

//MaxMessageLength is the length limitation for the message to be sent to the server
const MaxMessageLength = 1 << 22 // 2^2 * 2^10 * 2^10 = 4 MiB

//Offset is the length of byte array to indicate the msg length coming from server
const Offset int = 4

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
