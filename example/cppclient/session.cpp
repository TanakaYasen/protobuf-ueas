
#include "session.h"


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

#include "../../netlib/netlib.h"
#include "../../netlib/netlib.pb.h"

using namespace netlib;

ClientConnection::ClientConnection() {
	isActive = true;
	if( (sockfd = socket(AF_INET, SOCK_STREAM, 0)) < 0){
		printf("create socket error: %s(errno: %d)\n", strerror(errno),errno);
		return;
	}
	fcntl(sockfd , F_SETFL, O_NONBLOCK);
}
ClientConnection::~ClientConnection() {
	if (isActive) {
		close(sockfd);
	}
}

void ClientConnection::Connect(const char * addr, short port) {
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

void ClientConnection::Update() {
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


void ClientConnection::OnClose() {
	close(sockfd);
}

void ClientConnection::SendPackage(const std::string &msg) {
	size_t len = msg.length();
	uint8_t buf[2] = {uint8_t(len >> 8), uint8_t(len&0xff)};
	write(sockfd, buf, 2);
	write(sockfd, msg.c_str(), len);
}

std::string ClientConnection::MakeSendPkg(const std::string &name, const std::string &content) {
	netlib::Package pkg;
	pkg.set_data(content);
	pkg.set_name(name);
	return pkg.SerializeAsString();
}

std::string ClientConnection::MakeCallPkg(const std::string &name, const std::string &content) {
	netlib::Package pkg;
	pkg.set_data(content);
	pkg.set_name(name);
	return pkg.SerializeAsString();
}


extern GameC2SMessager *m;

void ClientSession :: OnConsoleInput(const std::string& cmd) {
	static int sceneId = 0;
	if (cmd == "abc") {
		std::cout << "abc" << std::endl;
		arpc::EnterSceneReq esreq;
		esreq.set_sceneid(sceneId++);
		m->SendEnterScene(esreq);
	}
}

	void ClientSession ::OnRecv(uint8_t *data, int len)  {
	incomeBuffer += std::string((char*)data, len);
	for (;;) {
		if (incomeBuffer.size() < 2) {
			return;
		}
		int pkglen = int(incomeBuffer[0])*256 + int(incomeBuffer[1]);
		if (incomeBuffer.size() < 2 + pkglen) {
			return;
		}
		std::string data = OnHandlePackage(incomeBuffer.substr(0, 2+pkglen));
		if (data.size() > 0) {
			SendPackage(data);
		}
		incomeBuffer = incomeBuffer.substr(2+pkglen);
	}
}
std::string ClientSession ::OnHandlePackage(const std::string& m) {
	Package req;
	if (!req.ParseFromString(m)) {
		return "";
	}

	// req.Name=="" indicates an rpc response
	if (req.name() == "") {
		return "";
	}

	// an request
	std::string payload = OnDispatchPackage(req.name(), req.data());
	if (payload.length() > 0) {
		SendPackage(payload);
	}
	return "";
}
