package model

type MTSResult struct {
	HasError  bool
	ErrorData error
	RpcId     int
	Jwt       string
	Data      interface{}
}
