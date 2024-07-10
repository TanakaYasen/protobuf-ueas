package main

import (
	"log"
	"net"
	game "protogen/generated"
)

type PlayerSession struct {
	buffer     []byte
	count      int64
	conn       net.Conn
	dispatcher *game.GameC2SDispatcher
}

func (ps *PlayerSession) SendPackage(buff []byte) {
	length := len(buff)
	//big header
	var header [2]byte = [2]byte{byte((length >> 8) & 0xff), byte(length & 0xff)}
	ps.conn.Write(header[:])
	ps.conn.Write(buff)
}

func (ps *PlayerSession) Close() {
	ps.conn.Close()
}

func NewPlayerSession(conn net.Conn, dispatcher *game.GameC2SDispatcher) *PlayerSession {
	var ps = new(PlayerSession)
	ps.buffer = make([]byte, 0)
	ps.count = 0
	ps.conn = conn
	ps.dispatcher = dispatcher
	return ps
}

func (ps *PlayerSession) OnRecv(buf []byte) {
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
		ps.dispatcher.OnHandlePkg(ps.buffer[2 : 2+pkglen])
		ps.buffer = ps.buffer[2+pkglen:]
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
		var buf [1024]byte
		dis := game.NewGameC2SDispatcher(new(handler))
		ps := NewSession(conn, dis)
		for {
			n, err := conn.Read(buf[:])
			if err != nil {
				break
			}
			ps.OnRecv(buf[:n])
		}
	}
}
