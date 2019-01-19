package DeviceManageServer

import (
	"encoding/json"
	"github.com/dustin/go-coap"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type coap_RegDevInfo struct {
	Version string `json:"version"`
	CodeID  int    `json:"codeID"`
}

func handleRegister(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
	log.Printf("Got message in handleRegister: path=%q: %#v from %v", m.Path(), m, a)

	var regDevinfo coap_RegDevInfo
	json.Unmarshal(m.Payload, &regDevinfo)

	if devinfo := business.getDevInfo(regDevinfo.CodeID); devinfo != nil {
		devinfo.keepaliveTime = time.Now()
		devinfo.status = true
		devinfo.version = regDevinfo.Version
		devinfo.host = a.String()
	} else {
		return nil //数据库不存在时不回复确认
	}

	//	if m.IsConfirmable() {
	res := &coap.Message{
		Type:      coap.Acknowledgement,
		Code:      coap.Changed,
		MessageID: m.MessageID,
		//Token:     m.Token,
		//Payload:   []byte("hello to you!"),
	}
	//		res.SetOption(coap.ContentFormat, coap.TextPlain)

	log.Printf("Transmitting from A %#v", res)
	return res
	//	}
}

func handleKeepalive(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
	//log.Printf("Got message in handleKeepalive: path=%q: %#v from %v", m.Path(), m, a)
	strCodeID := m.Option(15).(string)
	strCodeID = string([]byte(strCodeID)[strings.IndexByte(strCodeID, '=')+1:])
	codeID, _ := strconv.Atoi(strCodeID)
	log.Printf("dev codeID = %d", codeID)
	if devinfo := business.getDevInfo(codeID); devinfo != nil {
		devinfo.keepaliveTime = time.Now()
		devinfo.status = true
	}

	res := &coap.Message{
		Type:      coap.Acknowledgement,
		Code:      coap.Content,
		MessageID: m.MessageID,
	}
	//log.Printf("Transmitting from A %#v", res)
	return res
}

func routeRegistry(mux *coap.ServeMux) {
	mux.Handle("/register", coap.FuncHandler(handleRegister))
	mux.Handle("/keepalive", coap.FuncHandler(handleKeepalive))

}

func CoapServer_init() {
	mux := coap.NewServeMux()

	routeRegistry(mux)

	go func() {
		log.Fatal(coap.ListenAndServe("udp", ":5683", mux))
	}()

}
