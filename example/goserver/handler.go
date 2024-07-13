package main

import (
	"fmt"
	. "protogen/generated/game"
)

type handlerSvr struct {
}

func (*handlerSvr) EnterScene(req *EnterSceneReq) {
	fmt.Println("EnterScene", req.SceneId)
}

func (*handlerSvr) LeaveScene(req *LeaveSceneReq) {
	fmt.Println("LeaveScene", req.SceneId)
}

func (*handlerSvr) DoMovement(req *MoveReq) *MoveResp {
	return &MoveResp{
		X: req.X + 6.0,
		Y: req.Y + 7.0,
		Z: req.Z + 8.0,
	}
}

type cbkHandler struct {
}
