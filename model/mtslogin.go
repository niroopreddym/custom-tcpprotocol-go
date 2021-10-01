package model

import "github.com/niroopreddym/custom-tcpprotocol-go/enum"

//MtsLogin login the user
type MtsLogin struct {
	Username          *string
	Password          *string
	AppID             enum.AppID
	AppKey            []byte
	ClientCertificate []byte
}
