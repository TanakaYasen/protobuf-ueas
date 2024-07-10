package main

import (
	"log"
	"net"
	game "protogen/generated/game"
	"time"

	proto "github.com/gogo/protobuf/proto"
)

type PlayerSession struct {
	buffer     []byte
	count      int64
	conn       net.Conn
	dispatcher *game.SvrGameC2SDispatcher
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

func (ps *PlayerSession) onRecv(buf []byte, disp *game.SvrGameC2SDispatcher) {
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
		disp.OnHandlePkg(ps.buffer[2 : 2+pkglen])
		ps.buffer = ps.buffer[2+pkglen:]
	}
}

func NewPlayerSession(conn net.Conn, dispatcher *game.SvrGameC2SDispatcher) *PlayerSession {
	var ps = new(PlayerSession)
	ps.buffer = make([]byte, 0)
	ps.count = 0
	ps.conn = conn
	ps.dispatcher = dispatcher
	return ps
}

func main() {
	conn, err := net.Dial("tcp", ":3389")
	if err != nil {
		log.Fatalln(err)
	}

	var buf [1024]byte
	ps := NewPlayerSession(conn, nil)
	dx := game.NewGameC2SDispatcher(new(handlerCli), ps)

	go func() {
		var i uint64
		for {
			var pkg = &game.Package{}
			i++
			pkg.SessionId = i
			pkg.Route = "EnterScene"

			var esreq = &game.EnterSceneReq{}
			esreq.SceneId = int32(i * 2)
			d, _ := proto.Marshal(esreq)
			pkg.Data = d

			d, _ = proto.Marshal(pkg)

			ps.SendPackage(d)
			time.Sleep(time.Second)
		}
	}()
	for {
		n, err := conn.Read(buf[:])
		if err != nil {
			break
		}
		ps.onRecv(buf[:n], dx)
	}
}
