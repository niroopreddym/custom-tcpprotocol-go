package model

//MTSResult generic result response
type MTSResult struct {
	HasError  bool
	ErrorData error
	RPCID     int
	Jwt       string
	Data      interface{}
}
