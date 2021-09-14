package model

//MtsOplPayload login payload to the onity server
type MtsOplPayload struct {
	//RoomID is Destination Room ID
	RoomID string
	//ProxyMACAddress Proxy MAC Address
	ProxyMACAddress string
	//Data is OPL Message Data
	Data []byte
}
