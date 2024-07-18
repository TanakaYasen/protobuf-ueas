#include<stdio.h>
#include<stdlib.h>

#include <fcntl.h>
#include <sys/wait.h>
#include <sys/time.h>

#include "../../generated/game.pb.h"
#include "../../generated/game.arpc.h"
#include "session.h"

using namespace arpc;

class ClientStub : public GameS2CImplement  {
public:
	virtual void ItemUpdate(const ItemUpdateReq& req) override {
    }
	virtual void LevelUp(const LevelUpReq& req) override {
    };
};

class ClientRpcCallbakcs : public GameC2SRpcImplement {
public:
	virtual void DoMovement(uint32_t seq , const MoveResp &resp) override {
		std::cout << resp.x() << "," << resp.y() << "," << resp.z() << std::endl;
	}
};

ClientSession s(new ClientStub, new ClientRpcCallbakcs);

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
        std::string cmd(buf, n-1);
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
    for(;s.IsActive();)
        pause();
    return 0;
}
