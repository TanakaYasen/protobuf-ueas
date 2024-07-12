package main

import (
	"fmt"
	"log"
	"net"
	game "protogen/generated/game"
)

type PlayerSession struct {
	buffer     []byte
	count      int64
	conn       net.Conn
	dispatcher *game.GameC2SImplement
}

func (ps *PlayerSession) SendPackage(buff []byte) {
	length := len(buff)
	//big-endian header
	var header [2]byte = [2]byte{byte((length >> 8) & 0xff), byte(length & 0xff)}
	ps.conn.Write(header[:])
	ps.conn.Write(buff)
}

func (ps *PlayerSession) Close() {
	ps.conn.Close()
}

func (ps *PlayerSession) onRecv(buf []byte, disp *game.GameC2SDispatcher) {
	ps.buffer = append(ps.buffer, buf...)
	ps.count++

	for {
		if len(ps.buffer) < 2 {
			return
		}
		pkglen := int(ps.buffer[0])*256 + int(ps.buffer[1])
		if len(ps.buffer) < (2 + pkglen) {
			return
		}
		disp.OnHandlePackage(ps.buffer[2 : 2+pkglen])
		ps.buffer = ps.buffer[2+pkglen:]
	}
}

func newPlayerSession(conn net.Conn) *PlayerSession {
	return &PlayerSession{
		buffer: make([]byte, 0),
		count:  0,
		conn:   conn,
	}
}

func main() {
	listener, err := net.Listen("tcp", ":3389")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		var buf [2048]byte
		ps := newPlayerSession(conn)
		handler := new(handlerSvr)
		dp := game.MakeGameC2SDispatcher(handler, ps)

		for {
			n, err := conn.Read(buf[:])
			if err != nil {
				break
			}
			ps.onRecv(buf[:n], dp)
		}
		fmt.Println()
	}
}
