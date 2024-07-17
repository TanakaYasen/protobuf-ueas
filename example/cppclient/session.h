#pragma once

#include "../../netlib/netlib.h"
#include "../../netlib/netlib.pb.h"
#include "../../generated/game.arpc.h"

using namespace netlib;
using namespace arpc;
using std::string;

class ClientConnection : public INetConnection
{
public:
    ClientConnection();
    virtual ~ClientConnection();
    void Connect(const char * addr, short port);
    void Update();
    virtual void OnFailed() {}
    virtual void OnRecv(uint8_t *data, int len) = 0;
    virtual void OnConsoleInput(const string& cmd) = 0;
    virtual void OnClose();

	//override INetConnection
    virtual void SendPackage(const string &msg) override;
    virtual void Close() override {}

	bool 	IsActive() const {return isActive;}
private:
    int     sockfd;
    bool    isActive;
};


class ClientSession : public ClientConnection, public GameS2CDispatcher, public GameC2SHelper<ClientSession> {
	using StubMap = std::unordered_map<uint32_t, std::function<void()>>;
public:
	ClientSession(GameS2CImplement *impl, GameC2SRpcImplement *rpcImpl):GameS2CDispatcher(impl), GameC2SHelper(this, rpcImpl) {}
    void OnConsoleInput(const string& cmd) override;
    virtual void OnRecv(uint8_t *data, int len) override;
	
	string MakeSendPkg(const string &name, const string &content);
	string MakeCallPkg(const string &name, const string &content);
	
private:
	StubMap	stubs;
	string incomeBuffer;
	string OnHandlePackage(const string& m);
	
public:
	virtual void cbDoMovement(uint32_t, const std::string&) override {
	};
};
