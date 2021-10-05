package mtsclient

import "github.com/niroopreddym/custom-tcpprotocol-go/model"

//ITCPConnect exposes the required methods for the communiation with the Onity Server
type ITCPConnect interface {
	ConnectAndLogin()
	WithTLS(certificate []byte)
	SendMTSOPLPayload(mtsOPLPayload *model.MtsOplPayload)
}
