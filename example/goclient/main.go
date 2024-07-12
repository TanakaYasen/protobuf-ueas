package main

import (
	"log"
	"net"
	game "protogen/generated/game"
	"time"
)

type PlayerSession struct {
	buffer []byte
	count  int64
	conn   net.Conn
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

func (ps *PlayerSession) onRecv(buf []byte, disp *game.GameC2SPostHelper) {
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

func NewPlayerSession(conn net.Conn) *PlayerSession {
	var ps = new(PlayerSession)
	ps.buffer = make([]byte, 0)
	ps.count = 0
	ps.conn = conn
	return ps
}

func main() {
	conn, err := net.Dial("tcp", ":3389")
	if err != nil {
		log.Fatalln(err)
	}

	var buf [1024]byte
	ps := NewPlayerSession(conn)
	helper := game.MakeGameC2SHelper(ps, new(handlerCli))

	go func() {
		for {
			req := &game.EnterSceneReq{}
			req.SceneId = 9999
			helper.SendEnterScene(req)

			moveReq := &game.MoveReq{
				X: 1.0,
				Y: 2.0,
				Z: 3.0,
			}
			helper.CallDoMovement(moveReq)
			time.Sleep(time.Second)
		}
	}()
	for {
		n, err := conn.Read(buf[:])
		if err != nil {
			break
		}
		ps.onRecv(buf[:n], helper)
	}
}
