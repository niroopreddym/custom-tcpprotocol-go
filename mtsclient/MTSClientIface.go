package mtsclient

import "github.com/niroopreddym/custom-tcpprotocol-go/model"

//ClientInterface interface
type ClientInterface interface {
	Send(mtsMessage model.MTSMessage)

	Connect(hostname string, port int, mtsReceive func(model.MTSClient, model.MTSMessage), mtsDisconnect func(model.MTSClient), timoutMs int) model.MTSClient

	// MTSConnect(client model.MTSClient)

	// MTSReceive(client model.MTSClient, mtsMessage model.MTSMessage)

	// MTSDisconnect(client model.MTSClient)
}
