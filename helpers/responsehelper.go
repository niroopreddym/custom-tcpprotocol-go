package helper

import (
	"encoding/json"
	"fmt"

	"github.com/niroopreddym/custom-tcpprotocol-go/enum"
	"github.com/niroopreddym/custom-tcpprotocol-go/model"
)

//CreateResponse creates a MTSMessage as Response
func CreateResponse(requestMessage *model.MTSMessage, responseType enum.MTSRequest, attrRoute *string, isError bool, jwt *string, data []byte) model.MTSMessage {
	mtsMessage := model.MTSMessage{
		Version:        1,
		Route:          responseType,
		SrcID:          requestMessage.SrcID,
		DstID:          requestMessage.DstID,
		RPCID:          requestMessage.RPCID,
		Reply:          true,
		IsError:        isError,
		Data:           data,
		AttributeRoute: attrRoute,
		JWT:            jwt,
	}

	return mtsMessage
}

//CreateErrorResponse error response
func CreateErrorResponse(errorID enum.MtsErrorID, errorMsg string, requestMessage *model.MTSMessage, responseType enum.MTSRequest, attrRoute *string, jwt *string) model.MTSMessage {
	var errorResponseData = model.MtsErrorResponse{
		MtsError:        errorID,
		MtsErrorMessage: errorMsg,
	}

	errorResponseByteArray, err := json.Marshal(errorResponseData)
	if err != nil {
		fmt.Println("Error getting the client cert", err)
	}

	return CreateResponse(requestMessage, responseType, attrRoute, true, jwt, errorResponseByteArray)
}
