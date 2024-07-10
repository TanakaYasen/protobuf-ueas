package main

import (
	. "protogen/generated/game"
)

type handlerCli struct {
}

func (*handlerCli) EnterScene(EnterSceneReq) {

}

func (*handlerCli) LeaveScene(LeaveSceneReq) {

}

func (*handlerCli) DoMovement(MoveReq) MoveResp {
	return MoveResp{}
}
