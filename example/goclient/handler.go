package main

import (
	"fmt"
	. "protogen/generated/game"
)

type handlerCli struct {
}

func (*handlerCli) DoMovement(sessionId uint32, resp *MoveResp) {
	fmt.Println("move", resp.X, resp.Y, resp.Z)
}
