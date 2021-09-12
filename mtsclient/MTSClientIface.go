package mtsclient

import "github.com/niroopreddym/custom-tcpprotocol-go/model"

//ClientInterface interface
type ClientInterface interface {
	Send(mtsMessage model.MTSMessage)
	ConnectAndLogin() error

	Connect(hostname string, port int, mtsReceive func(MTSClient, model.MTSMessage), mtsDisconnect func(MTSClient), timoutMs int) MTSClient

	// MTSConnect(client model.MTSClient)

	// MTSReceive(client model.MTSClient, mtsMessage model.MTSMessage)

	// MTSDisconnect(client model.MTSClient)
}
