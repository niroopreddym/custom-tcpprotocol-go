package enum

//MtsErrorID is the enum
type MtsErrorID int

const (
	//SystemError is error
	SystemError = 1
	//InvalidLogin is error
	InvalidLogin = 2
	//InvalidAppKey is error
	InvalidAppKey = 3
	//InvalidAppID is error
	InvalidAppID = 4
	//InvalidRequest is error
	InvalidRequest = 5
	//UnroutableMessage is error
	UnroutableMessage = 6
	//InvalidFormat is error
	InvalidFormat = 7
	//InvalidJWT is error
	InvalidJWT = 8
)

func (v MtsErrorID) String() string {
	dictMap := map[MtsErrorID]string{
		1: "SystemError",
		2: "InvalidLogin",
		3: "InvalidAppKey",
		4: "InvalidAppID",
		5: "InvalidRequest",
		6: "UnroutableMessage",
		7: "InvalidFormat",
		8: "InvalidJWT",
	}

	return dictMap[v]
}
