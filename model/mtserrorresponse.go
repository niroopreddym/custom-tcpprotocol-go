package model

import "github.com/niroopreddym/custom-tcpprotocol-go/enum"

//MtsErrorResponse response ping to server
type MtsErrorResponse struct {
	/// Error ID
	MtsError enum.MtsErrorID
	/// Error Message
	MtsErrorMessage string
}
