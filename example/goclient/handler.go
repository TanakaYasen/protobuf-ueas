package main

import (
	"fmt"
	game "protogen/generated/game"
)

type handlerCli struct {
}

func (*handlerCli) ItemUpdate(*game.ItemUpdateReq) {
}

func (*handlerCli) LevelUp(*game.LevelUpReq) {
}

type cbkHandler struct {
}

func (c *cbkHandler) DoMovement(u uint32, resp *game.MoveResp) {
	fmt.Println(resp.X, resp.Y, resp.Z)
}
