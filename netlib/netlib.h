#pragma once

#include <string>

using std::string;

class INetConnection {
public:
    virtual void SendPackage(const string &)=0;
    virtual void Close()=0;
};

