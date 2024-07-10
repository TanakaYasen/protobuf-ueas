package main

import (
	"fmt"
	. "protogen/generated/game"
)

type handlerSvr struct {
}

func (*handlerSvr) EnterScene(req EnterSceneReq) {
	fmt.Println("EnterScene", req.SceneId)
}

func (*handlerSvr) LeaveScene(req LeaveSceneReq) {
	fmt.Println("LeaveScene", req.SceneId)
}

func (*handlerSvr) DoMovement(req MoveReq) MoveResp {
	return MoveResp{}
}
