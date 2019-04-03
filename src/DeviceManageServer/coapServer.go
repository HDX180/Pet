package DeviceManageServer

import (
	"encoding/json"
	"fmt"
	"github.com/dustin/go-coap"
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
	//logger.Info(fmt.Sprintf("Got message in handleRegister: path=%q: %#v from %v", m.Path(), m, a))

	var regDevinfo coap_RegDevInfo
	json.Unmarshal(m.Payload, &regDevinfo)

	if devinfo := business.getDevInfo(regDevinfo.CodeID); devinfo != nil {
		devinfo.keepaliveTime = time.Now()
		devinfo.status = true
		devinfo.version = regDevinfo.Version
		devinfo.host = a.String()
	} else {
		//	logger.Info(fmt.Sprintf("dev codeID = %d is not exist", regDevinfo.CodeID))
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

	//logger.Info(fmt.Sprintf("Transmitting from A %#v", res))
	return res
	//	}
}

func handleKeepalive(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
	strCodeID := m.Option(15).(string)
	strCodeID = string([]byte(strCodeID)[strings.IndexByte(strCodeID, '=')+1:])
	codeID, _ := strconv.Atoi(strCodeID)
	//	logger.Info(fmt.Sprintf("dev codeID = %d keep alive", codeID))
	if bIsClose {
		return nil
	}
	business.devAliChan <- codeID

	res := &coap.Message{
		Type:      coap.Acknowledgement,
		Code:      coap.Content,
		MessageID: m.MessageID,
	}
	return res
}

func routeRegistry(mux *coap.ServeMux) {
	mux.Handle("/register", coap.FuncHandler(handleRegister))
	mux.Handle("/keepalive", coap.FuncHandler(handleKeepalive))

}

func StartCoapServer() {
	mux := coap.NewServeMux()

	routeRegistry(mux)

	go func() {
		err := coap.ListenAndServe("udp", ":5683", mux)
		if err != nil {
			logger.Error(fmt.Sprintf("coap ListenAndServe error : %s", err.Error()))
		}
	}()

}
