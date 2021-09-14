package enum

//TestEvent is the enum to test events
type TestEvent int

const (
	//SendOPLPayload is test payload enum
	SendOPLPayload = 1
	//Exit is the exit enum
	Exit = 2
)

func (v TestEvent) String() string {
	dictMap := map[TestEvent]string{
		1: "SendOPLPayload",
		2: "Exit",
	}

	return dictMap[v]
}
