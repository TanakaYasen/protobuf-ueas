
#include <iostream>

#include "wire.cpp"

const uint8_t v[] = 
"\x0A\x04\x6E\x61\x6D\x65\x10\xFA\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF\x01\x19\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF\x25\xFF\xFF\xFF\xFF\x29\x1F\x85\xEB\x51\xB8\x1E\x09\x40\x30\x9C\x01";


//#include "../generated/addressbook.h"

class Test1 {
public:
    std::string name_;
    int64_t age_;
    uint64_t f64_;
    uint32_t f32_;
    double money_;
    int64_t en_;
public:
    std::string Serialize() const;
    bool Unserialize(std::string_view sv)
    {
        uint64_t fn;
        WireDecoder decoder((const uint8_t*)sv.data(), sv.length());
        while ((fn = decoder.ReadTag()) && decoder.IsOk()) {
            switch(fn) {
                case 1:
                    this->name_ = decoder.DecodeString();
                break;
                case 2:
                    this->age_ = decoder.DecodeInt64();
                break;
                case 3:
                    this->f64_ = decoder.DecodeFixed64();
                break;
                case 4:
                    this->f32_ = decoder.DecodeFixed32();
                break;
                case 5:
                    this->money_ = decoder.DecodeDouble();
                break;
                case 6:
                    this->en_ = decoder.DecodeSint64();
                break;
                default:
                    decoder.DecodeUnknown();
            }
        }
        return decoder.IsOk();
    }
};


int main() {
    uint64_t fn;

    Test1 t1;
    t1.Unserialize(std::string_view{(char*)v, sizeof(v)-1});

    std::cout << std::endl << t1.name_
        << std::endl << t1.age_
        << std::endl << t1.f64_
        << std::endl << t1.f32_
        << std::endl << t1.money_
        << std::endl << t1.en_
        << std::endl;

    return 0;
}