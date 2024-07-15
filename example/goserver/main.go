package main

import (
	"fmt"
	"log"
	"net"
	game "protogen/generated/game"
	"protogen/netlib"
)

type PlayerSession struct {
	buffer []byte
	count  int64
	conn   net.Conn
	core   *netlib.MessagerCore
	*game.GameS2CMessager
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

func (ps *PlayerSession) onRecv(buf []byte) {
	ps.buffer = append(ps.buffer, buf...)
	ps.count++

	for {
		if len(ps.buffer) < 2 {
			return
		}
		// big-endian
		pkglen := int(ps.buffer[0])*256 + int(ps.buffer[1])
		if len(ps.buffer) < (2 + pkglen) {
			return
		}
		data := ps.core.OnHandlePackage(ps.buffer[2 : 2+pkglen])
		if data != nil {
			ps.SendPackage(data)
		}
		ps.buffer = ps.buffer[2+pkglen:]
	}
}

func (ps *PlayerSession) Poll() {
	var buf [2048]byte
	for {
		n, err := ps.conn.Read(buf[:])
		if err != nil {
			return
		}
		ps.onRecv(buf[:n])
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
		fmt.Println(conn.RemoteAddr().String(), " has Connected")
		ps := &PlayerSession{
			buffer: make([]byte, 0),
			count:  0,
			conn:   conn,
			core:   netlib.CreateMessageCore(game.MakeGameC2SDispatcher(new(handlerSvr))),
		}
		ps.GameS2CMessager = game.MakeGameS2CMessager(new(cbkHandler), ps.core, ps)
		ps.Poll()

		fmt.Println("conn disconnected")
	}
}
