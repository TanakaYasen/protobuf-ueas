
#include <iostream>


#include "wire.cpp"

#include "../generated/addressbook.cpp"

const uint8_t v[] = 
"\x0A\x0D\x70\x72\x6F\x74\x6F\x62\x75\x66\x20\x74\x65\x73\x74\x12\x0E\x08\x0B\x0E\x11\x14\x17\x1A\x1D\x20\x7B\x57\xF1\x85\x79\x19\x43\x01\x00\x00\x00\x00\x00\x00\x25\x2C\x16\x00\x00\x29\x52\xB8\x1E\x85\xEB\x51\x12\x40\x30\x9C\x01\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x3A\x04\x67\x61\x6D\x65\x42\x1D\x08\xFF\x01\x11\xAA\xEE\x00\xFF\x00\x00\x00\x00\x29\x77\xBE\x9F\x1A\x2F\xDD\x5E\x40\x3D\x81\x04\xB5\x3F\x60\xFF\x01"
;

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

	tutorial::Test1 t1(std::string_view((char*)v, sizeof(v)-1));

    return 0;
}