package DomainSocket

import (
	"bufio"
	//	"log"
	"net"
)

type DomainSocketServer struct {
	ln *net.UnixListener
	//clients [10]*net.UnixConn
}

func NewServer() *DomainSocketServer { return &DomainSocketServer{} }

type ClientStatus func(*net.UnixConn, int)

type ReciveMsg func(*net.UnixConn, string)

func (d *DomainSocketServer) Listen(address string, fun ClientStatus, r ReciveMsg, delim byte) error {
	var unixAddr *net.UnixAddr
	unixAddr, _ = net.ResolveUnixAddr("unix", address)
	lsn, err := net.ListenUnix("unix", unixAddr)
	d.ln = lsn
	if err != nil {
		return err
	}
	go accpet(d.ln, fun, r, delim)

	return nil
}

func accpet(ln *net.UnixListener, fun ClientStatus, r ReciveMsg, delim byte) {
	for {
		unixConn, err := ln.AcceptUnix()
		if err != nil {
			continue
		}

		//clients = append(clients, unixConn)

		fun(unixConn, 1) //connect
		go readBuf(unixConn, fun, r, delim)
	}
}

func readBuf(conn *net.UnixConn, fun ClientStatus, r ReciveMsg, delim byte) {
	defer func() {
		fun(conn, 0) //disconnect
		conn.Close()
	}()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString(delim)
		if err != nil {
			return
		}
		r(conn, message)
	}
}

func (d *DomainSocketServer) WriteBuf(con *net.UnixConn, msg string) {
	bmsg := []byte(msg)
	con.Write(bmsg)
}

func (d *DomainSocketServer) Stop() {
	d.ln.Close()
}
