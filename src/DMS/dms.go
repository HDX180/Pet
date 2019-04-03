package main

import (
	d "DeviceManageServer"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

var b *d.StruBusiness = d.GetBusinessInstance()
var h *d.StruHttpHandle = d.GetHttpInstance()
var w *d.StruWebsocketHandle = d.GetWebsocketInstance()
var c *d.StruConfig = d.GetConfigInstance()

func dms_init() {
	d.InitLogger("./log/dms.log")

	c.Init("./config.yml")

	d.OpenDB(c.GetMySqlUri())
	b.Init()
	h.Init()
	//	w.Init()
}

func dms_start() {
	b.Start()
	h.Start()
	d.StartCoapServer()
	//	w.Start()
}

func dms_close() {
	h.Close()
	b.Close()

	//	w.Close()
}

func main() {

	dms_init()
	dms_start()

	go func() {
		fmt.Println(http.ListenAndServe("localhost:10000", nil))
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	s := <-c
	fmt.Println("shut down!", s)
	dms_close()
}
