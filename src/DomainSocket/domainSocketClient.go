package DomainSocket

import (
	"bufio"
	"net"
	//	"log"
)

type DomainSocketClient struct {
	conn *net.UnixConn
}

func NewClient() *DomainSocketClient { return &DomainSocketClient{} }

type ReciveMsg func(string)

func (c *DomainSocketClient) Connect(address string, fun ReciveMsg, delim byte) {
	var unixAddr *net.UnixAddr
	unixAddr, _ = net.ResolveUnixAddr("unix", address)

	con, err := net.DialUnix("unix", nil, unixAddr)
	c.conn = con
	if err != nil {
		//fmt.Printf("conn server failed,error:%s\n", err)
	}
	go onMessageRecived(c.conn, fun, delim)
}

func (c *DomainSocketClient) DisConnect() {
	c.conn.Close()
}

func (c *DomainSocketClient) WriteBuf(msg string) {
	b := []byte(msg)
	c.conn.Write(b)
}

func onMessageRecived(conn *net.UnixConn, fun ReciveMsg, delim byte) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString(delim)
		fun(msg)
		if err != nil {
			conn.Close()
			//	fmt.Printf("reader.ReadString failed,error:%s\n", err)
			break
		}
	}
}
