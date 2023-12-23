
#include <iostream>


#include "wire.cpp"

#include "../generated/addressbook.cpp"

const uint8_t v[] = 
"\x0A\x04\x6E\x61\x6D\x65\x10\xFA\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF\x01\x19\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF\x25\xFF\xFF\xFF\xFF\x29\x1F\x85\xEB\x51\xB8\x1E\x09\x40\x30\x9C\x01";


//#include "../generated/addressbook.h"

void DumpHex(const void* data, size_t size) {
	char ascii[17];
	size_t i, j;
	ascii[16] = '\0';
	for (i = 0; i < size; ++i) {
		printf("%02X ", ((unsigned char*)data)[i]);
		if (((unsigned char*)data)[i] >= ' ' && ((unsigned char*)data)[i] <= '~') {
			ascii[i % 16] = ((unsigned char*)data)[i];
		} else {
			ascii[i % 16] = '.';
		}
		if ((i+1) % 8 == 0 || i+1 == size) {
			printf(" ");
			if ((i+1) % 16 == 0) {
				printf("|  %s \n", ascii);
			} else if (i+1 == size) {
				ascii[(i+1) % 16] = '\0';
				if ((i+1) % 16 <= 8) {
					printf(" ");
				}
				for (j = (i+1) % 16; j < 16; ++j) {
					printf("   ");
				}
				printf("|  %s \n", ascii);
			}
		}
	}
}
void DumpHex(const std::string &s) {
    DumpHex(s.c_str(), s.size());
}

int main() {

    tutorial::Test1 t1, t2;
    t1.name_ = "protobuf test";
    t1.f32_ = 5676;
    t1.money_ = 4.58;
    t1.f64_ = 323;
    t1.en_ = 78;
	for (int i = 1; i < 10; i++) {
		t1.age_.push_back(i*3+5);
	}
	t1.age_.push_back(123);
	t1.age_.push_back(87);
	t1.age_.push_back(1983217);
	t1.test2_.i32_=255;
	t1.test2_.u64_=255;
	t1.test2_.d_=123.456;
	t1.test2_.f_=1.4142;
	t1.test2_.fix_=0xff00eeaa;

	for (int i = 0; i < 2; ++i)
		t1.alias_.push_back("game");

    auto sss = t1.Serialize();
    DumpHex(sss);

    t2.Unserialize(std::string_view{(char*)sss.c_str(), sss.size()});

	for (auto &v : t2.age_) {
		std::cout << v << std::endl;
	}
	for (auto &v : t2.alias_) {
		std::cout << v << std::endl;
	}
    return 0;
}