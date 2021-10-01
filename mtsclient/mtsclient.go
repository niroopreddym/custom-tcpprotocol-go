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
