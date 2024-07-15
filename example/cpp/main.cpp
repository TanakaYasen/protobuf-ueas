#include<stdio.h>
#include<stdlib.h>
#include<string.h>
#include<errno.h>

#include <fcntl.h>
#include <unistd.h>
#include <sys/stat.h>
#include <sys/wait.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/time.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#define MAXLINE 4096

#include "../../generated/netlib.h"
#include "../../generated/game.arpc.h"

using namespace pb;

class ClientConnection : public INetConnection
{
public:
    ClientConnection() {
        isActive = true;
        if( (sockfd = socket(AF_INET, SOCK_STREAM, 0)) < 0){
            printf("create socket error: %s(errno: %d)\n", strerror(errno),errno);
            return;
        }
        fcntl(sockfd , F_SETFL, O_NONBLOCK);
    }
    virtual ~ClientConnection() {
        if (isActive) {
            close(sockfd);
        }
    }

    void Connect(const char * addr, short port) {
        struct sockaddr_in  servaddr;
        memset(&servaddr, 0, sizeof(servaddr));
        servaddr.sin_family = AF_INET;
        servaddr.sin_port = htons(3389);
        inet_pton(AF_INET, "127.0.0.1", &servaddr.sin_addr);

        int r = connect(sockfd, (struct sockaddr*)&servaddr, sizeof(servaddr));
        if (r < 0) {
            if (errno == EINPROGRESS)
                return;
            printf("connect error: %s(errno: %d)\n",strerror(errno),errno);
            return;
        }
    }

    void Update() {
        uint8_t buf[1024];
        while (isActive) {
            int n = read(sockfd, buf, 1024);
            if (n  == 0) {
                isActive = false;
                OnClose();
            }
            if (n == -1) {
                if (errno == EWOULDBLOCK) {
                    break;
                }
                isActive = false;
                OnClose();
            }
            OnRecv(buf, n);
        }
    }

    virtual void OnFailed() {

    }
    virtual void OnRecv(uint8_t *data, int len) {
    }
    virtual void OnConsoleInput(const std::string& cmd) {
    }

    virtual void OnClose() {
        close(sockfd);
    }

    virtual void SendPackage(const std::string &msg) override {
        char buf[2];
        size_t len = msg.length();
        write(sockfd, buf, 2);
        write(sockfd, msg.c_str(), len);
    }
    virtual void Close() override {

    }

private:
    int     sockfd;
    bool    isActive;
};

class ClientSession : public ClientConnection {
public:
    void OnConsoleInput(const std::string& cmd) override{
        if (cmd == "abc") {
            std::cout << "abc" << std::endl;
        }
    }
};

class DataCenter : public GameS2CImplement  {
	virtual void ItemUpdate(const ItemUpdateReq& req) override {

    }
	virtual void LevelUp(const LevelUpReq& req) override {

    };
};


ClientSession s;

void timer_handler(int sig)
{
    s.Update();
    while (true) {
        char buf[1024];
        int n = read(0, buf, 1024);
        if (n == 0) {
            exit(0);
        }
        if (n == -1) {
            if (errno == EWOULDBLOCK) {
                break;
            }
        }
        std::string cmd(buf, n);
        s.OnConsoleInput(cmd);
    }
}


int main(int argc, char** argv){
    struct timeval tv_interval = {1, 0};
    if (signal(SIGALRM, timer_handler) == SIG_ERR)
        return -1;

    fcntl(0, F_SETFL, fcntl(0, F_GETFL) | O_NONBLOCK);

    struct timeval tv_value = {1, 0};
    struct itimerval it;
    it.it_interval = tv_interval;
    it.it_value = tv_value;
    setitimer(ITIMER_REAL, &it, NULL);

    s.Connect("127.0.0.1", 3389);
    for(;;)
        pause();
    return 0;
}