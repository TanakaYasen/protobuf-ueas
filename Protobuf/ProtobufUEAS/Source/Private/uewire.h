#pragma once

#include "pmessage.h"
// Wire Format
// https://protobuf.dev/programming-guides/encoding/


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


class WireEncoder {
    uint8 *ps, *pe;
    uint8 *pcur;
    TArray<uint8>  _inner;

    union pbfixed64 {
        uint64 i64;
        double  d;
    };

    union pbfixed32 {
        uint32 i32;
        float   f;
    };

    static size_t VarintLen(uint64);

    void CheckSpace(size_t);
    WireEncoder& WriteTag(uint64 fn, WireType wt);
    WireEncoder& WriteVarint(uint64);
    WireEncoder& WriteI32(uint32);
    WireEncoder& WriteI64(uint64);
    WireEncoder& WriteBytes(const uint8*, size_t);

public:
    WireEncoder();

    ustring Dump() const;

    //VARINT
    WireEncoder& EncodeInt32(uint64 fn, int32);
    WireEncoder& EncodeSint32(uint64 fn, int32);
    WireEncoder& EncodeUint32(uint64 fn, uint32);
    WireEncoder& EncodeInt64(uint64 fn, int64);
    WireEncoder& EncodeSint64(uint64 fn, int64);
    WireEncoder& EncodeUint64(uint64 fn, uint64);
    WireEncoder& EncodeBool(uint64 fn, bool);
    WireEncoder& EncodeEnum(uint64 fn, unsigned);

    //LEN
    WireEncoder& EncodeString(uint64 fn, const std::string&);
    WireEncoder& EncodeBytes(uint64 fn, const TArray<uint8>&);
    template <typename T>
    WireEncoder& EncodeSubmessage(uint64 fn, const T&msg) {
        // static_assert(std::is_base_of_v<T, >)
        return EncodeString(fn, msg.Serialize());
    }

    //I32
    WireEncoder& EncodeSfixed32(uint64 fn, int32);
    WireEncoder& EncodeFixed32(uint64 fn, uint32);
    WireEncoder& EncodeFloat(uint64 fn, float);

    //I64
    WireEncoder& EncodeSfixed64(uint64 fn, int64);
    WireEncoder& EncodeFixed64(uint64 fn, uint64);
    WireEncoder& EncodeDouble(uint64 fn, double);

    //reps
    WireEncoder& EncodeRepBool(uint64 fn, const TArray<bool>&);
    WireEncoder& EncodeRepSfixed32(uint64 fn, const TArray<int32>&);
    WireEncoder& EncodeRepFixed32(uint64 fn, const TArray<uint32>&);
    WireEncoder& EncodeRepFloat(uint64 fn, const TArray<float>&);
    WireEncoder& EncodeRepSfixed64(uint64 fn, const TArray<int64>&);
    WireEncoder& EncodeRepFixed64(uint64 fn, const TArray<uint64>&);
    WireEncoder& EncodeRepDouble(uint64 fn, const TArray<double>&);

    WireEncoder& EncodeRepInt32(uint64 fn, const TArray<int32>&);
    WireEncoder& EncodeRepSint32(uint64 fn, const TArray<int32>&);
    WireEncoder& EncodeRepUint32(uint64 fn, const TArray<uint32>&);
    WireEncoder& EncodeRepInt64(uint64 fn, const TArray<int64>&);
    WireEncoder& EncodeRepSint64(uint64 fn, const TArray<int64>&);
    WireEncoder& EncodeRepUint64(uint64 fn, const TArray<uint64>&);
    WireEncoder& EncodeRepString(uint64 fn, TArray<ustring>&vs);
};


class WireDecoder {
    const uint8  *ps, *pe;
    bool    valid = true;
    WireType wt;

    union pbfixed64 {
        uint64 i64;
        double  d;
    };

    union pbfixed32 {
        uint32 i32;
        float   f;
    };

    pbfixed32 ReadFixed32();
    pbfixed64 ReadFixed64();

    uint64 ReadVarint();
    std::string ReadString(uint64 len);
    std::string_view ReadLength(uint64 len);
    TArray<uint8> ReadBytes(uint64 len);
    std::string ReadMessage(uint64 len);

public:
    WireDecoder(const uint8 *data, size_t len);

    bool IsOk() const {return valid;}

    uint64 ReadTag();

    //VARINT
    int32 DecodeInt32();
    int32 DecodeSint32();
    int64 DecodeInt64();
    int64 DecodeSint64();
    uint32 DecodeUint32();
    uint64 DecodeUint64();
    bool    DecodeBool();
    unsigned DecodeEnum();

    //I32
    int32 DecodeSfixed32();
    uint32 DecodeFixed32();
    float   DecodeFloat();

    //LEN
    std::string DecodeString();
    TArray<uint8> DecodeBytes();
    std::string_view DecodeSubmessage();

    //I64
    int64     DecodeSfixed64();
    uint64    DecodeFixed64();
    double      DecodeDouble();

    //skip
    void        DecodeUnknown();

    //reps
#define DecodeRepDecl(suf, ct) void DecodeRep##suf(TArray<ct>&)
    DecodeRepDecl(Bool, bool);
    DecodeRepDecl(Int32, int32);
    DecodeRepDecl(Int64, int64);
    DecodeRepDecl(Uint32, uint32);
    DecodeRepDecl(Uint64, uint64);
    DecodeRepDecl(Sint32, int32);
    DecodeRepDecl(Sint64, int64);
    DecodeRepDecl(Sfixed32, int32);
    DecodeRepDecl(Fixed32, uint32);
    DecodeRepDecl(Float, float);
    DecodeRepDecl(Sfixed64, int64);
    DecodeRepDecl(Fixed64, uint64);
    DecodeRepDecl(Double, double);
#undef DecodeRepDecl
    void DecodeRepString(TArray<ustring>&);
    template <typename T>
    void DecodeRepSubmessage(TArray<T>& values) {
        ustringview v = DecodeSubmessage();
        if (!valid) return;
        T msg(v);
        if (msg.IsValid()) {
            values.EmplaceBack(MoveTemp(msg));
            return;
        }
        valid = false;
    }
};

