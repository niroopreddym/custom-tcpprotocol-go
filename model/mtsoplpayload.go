package model

//MtsOplPayload login payload to the onity server
type MtsOplPayload struct {
	//RoomID is Destination Room ID
	RoomID string `json:"RoomId"`
	//ProxyMACAddress Proxy MAC Address
	ProxyMACAddress *string `json:"ProxyMACAddress"`
	//Data is OPL Message Data
	Data []byte `json:"data"`
}
