#pragma once

#include <cstdlib>
#include <cstdint>
#include <string>
#include <vector>

// Wire Format
// https://protobuf.dev/programming-guides/encoding/
// https://blog.51cto.com/u_6526235/7273612


// Message Structure 
// message = recode[]
// recode = field number + wiretype + payload

enum WireType {
    WT_VARINT = 0,	//int32, int64, uint32, uint64, sint32, sint64, bool, enum
    WT_I64	= 1,    // fixed64, sfixed64, double
    WT_LEN = 2,     //	string, bytes, embedded messages, packed repeated fields
    WT_SGROUP = 3, //	group start (deprecated)
    WT_EGROUP = 4, //	group end (deprecated)
    WT_I32 = 5,     //	fixed32, sfixed32, float
};

// Length-Delimited Records


class WireEncoder {
    uint8_t *ps, *pe;
    uint8_t *pcur;
    std::vector<uint8_t>  _inner;

    union pbfixed64 {
        uint64_t i64;
        double  d;
    };

    union pbfixed32 {
        uint32_t i32;
        float   f;
    };

    WireEncoder& WriteTag(uint64_t fn, WireType wt);
    WireEncoder& WriteVarint(uint64_t);
    WireEncoder& WriteI32(uint32_t);
    WireEncoder& WriteI64(uint64_t);

    void CheckSpace(size_t);

public:
    WireEncoder();

    //VARINT
    WireEncoder& EncodeInt32(uint64_t fn, int32_t);
    WireEncoder& EncodeSint32(uint64_t fn, int32_t);
    WireEncoder& EncodeUint32(uint64_t fn, uint32_t);
    WireEncoder& EncodeInt64(uint64_t fn, int64_t);
    WireEncoder& EncodeSint64(uint64_t fn, int64_t);
    WireEncoder& EncodeUint64(uint64_t fn, uint64_t);
    WireEncoder& EncodeBool(uint64_t fn, bool);
    WireEncoder& EncodeEnum(uint64_t fn, unsigned);

    //LEN
    WireEncoder& EncodeString(uint64_t fn, const std::string&);
    WireEncoder& EncodeBytes(uint64_t fn, const std::vector<uint8_t>&);
    template <typename T>
    WireEncoder& EncodeSubmessage(uint64_t fn, const T&msg) {
        // static_assert(std::is_base_of_v<T, >)
        return EncodeString(msg.Serialize());
    }

    //I32
    WireEncoder& EncodeSfixed32(uint64_t fn, int32_t);
    WireEncoder& EncodeFixed32(uint64_t fn, uint32_t);
    WireEncoder& EncodeFloat(uint64_t fn, float);

    //I64
    WireEncoder& EncodeSfixed64(uint64_t fn, int64_t);
    WireEncoder& EncodeFixed64(uint64_t fn, uint64_t);
    WireEncoder& EncodeDouble(uint64_t fn, double);

    //reps
    WireEncoder& EncodeRepSfixed32(uint64_t fn, const std::vector<int32_t>&);
    WireEncoder& EncodeRepFixed32(uint64_t fn, const std::vector<uint32_t>&);
    WireEncoder& EncodeRepFloat(uint64_t fn, const std::vector<float>&);
    WireEncoder& EncodeRepSfixed64(uint64_t fn, const std::vector<int64_t>&);
    WireEncoder& EncodeRepFixed64(uint64_t fn, const std::vector<uint64_t>&);
    WireEncoder& EncodeRepDouble(uint64_t fn, const std::vector<double>&);

    WireEncoder& EncodeRepInt32(uint64_t fn, const std::vector<int32_t>&);
    WireEncoder& EncodeRepSint32(uint64_t fn, const std::vector<int32_t>&);
    WireEncoder& EncodeRepUint32(uint64_t fn, const std::vector<uint32_t>&);
    WireEncoder& EncodeRepInt64(uint64_t fn, const std::vector<int64_t>&);
    WireEncoder& EncodeRepSint64(uint64_t fn, const std::vector<int64_t>&);
    WireEncoder& EncodeRepUint64(uint64_t fn, const std::vector<uint64_t>&);

};


class WireDecoder {
    const uint8_t  *ps, *pe;
    bool    valid = true;
    WireType wt;

    union pbfixed64 {
        uint64_t i64;
        double  d;
    };

    union pbfixed32 {
        uint32_t i32;
        float   f;
    };

    pbfixed32 ReadFixed32();
    pbfixed64 ReadFixed64();

    uint64_t ReadVarint();
    std::string ReadString(uint64_t len);
    std::string_view ReadLength(uint64_t len);
    std::vector<uint8_t> ReadBytes(uint64_t len);
    std::string ReadMessage(uint64_t len);

public:
    WireDecoder(const uint8_t*data, size_t len);

    bool IsOk() const {return valid;}

    uint64_t ReadTag();

    //VARINT
    int32_t DecodeInt32();
    int32_t DecodeSint32();
    int64_t DecodeInt64();
    int64_t DecodeSint64();
    uint32_t DecodeUint32();
    uint64_t DecodeUint64();
    bool    DecodeBool();
    unsigned DecodeEnum();

    //I32
    int32_t DecodeSfixed32();
    uint32_t DecodeFixed32();
    float   DecodeFloat();

    //LEN
    std::string DecodeString();
    std::vector<uint8_t> DecodeBytes();
    std::string_view DecodeSubmessage();

    //I64
    int64_t     DecodeSfixed64();
    uint64_t    DecodeFixed64();
    double      DecodeDouble();

    //reps
    std::vector<int32_t> DecodeRepInt32();
    std::vector<int64_t> DecodeRepInt64();
    std::vector<uint32_t> DecodeRepUint32();
    std::vector<uint64_t> DecodeRepUint64();
    std::vector<int32_t> DecodeRepSint32();
    std::vector<int64_t> DecodeRepSint64();

    std::vector<int32_t>   DecodeRepSfix32();
    std::vector<uint32_t>   DecodeRepFix32();
    std::vector<float>   DecodeRepFloat();
    std::vector<int64_t>   DecodeRepSfix64();
    std::vector<uint64_t>   DecodeRepFix64();
    std::vector<double>   DecodeRepDouble();

    void        DecodeUnknown();
};

