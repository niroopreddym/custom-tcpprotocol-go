package mtsclient

import (
	"net"
)

//MTSClient provides the struct
type MTSClient struct {
	ID                    string
	Connected             bool
	Hostname              string
	HostAddress           string
	Port                  int
	UseTLS                bool
	ClientCertificate     []byte
	ProxyHostname         string
	ProxyPort             int
	ProxyUser             string
	ProxyPassword         string
	ProxyTransactComplete bool
	ReadBuffer            []byte
	BufferLen             int
	ExpectedLen           uint
	TCPClient             net.Conn
}

// //Send sends the data to MTS Client
// func (client *MTSClient) Send(mtsMessage model.MTSMessage) {

// 	mtsMessageByteData, err := json.Marshal(mtsMessage)
// 	if err != nil {
// 		fmt.Printf("error in marshalling the MTSMessage Data: %w", err)
// 	}

// 	send(mtsMessageByteData)
// }

// func send(msg []byte) {
// 	//find a way to prepare the data
// 	// var data = PrepareData(msg)

// 	//open the connection and send the data here

// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			continue
// 		}

// 		daytime := time.Now().String()
// 		conn.Write([]byte(daytime)) // don't care about return value
// 		conn.Close()                // we're finished with this client
// 	}
// }

// //MaxMessageLength max chunk size that can be written to the server
// const MaxMessageLength = 1 << 22 //  1 x 2 ^ 22 = 4 MB

// func PrepareData() {

// }
