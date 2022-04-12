package main

import (
	"github.com/ligao-cloud-native/cloud-shell/servers"
	"k8s.io/component-base/logs"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	logs.InitLogs()
	defer logs.FlushLogs()

	servers.StartHttpServer()

}
