package model

//MtsLoginResponse response type fro login
type MtsLoginResponse struct {
	//NodeAuth Not used
	NodeAuth string
	//ClientCertificate Client side certificate to be used in a certificate login.
	ClientCertificate []byte
	//ServerCertInfo Not used
	ServerCertInfo string
	//MtuBluetooth Not used
	MtuBluetooth int
	//MtuOpl Not used
	MtuOpl int
	//MtuMts Not used
	MtuMts int
}
