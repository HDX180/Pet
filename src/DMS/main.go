package main

import (
	d "DeviceManageServer"
	"fmt"
	"os"
	"os/signal"
)

var b *d.StruBusiness = d.GetBusinessInstance()
var h *d.StruHttpHandle = d.GetHttpInstance()
var w *d.StruWebsocketHandle = d.GetWebsocketInstance()

func dms_init() {
	d.InitLogger("./log/dms.log")
	d.OpenDB("root:123456@tcp(127.0.0.1:3306)/test")
	d.CoapServer_init()
	b.Init()
	h.Init()
	//	w.Init()
}

func dms_start() {
	b.Start()
	h.Start()
	//	w.Start()
}

func dms_close() {
	b.Close()
	h.Close()
	//	w.Close()
}

func main() {

	dms_init()
	dms_start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	s := <-c
	fmt.Println("shut down!", s)
	dms_close()
}
