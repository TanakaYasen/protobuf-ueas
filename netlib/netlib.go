package netlib

import (
	proto "github.com/gogo/protobuf/proto"
)

type INetConnection interface {
	SendPackage([]byte)
	Close()
}

type IDispatcher interface {
	OnDispatchPackage(string, []byte) []byte
	Name() string
}

type IPkgMaker interface {
	MakeSendPkg(name string, payload []byte) []byte
	MakeCallPkg(name string, payload []byte, cbk func(uint32, []byte)) []byte
}

type MessagerCore struct {
	seq        uint32
	stubs      map[uint32]func(uint32, []byte)
	dispatcher IDispatcher
}

//go:generate protoc --gogofast_out=. netlib.proto
func (m *MessagerCore) OnHandlePackage(msg []byte) []byte {
	req := &Package{}
	if err := proto.Unmarshal(msg, req); err != nil {
		return nil
	}

	// req.Name=="" indicates an rpc response
	if req.Name == "" {
		m.stubs[req.Seq](req.Seq, req.Data)
		return nil
	}

	// an request
	payload := m.dispatcher.OnDispatchPackage(req.Name, req.Data)
	if payload != nil {
		resp := &Package{}
		resp.Seq = req.Seq
		resp.Data = payload
		newPayload, _ := proto.Marshal(resp)
		return newPayload
	}
	return nil
}

func (m *MessagerCore) MakeSendPkg(name string, payload []byte) []byte {
	req := &Package{
		Name: name,
		Data: payload,
	}
	data, _ := proto.Marshal(req)
	return data
}

func (m *MessagerCore) MakeCallPkg(name string, payload []byte, cb func(uint32, []byte)) []byte {
	m.seq++
	if m.seq == 0 {
		m.seq++
	}
	req := &Package{
		Name: name,
		Data: payload,
		Seq:  m.seq,
	}
	data, _ := proto.Marshal(req)
	m.stubs[m.seq] = cb
	return data
}

func CreateMessageCore(dispatcher IDispatcher) *MessagerCore {
	return &MessagerCore{
		stubs:      make(map[uint32]func(uint32, []byte)),
		seq:        0,
		dispatcher: dispatcher,
	}
}
