package main

import (
	"log"
	"net"
	game "protogen/generated/game"
	"protogen/netlib"
	"time"
)

type PlayerSession struct {
	buffer []byte
	count  int64
	conn   net.Conn
	core   *netlib.MessagerCore
	*game.GameC2SMessager
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
	conn, err := net.Dial("tcp", ":3389")
	if err != nil {
		log.Fatalln(err)
	}
	core := netlib.CreateMessageCore(game.MakeGameS2CDispatcher(new(handlerCli)))
	ps := &PlayerSession{
		buffer: make([]byte, 0),
		count:  0,
		conn:   conn,
		core:   core,
	}
	ps.GameC2SMessager = game.MakeGameC2SMessager(new(cbkHandler), core, ps)

	go func() {
		for {
			req := &game.EnterSceneReq{}
			req.SceneId = 9999
			ps.SendEnterScene(req)

			moveReq := &game.MoveReq{
				X: 1.0,
				Y: 2.0,
				Z: 3.0,
			}
			ps.CallDoMovement(moveReq)
			time.Sleep(time.Second)
		}
	}()
	ps.Poll()
}
