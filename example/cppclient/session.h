#pragma once

#include "../../netlib/netlib.h"
#include "../../netlib/netlib.pb.h"
#include "../../generated/game.arpc.h"

using namespace netlib;
using namespace arpc;

class ClientConnection : public INetConnection, public IPkgMaker
{
public:
    ClientConnection();
    virtual ~ClientConnection();
    void Connect(const char * addr, short port);
    void Update();
    virtual void OnFailed() {}
    virtual void OnRecv(uint8_t *data, int len) = 0;
    virtual void OnConsoleInput(const std::string& cmd) = 0;
    virtual void OnClose();

    virtual void SendPackage(const std::string &msg) override;
    virtual void Close() override {}
	
	virtual std::string MakeSendPkg(const std::string &name, const std::string &content) override;
	virtual std::string MakeCallPkg(const std::string &name, const std::string &content) override;

	bool 	IsActive() const {return isActive;}
private:
    int     sockfd;
    bool    isActive;
};



class ClientSession : public ClientConnection, public GameS2CDispatcher {
public:
	ClientSession(GameS2CImplement *impl):GameS2CDispatcher(impl) {}
    void OnConsoleInput(const std::string& cmd) override;
    virtual void OnRecv(uint8_t *data, int len) override;
private:
	std::string incomeBuffer;
	std::string OnHandlePackage(const std::string& m);
};
