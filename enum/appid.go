package enum

//AppID is the enum
type AppID int

const (
	//RMSServer is for RMS Integrators
	RMSServer = 100
	//RMSEmulator is for RMS Emulators
	RMSEmulator = 101
	//BTPP is a enum
	BTPP = 201
	//MobilePP is a enum
	MobilePP = 202
)

func (v AppID) String() string {
	dictMap := map[AppID]string{
		100: "RMSServer",
		101: "RMSEmulator",
		201: "BTPP",
		202: "MobilePP",
	}

	return dictMap[v]
}
