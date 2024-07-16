#pragma once

#include <string>

using std::string;

class INetConnection {
public:
    virtual void SendPackage(const string &)=0;
    virtual void Close()=0;
};

class IPkgMaker {
public:
    virtual string MakeSendPkg(const string &name, const string &content) = 0;
    virtual string MakeCallPkg(const string &name, const string &content) = 0;
};

