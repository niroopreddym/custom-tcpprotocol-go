package mtsclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/niroopreddym/custom-tcpprotocol-go/model"
)

//Login logins
type Login struct {
	MTSClient model.MTSClient
}

//NewMTSClientLogin ctor
func NewMTSClientLogin() *Login {
	return &Login{
		MTSClient: model.MTSClient{},
	}
}

//Connect connects to the on premise server
func (login *Login) Connect(hostname string, port int, mtsReceive func(model.MTSClient, model.MTSMessage), mtsDisconnect func(model.MTSClient), timoutMs int) *model.MTSClient {

	conn, err := net.Dial("tcp", "[mtstest:Test123@localhost]:10001")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()

	connbuf := bufio.NewReader(conn)
	for {
		str, err := connbuf.ReadString('\n')
		if err != nil {
			break
		}

		if len(str) > 0 {
			fmt.Println(str)
		}
	}
//fill in the model deatils and send it
	return nil
}

//Send senfs the message to client
func (login *Login) Send(mtsMessage model.MTSMessage) {
	fmt.Println(mtsMessage)
	mutex := sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	data, err := json.Marshal(mtsMessage)
	if err != nil {
		fmt.Println(err)
	}


	Send(data)
}

func send(byte[] msg){
	var data = PrepareData(msg)
	fmt.Printf("Sending message, data size %d", len(data))

	if (Stream.CanWrite)
		Stream.BeginWrite(data, 0, data.Length, WriteComplete, this)
	else
	{
		Log?.LogWarning($"{LoggerMods.CallerInfo()}Cannot write to stream")
		throw new MTSConnectionClosedException("Cannot write to stream")
	}
}

func  PrepareData(byte[] msg) []byte {
	// Expected length should be uint (TA8319)
	uint msgLength = (uint)msg.Length
	fmt.Printf("Send message length %d", msgLength)

	if (msgLength > MTS.MAX_MESSAGE_LENGTH)
	{
		var errorMsg = $"{LoggerMods.CallerInfo()}MTS Message with Expected Length of {msgLength} cannot be sent. " +
			$"Messages longer than {MTS.MAX_MESSAGE_LENGTH} are not supported.";
		Log?.LogError(errorMsg)
		throw new MTSMessageLengthException(errorMsg)
	}

	var len = BitConverter.GetBytes(msgLength)
	byte[] data = new byte[len.Length + msg.Length]
	System.Buffer.BlockCopy(len, 0, data, 0, len.Length)
	System.Buffer.BlockCopy(msg, 0, data, len.Length, msg.Length)

	return data;
}

 func WriteComplete(IAsyncResult ar)
{
	Stream.EndWrite(ar);
	lock (this)
	{
		if (Queue.Count > 0)
		{
			MTSMessage mtsMessage = null;
			Queue.TryDequeue(out mtsMessage);
			Send(JsonConvert.SerializeObject(mtsMessage, Formatting.None));
		}
		else
		{
			QueueNext = false;
		}
	}
}