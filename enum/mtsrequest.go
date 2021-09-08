package enum

//MTSRequest is the enum
type MTSRequest int

const (
	//ErrorResponse least level iota enum
	ErrorResponse = -1
	//UseAttributeRoute is enum
	UseAttributeRoute = 0
	//OPL is enum
	OPL = 1
	//Login is enum
	Login = 2
	//LoginResponse is enum
	LoginResponse = 3
	//CommunicationKeyReq is enum
	CommunicationKeyReq = 4
	//PPCommunicationKeys is enum
	PPCommunicationKeys = 5
	//RMSCommunicationKeys is enums
	RMSCommunicationKeys = 6
	//RoomsMap is enums
	RoomsMap = 7
	//Firmware is enums
	Firmware = 8
	//RMSPing is enums
	RMSPing = 9
	//RMSPingResponse is enums
	RMSPingResponse = 10
	//OplCommands is enums
	OplCommands = 11
	//InitializeLock is enums
	InitializeLock = 12
	//MessageCounter is enums
	MessageCounter = 13
	//RMSDevices is enums
	RMSDevices = 14
)

func (v MTSRequest) String() string {
	dictMap := map[MTSRequest]string{
		-1: "ErrorResponse",
		0:  "UseAttributeRoute",
		1:  "OPL",
		2:  "Login",
		3:  "LoginResponse",
		4:  "CommunicationKeyReq",
		5:  "PPCommunicationKeys",
		6:  "RMSCommunicationKeys",
		7:  "RoomsMap",
		8:  "Firmware",
		9:  "RMSPing",
		10: "RMSPingResponse",
		11: "OplCommands",
		12: "InitializeLock",
		13: "MessageCounter",
		14: "RMSDevices",
	}

	return dictMap[v]
}
