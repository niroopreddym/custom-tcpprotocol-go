package main

import (
	"github.com/niroopreddym/custom-tcpprotocol-go/model"
	"github.com/niroopreddym/custom-tcpprotocol-go/mtsclient"
)

func main() {
	login := mtsclient.NewMTSClientLogin()
	login.Connect("127.0.0.1", 10001, sample, sample2, 10000)
}

func sample(model.MTSClient, model.MTSMessage) {

}

func sample2(model.MTSClient) {

}
