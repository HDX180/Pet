package DeviceManageServer

import (
	"crypto/md5"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"set"
	"strconv"
	"time"
)

var websocketHandle *StruWebsocketHandle = &StruWebsocketHandle{}

func GetWebsocketInstance() *StruWebsocketHandle {
	return websocketHandle
}

type topicType int

const (
	TOPIC_DEVSTATUS topicType = iota // value --> 0
)

type struCoapPushMag struct {
	topic     topicType
	devCodeID int
	struMsg   []byte
}

type StruWebsocketHandle struct {
	subInfoMap map[int][]*clientSubInfo //<codeID,[]*clientSubInfo> codeID ->userID
	clientsMap map[int]*clientConn      //<userID,*clientConn>      userID->Conn
	msg        chan struCoapPushMag
	mux        *http.ServeMux
	upgrader   websocket.Upgrader
}

type clientSubInfo struct {
	topics *set.HashSet
	userID int
}

type clientConn struct {
	con     *websocket.Conn
	msg     chan struCoapPushMag
	isClose bool
}

func initClientConn(wsConn *websocket.Conn) *clientConn {
	conn := &clientConn{
		con: wsConn,
		msg: make(chan struCoapPushMag, 1000),
	}
	return conn
}

func makeUserID() string {
	curtime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(curtime, 10))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func wsHandle(w http.ResponseWriter, r *http.Request) {
	conn, err := websocketHandle.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	clientCon := initClientConn(conn)
	defer func() {
		clientConn.isClose = true //不要在接收端关闭chan,这里置标志位
	}()

	vars := r.URL.Query()
	userID, _ := strconv.Atoi(vars["userID"][0])
	websocketHandle.clientsMap[userID] = clientCon

	for {
		select {
		case msg, ok := <-clientCon.msg:
			if ok {
				err := conn.WriteMessage(1, msg.struMsg)
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
		}
	}
}

func (w *StruWebsocketHandle) pushCoapMsg() {
	for {
		pushmsg, ok := <-w.msg
		clients, ok := w.subInfoMap[pushmsg.devCodeID]
		if !ok {
			log.Println("no client subscribe this dev:%d", pushmsg.devCodeID)
		}

		for _, c := range clients {
			if c.topics.Contains(pushmsg.topic) {
				clientCon, ok := w.clientsMap[c.userID]
				if ok {
					if clientCon.isClose == true {
						clientCon.con.Close()
						delete(w.clientsMap, c.userID)
					} else {
						clientCon.msg <- pushmsg
					}
				}
			}
		}
	}
}

func (w *StruWebsocketHandle) addSubscribe(userID int, topics []topicType, codeID int) bool {
	bIsFound := false
	var clients []*clientSubInfo
	clients, ok := w.subInfoMap[codeID]
	if !ok {
		clients = make([]*clientSubInfo, 1)
		w.subInfoMap[codeID] = clients
	}
	for _, c := range clients {
		if c != nil && c.userID == userID {
			for _, t := range topics {
				c.topics.Add(t)
			}
			bIsFound = true
		}
	}
	if !bIsFound {
		newTopics := set.NewHashSet()
		for _, t := range topics {
			newTopics.Add(t)
		}

		client := &clientSubInfo{
			topics: newTopics,
			userID: userID,
		}
		w.subInfoMap[codeID] = append(clients, client)
	}
	return true
}

func (w *StruWebsocketHandle) rmSubscribe(userID int, topics []topicType, codeID int) bool {
	clients, ok := w.subInfoMap[codeID]
	if !ok {
		return false
	}
	for _, c := range clients {
		if c.userID == userID {
			for _, t := range topics {
				c.topics.Remove(t)
			}
		}
	}
	return true
}

func (w *StruWebsocketHandle) Init() {
	mux := http.NewServeMux()
	w.mux = mux

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	w.upgrader = upgrader

	mux.HandleFunc("/ws", wsHandle)

}

func (w *StruWebsocketHandle) Start() {
	go func() {
		log.Fatal(http.ListenAndServe(":6520", w.mux))
	}()
	go w.pushCoapMsg()
}

func (w *StruWebsocketHandle) Close() {

}
