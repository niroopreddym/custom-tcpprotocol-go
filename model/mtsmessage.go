package model

import "github.com/niroopreddym/custom-tcpprotocol-go/enum"

//MTSMessage message format
type MTSMessage struct {
	Version byte `json:"version"`

	AttributeRoute string `json:"attributeRoute"`

	Route enum.MTSRequest `json:"route"`

	SrcID int `json:"srcId"`

	DstID int `json:"dstId"`

	RPCID int `json:"rpcId"`

	Reply bool `json:"reply"`

	JWT string `json:"jwt"`

	Data []byte `json:"data"`

	IsError bool `json:"isError"`
}
