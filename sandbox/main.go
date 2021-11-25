package main

import (
	"time"

	"github.com/Ishan27g/go-utils/pEnv"
)

func main() {
	p := pEnv.Init(time.Millisecond * 50)
	p.CmdFromMap(map[string]string{"HOST_NAME": "localhost", "UDP_PORT": "1101", "ProcessName": "P1"}, "./gossiper", nil)
	p.CmdFromMap(map[string]string{"HOST_NAME": "localhost", "UDP_PORT": "1102", "ProcessName": "P2"}, "./gossiper", nil)
	//p.CmdFromMap(map[string]string{"HOST_NAME": "localhost", "UDP_PORT": "1103", "ProcessName": "P3"}, "./gossiper", nil)
	//p.CmdFromMap(map[string]string{"HOST_NAME": "localhost", "UDP_PORT": "1104", "ProcessName": "P4"}, "./gossiper", nil)
	<-make(chan bool)
}
