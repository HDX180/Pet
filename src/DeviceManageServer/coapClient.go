package DeviceManageServer

import (
	//	"fmt"
	"encoding/json"
	"github.com/dustin/go-coap"

	"log"
)

type coap_struGetTempReq struct {
	host string
}

type coap_struGetTempResp struct {
	CodeID      int `json:"CodeID"`
	Temperature int `json:"Temperature"`
}

func coapclient_getTemperature(m_req *coap_struGetTempReq, m_resp *coap_struGetTempResp) error {
	req := coap.Message{
		Type:      coap.Confirmable,
		Code:      coap.GET,
		MessageID: 12345,
	}

	path := "/pet/health"

	req.SetPathString(path)

	c, err := coap.Dial("udp", m_req.host)
	if err != nil {
		log.Printf("Error dialing: %v", err)
		return err
	}

	rv, err := c.Send(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}

	if rv != nil {
		//log.Printf("response rv: %v", rv)
		log.Printf("Response payload: %s", rv.Payload)
		json.Unmarshal(rv.Payload, m_resp)
	}

	return err
}
