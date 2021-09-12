package mtsclient

import "github.com/niroopreddym/custom-tcpprotocol-go/enum"

//MtsLogin login the user
type MtsLogin struct {
	Username          *string
	Password          *string
	AppID             enum.AppID
	AppKey            []byte
	ClientCertificate []byte
}

//OperatorLoginPayload Construct an operator message payload
func (mtsLogin *MtsLogin) OperatorLoginPayload(username string, password string, appID enum.AppID, appKey []byte) {
	mtsLogin.ClientCertificate = nil
}

//CertificateLoginPayload Construct a certificate message payload
func (mtsLogin *MtsLogin) CertificateLoginPayload(appID enum.AppID, appKey []byte, cert []byte) {
	mtsLogin.Username = nil
	mtsLogin.Password = nil
	mtsLogin.ClientCertificate = cert
}
