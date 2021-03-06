package DeviceManageServer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var httpHandle *StruHttpHandle = &StruHttpHandle{}
var bIsClose = false

type StruHttpHandle struct {
	mux *http.ServeMux
}

func GetHttpInstance() *StruHttpHandle {
	return httpHandle
}

func (h *StruHttpHandle) Init() {
	mux := http.NewServeMux()
	h.mux = mux
	mux.HandleFunc("/pet/health", getPetHealthHandler)
	mux.HandleFunc("/time", getTimeHandler)
	mux.HandleFunc("/cacheTest", cacheTestHandler)
	mux.HandleFunc("/subscribe", subHandler)
	mux.HandleFunc("/unSubscribe", unsubHandler)
}

func checkAccessToken() bool {
	// resp, err:= http.Get("http://")
	// if err != nil {
	// 	log.Println(err)
	// }
	// defer resp.Body.Close()

	// buf:=bytes.NewBuffer(make([]byte,0,512))
	// length,_ := buf.ReadForm(resp.Body)
	return true
}

func cacheTestHandler(w http.ResponseWriter, r *http.Request) {
	cacheResp := &struCacheTestReq{
		Cache: intoCache,
		Total: totalReqNum,
		Rate:  intoCache * 100 / totalReqNum,
	}
	if data, err := json.Marshal(*cacheResp); err == nil {
		w.Write(data)
	}
}

func getPetHealthHandler(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	//accessToken := vars["accessToken"][0]
	//checkAccessToken()
	codeID, _ := strconv.Atoi(vars["codeID"][0])

	getPetHealthReq := &struGetPetHealthReq{codeID: codeID}
	getPetHealthResp := new(struGetPetHealthResp)
	business.getPetHealth(getPetHealthReq, getPetHealthResp)

	//stru->Json
	if data, err := json.Marshal(*getPetHealthResp); err == nil {
		w.Write(data)
	}
}

func getTimeHandler(w http.ResponseWriter, r *http.Request) {

	tm := time.Now().Format(time.RFC1123)
	w.Write([]byte("The time is: " + tm))
}

func subHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("read body err, %v\n", err)
		return
	}
	var req struSubscribeReq
	if err = json.Unmarshal(body, &req); err != nil {
		fmt.Printf("Unmarshal err, %v\n", err)
		return
	}
	ret := websocketHandle.addSubscribe(req.UserID, req.Topics, req.CodeID)
	resp := new(StruCommonResp)
	if ret {
		resp.setCommonResp(DMS_ERR_SUCCESS)
	}

	//stru->Json
	if data, err := json.Marshal(*resp); err == nil {
		w.Write(data)
	}
}

func unsubHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("read body err, %v\n", err)
		return
	}
	var req struSubscribeReq
	if err = json.Unmarshal(body, &req); err != nil {
		fmt.Printf("Unmarshal err, %v\n", err)
		return
	}
	ret := websocketHandle.rmSubscribe(req.UserID, req.Topics, req.CodeID)
	resp := new(StruCommonResp)
	if ret {
		resp.setCommonResp(DMS_ERR_SUCCESS)
	}

	//stru->Json
	if data, err := json.Marshal(*resp); err == nil {
		w.Write(data)
	}
}

func (h *StruHttpHandle) Start() {
	go func() {
		err := http.ListenAndServe(":4637", h.mux)
		if err != nil {
			logger.Error(fmt.Sprintf("http ListenAndServe error : %s", err.Error()))
		}
	}()
}

func (h *StruHttpHandle) Close() {
	bIsClose = true
	return
}
