#include "wire.h"
#include <cstdlib>
#include <cstring>

static inline uint32_t read_dword(const uint8_t *data)
{
    uint32_t u;
    u = data[0] | (data[1] << 8) | (data[2] << 16) | (data[3] << 24);
    return u;
}

static inline uint64_t read_qword(const uint8_t *data)
{
    uint64_t high = data[4] | (data[5] << 8) | (data[6] << 16) | (data[7] << 24);
    uint32_t low = data[0] | (data[1] << 8) | (data[2] << 16) | (data[3] << 24);
    return (high<<32) | low;
}

static inline uint8_t *write_dword(uint8_t data[], uint32_t v)
{
    data[0] = v & 0xff;
    data[1] = (v>>8) & 0xff;
    data[2] = (v>>16) & 0xff;
    data[3] = (v>>24) & 0xff;
    return &data[4];
}

static inline uint8_t *write_qword(uint8_t data[], uint64_t v)
{
    data[0] = v & 0xff;
    data[1] = (v>>8) & 0xff;
    data[2] = (v>>16) & 0xff;
    data[3] = (v>>24) & 0xff;
    data[4] = (v>>32) & 0xff;
    data[5] = (v>>40) & 0xff;
    data[6] = (v>>48) & 0xff;
    data[7] = (v>>56) & 0xff;
    return &data[8];
}


#define KeepCV(x) (x)
#define ZipZagE(x)   (x<<1)^(x>>(8*sizeof(x)-1))
#define ZipZagD(x)   ((x>>1)^(-(x&1)))

#define EncodeRepTempl(suf, ct, CV, writer) \
WireEncoder& WireEncoder::EncodeRep##suf(uint64_t fn, const std::vector<ct>&vs) {\
    if (vs.size() == 0) return *this; \
    WriteTag(fn, WT_LEN); \
    off64_t offold = pcur-ps; \
    for (auto v : vs) { \
        auto q = CV(v); \
        writer; \
    } \
    size_t totallen = pcur-(ps+offold); \
    size_t lenvarint = VarintLen(totallen); \
    uint8_t *pold = ps+offold;\
    memmove(pold+lenvarint, pold, totallen); \
    pcur = pold; WriteVarint(totallen); /*patch back*/\
    pcur += totallen; \
    return *this; \
}

#define DecodeRepTempl(suf, ct, w, CV, reader) \
void WireDecoder::DecodeRep##suf(std::vector<ct> &values) { \
    if (wt == w) { \
        ct v = (ct)reader; \
        values.push_back(CV(v)); \
    } \
    else if (wt == WT_LEN) { \
        uint64_t len = ReadVarint(); \
        if (!valid) { \
            return;\
        } \
        const uint8_t* sential = ps + len; \
        while (ps < sential) \
        { \
            ct v = (ct)reader; \
            values.push_back(CV(v)); \
        } \
    } \
}


WireEncoder::WireEncoder() {
    _inner.resize(1024);
    pcur = ps = &_inner[0];
    pe = ps + _inner.size();
}

std::string WireEncoder::Dump() const {
    return std::string{(const char*)&_inner[0], (size_t)(pcur-ps)};
}


void WireEncoder::CheckSpace(size_t sz) {
    if (pcur + sz <= pe) { return; }
    
    size_t cap = _inner.size();
    while (pcur + sz > pe) {
        cap *= 2;
        pe = ps + cap;
    }

    _inner.resize(cap);
    pcur = &_inner[0] + (pcur-ps);
    ps = &_inner[0];
    pe = ps + cap;
}

WireEncoder& WireEncoder::WriteTag(uint64_t fn, WireType wt){
    uint64_t tag = (fn << 3) | wt;
    return WriteVarint(tag);
}
WireEncoder& WireEncoder::WriteVarint(uint64_t v){
    CheckSpace(10);
    while (v > 0x7f) {
        *pcur++ = (v & 0x7f) | 0x80;
        v >>= 7;
    }
    *pcur++ = v & 0x7f;
    return *this;
}

size_t WireEncoder::VarintLen(uint64_t v){
    size_t off = 0;
    while (v > 0x7f) {
        ++off; v >>= 7;
    }
    ++off;
    return off;
}

WireEncoder& WireEncoder::WriteI32(uint32_t v){
    CheckSpace(4);
    pcur = write_dword(pcur, v);
    return *this;
}
WireEncoder& WireEncoder::WriteI64(uint64_t v){
    CheckSpace(8);
    pcur = write_qword(pcur, v);
    return *this;
}

//VARINT
WireEncoder& WireEncoder::EncodeInt32(uint64_t fn, int32_t v){
    CheckSpace(10+10);
    WriteTag(fn, WT_VARINT);
    WriteVarint(v);
    return *this;
}
WireEncoder& WireEncoder::EncodeSint32(uint64_t fn, int32_t v){
    CheckSpace(10+10);
    WriteTag(fn, WT_VARINT);
    uint32_t q = ZipZagE(v);
    WriteVarint(q);
    return *this;
}
WireEncoder& WireEncoder::EncodeUint32(uint64_t fn, uint32_t v){
    CheckSpace(10+10);
    WriteTag(fn, WT_VARINT);
    WriteVarint(v);
    return *this;
}
WireEncoder& WireEncoder::EncodeInt64(uint64_t fn, int64_t v){
    CheckSpace(10+10);
    WriteTag(fn, WT_VARINT);
    WriteVarint(v);
    return *this;
}
WireEncoder& WireEncoder::EncodeSint64(uint64_t fn, int64_t v){
    CheckSpace(10+10);
    WriteTag(fn, WT_VARINT);
    uint64_t q = ZipZagE(v);
    WriteVarint(q);
    return *this;
}
WireEncoder& WireEncoder::EncodeUint64(uint64_t fn, uint64_t v){
    CheckSpace(10+10);
    WriteTag(fn, WT_VARINT);
    WriteVarint(v);
    return *this;
}
WireEncoder& WireEncoder::EncodeBool(uint64_t fn, bool v){
    CheckSpace(10+10);
    WriteTag(fn, WT_VARINT);
    WriteVarint((uint64_t)v);
    return *this;
}


//LEN
WireEncoder& WireEncoder::EncodeString(uint64_t fn, const std::string&s){
    size_t len = s.length();
    CheckSpace(10+10+len);
    WriteTag(fn, WT_LEN);
    WriteVarint(len);
    ::memcpy(pcur, (uint8_t*)s.c_str(), len);
    pcur += len;
    return *this;
}
WireEncoder& WireEncoder::EncodeBytes(uint64_t fn, const std::vector<uint8_t>&v){
    size_t len = v.size();
    CheckSpace(10+10+len);
    WriteTag(fn, WT_LEN);
    WriteVarint(len);
    ::memcpy(pcur, (uint8_t*)&v[0], len);
    pcur += len;
    return *this;
}

//I32
WireEncoder& WireEncoder::EncodeSfixed32(uint64_t fn, int32_t v){
    CheckSpace(10+4);
    WriteTag(fn, WT_I32);
    WriteI32((uint32_t)v);
    return *this;
}
WireEncoder& WireEncoder::EncodeFixed32(uint64_t fn, uint32_t v){
    CheckSpace(10+4);
    WriteTag(fn, WT_I32);
    WriteI32(v);
    return *this;
}
WireEncoder& WireEncoder::EncodeFloat(uint64_t fn, float v){
    CheckSpace(10+4);
    WriteTag(fn, WT_I32);
    pbfixed32 u;
    u.f = v;
    WriteI32(u.i32);
    return *this;
}

//I64
WireEncoder& WireEncoder::EncodeSfixed64(uint64_t fn, int64_t v){
    CheckSpace(10+8);
    WriteTag(fn, WT_I64);
    WriteI64((uint64_t)v);
    return *this;
}
WireEncoder& WireEncoder::EncodeFixed64(uint64_t fn, uint64_t v){
    CheckSpace(10+8);
    WriteTag(fn, WT_I64);
    WriteI64(v);
    return *this;
}
WireEncoder& WireEncoder::EncodeDouble(uint64_t fn, double v){
    CheckSpace(10+8);
    WriteTag(fn, WT_I64);
    pbfixed64 u;
    u.d = v;
    WriteI64(u.i64);
    return *this;
}


EncodeRepTempl(Sfixed32, int32_t, KeepCV, pbfixed32 u; u.i32=q; WriteI32(u.i32))
EncodeRepTempl(Fixed32, uint32_t, KeepCV, pbfixed32 u; u.i32=q; WriteI32(u.i32))
EncodeRepTempl(Float, float, KeepCV, pbfixed32 u; u.f=q; WriteI32(u.i32))
EncodeRepTempl(Sfixed64, int64_t, KeepCV, pbfixed64 u; u.i64=q; WriteI64(u.i64))
EncodeRepTempl(Fixed64, uint64_t, KeepCV, pbfixed64 u; u.i64=q; WriteI64(u.i64))
EncodeRepTempl(Double, double, KeepCV, pbfixed64 u; u.d=q; WriteI32(u.i64))

EncodeRepTempl(Int64, int64_t, KeepCV, WriteVarint(q))
EncodeRepTempl(Uint64, uint64_t, KeepCV, WriteVarint(q))
EncodeRepTempl(Int32, int32_t, KeepCV, WriteVarint(q))
EncodeRepTempl(Uint32, uint32_t, KeepCV, WriteVarint(q))
EncodeRepTempl(Sint64, int64_t, ZipZagE, WriteVarint(q))
EncodeRepTempl(Sint32, int32_t, ZipZagE, WriteVarint(q))

WireDecoder::WireDecoder(const uint8_t *data, size_t len)
{
    ps = data;
    pe = data + len;
}

uint64_t WireDecoder::ReadTag() {
    uint64_t tag = 0;
    while (ps < pe && ps[0] & 0x80) {
        tag |= ps[0] & 0x7f;
        tag <<= 7;
        ps++;
    }
    if (ps < pe) {
        tag |= ps[0] & 0x7f;
        wt = (WireType)(tag & 0x7);
        ps++;
        return tag >> 3;
    }
    valid = false;
    return 0;
}

std::string WireDecoder::ReadString(size_t len) {
    if (ps + len <= pe) 
    {
        std::string tmp((char*)ps, len);
        ps += len;
        return tmp;
    }
    valid = false;
    return std::string{};
}
std::vector<uint8_t> WireDecoder::ReadBytes(size_t len) {
    if (ps + len <= pe) 
    {
        std::vector<uint8_t> res(len);
        ::memcpy(&res[0], ps, len);
        ps+=len;
        return res;
    }
    valid = false;
    return std::vector<uint8_t>{};
}
std::string_view WireDecoder::ReadLength(size_t len) {
    if (ps + len <= pe) 
    {
        std::string_view tmp{(char*)ps, len};
        ps += len;
        return tmp;
    }
    valid = false;
    return std::string_view{};
}
uint64_t WireDecoder::ReadVarint() {
    uint64_t varint = 0;
    const uint8_t *oldps = ps;
    while (ps < pe && ps[0] & 0x80) {
        ps++;
    }
    if (ps < pe) {
        ps++;
        const uint8_t *c = ps;
        while (--c>=oldps) {
            varint <<=7;
            varint |= c[0] & 0x7f;
        }
        return varint;
    }
    valid = false;
    return 0;
}
WireDecoder::pbfixed32 WireDecoder::ReadFixed32() {
    pbfixed32 u{0};
    if (ps + sizeof(pbfixed32) <= pe) 
    {
        u.i32 = read_dword(ps);
        ps += sizeof(WireDecoder::pbfixed32);
        return u;
    }
    valid = false;
    return u;
}
WireDecoder::pbfixed64 WireDecoder::ReadFixed64() {
    pbfixed64 u{0};
    if (ps + sizeof(pbfixed64) <= pe) 
    {
        u.i64 = read_qword(ps);
        ps += sizeof(WireDecoder::pbfixed64);
        return u;
    }
    valid = false;
    return u;
}


int32_t WireDecoder::DecodeInt32() {
    switch (wt) {
    case WT_VARINT:
        return (int32_t)ReadVarint();
    case WT_I64:    // fixed64, sfixed64, double
        return (int32_t)ReadFixed64().i64;
    case WT_I32:
        return ReadFixed32().i32;
    default:
        break;
    }
    return 0;
}

int32_t WireDecoder::DecodeSint32() {
    switch (wt) {
    case WT_VARINT: {
        uint32_t v = (uint32_t)ReadVarint();
        return (v>>1)^-(v&1);
    }
    default:
        break;
    }
    return 0;
}

int64_t WireDecoder::DecodeSint64() {
    switch (wt) {
    case WT_VARINT: {
        int64_t v = ReadVarint();
        return (v>>1)^-(v&1);
    }
    default:
        break;
    }
    return 0;
}

uint32_t WireDecoder::DecodeUint32() {
    switch (wt) {
    case WT_VARINT:
        return (uint32_t)ReadVarint();
    case WT_I64:    // fixed64, sfixed64, double
        return (uint32_t)ReadFixed64().i64;
    case WT_I32:
        return ReadFixed32().i32;
    default:
        break;
    }
    return 0;
}

int64_t WireDecoder::DecodeInt64() {
    return (int64_t)DecodeUint64();
}

uint64_t WireDecoder::DecodeUint64() {
    switch (wt) {
    case WT_VARINT:
        return ReadVarint();
    case WT_I64:    // fixed64, sfixed64, double
        return ReadFixed64().i64;
    case WT_I32:
        return ReadFixed32().i32;
    default:
        break;
    }
    return 0;
}

bool WireDecoder::DecodeBool() {
    if (wt == WT_VARINT)
        return (bool)ReadVarint();
    return 0;
}

unsigned WireDecoder::DecodeEnum() {
    if (wt == WT_VARINT)
        return (unsigned)ReadVarint();
    return 0;
}


double WireDecoder::DecodeDouble() {
    switch (wt) {
    case WT_I64:    // fixed64, sfixed64, double
        return ReadFixed64().d;
    default:
        break;
    }
    return 0.0;
}

float WireDecoder::DecodeFloat() {
    switch (wt) {
    case WT_I32:    // fixed64, sfixed64, double
        return ReadFixed32().f;
    default:
        break;
    }
    return 0.0f;
}

#if 0
std::vector<int32_t> WireDecoder::DecodeRepSf32() {
    std::vector<int32_t> values;
    if (wt == WT_I32) {
        values.push_back((int32_t)ReadFixed32().i32);
    }
    else if (wt == WT_LEN) {
        uint64_t len = ReadVarint();
        if (!valid) {
            return values;
        }
        const uint8_t* sential = ps + len;
        while (ps < sential) 
        {
            values.push_back((int32_t)ReadFixed32().i32);
        }
    }
    return values;
}

std::vector<uint32_t> WireDecoder::DecodeRepF32() {
    std::vector<uint32_t> values;
    if (wt == WT_I32) {
        values.push_back(ReadFixed32().i32);
    }
    else if (wt == WT_LEN) {
        uint64_t len = ReadVarint();
        if (!valid) {
            return values;
        }
        const uint8_t* sential = ps + len;
        while (ps < sential) 
        {
            values.push_back(ReadFixed32().i32);
        }
    }
    return values;
}

std::vector<float> WireDecoder::DecodeRepFloat() {
    std::vector<float> values;
    if (wt == WT_I32) {
        values.push_back(ReadFixed32().f);
    }
    else if (wt == WT_LEN) {
        uint64_t len = ReadVarint();
        if (!valid) {
            return values;
        }
        const uint8_t* sential = ps + len;
        while (ps < sential) 
        {
            values.push_back(ReadFixed32().f);
        }
    }
    return values;
}
#endif

int64_t WireDecoder::DecodeSfixed64() {
    return (int64_t)DecodeFixed64();
}
uint64_t WireDecoder::DecodeFixed64() {
    switch (wt) {
    case WT_VARINT:
        return ReadVarint();
    case WT_I64:    // fixed64, sfixed64, double
        return ReadFixed64().i64;
    case WT_I32:
        return ReadFixed32().i32;
    default:
        break;
    }
    return 0;
}
int32_t WireDecoder::DecodeSfixed32() {
    return (int32_t)DecodeFixed32();
}
uint32_t WireDecoder::DecodeFixed32() {
    switch (wt) {
    case WT_VARINT:
        return (uint32_t)ReadVarint();
    case WT_I64:    // fixed64, sfixed64, double
        return (uint32_t)ReadFixed64().i64;
    case WT_I32:
        return ReadFixed32().i32;
    default:
        break;
    }
    return 0;
}

DecodeRepTempl(Sfixed32, int32_t, WT_I32, KeepCV, ReadFixed32().i32)
DecodeRepTempl(Fixed32, uint32_t, WT_I32, KeepCV, ReadFixed32().i32)
DecodeRepTempl(Float, float,    WT_I32, KeepCV, ReadFixed32().f)
DecodeRepTempl(Sfixed64, int64_t, WT_I64, KeepCV, ReadFixed64().i64)
DecodeRepTempl(Fixed64, uint64_t, WT_I64, KeepCV, ReadFixed64().i64)
DecodeRepTempl(Double, double,  WT_I64, KeepCV, ReadFixed64().d)

DecodeRepTempl(Bool, bool, WT_VARINT, KeepCV, ReadVarint())
DecodeRepTempl(Int64, int64_t,  WT_VARINT, KeepCV, ReadVarint())
DecodeRepTempl(Uint64, uint64_t,  WT_VARINT, KeepCV, ReadVarint())
DecodeRepTempl(Int32, int32_t,  WT_VARINT, KeepCV, ReadVarint())
DecodeRepTempl(Uint32, uint32_t,  WT_VARINT, KeepCV, ReadVarint())
DecodeRepTempl(Sint64, int64_t,  WT_VARINT, ZipZagD, ReadVarint())
DecodeRepTempl(Sint32, int32_t,  WT_VARINT, ZipZagD, ReadVarint())

std::string WireDecoder::DecodeString() {
    switch (wt) {
    case WT_LEN:
    {
        uint64_t len = ReadVarint();
        if (valid) 
        {
            return ReadString(len);
        }
    }
    break;
    default:
        break;
    }
    return std::string{};
}

std::vector<uint8_t> WireDecoder::DecodeBytes() {
    switch (wt) {
    case WT_LEN:
    {
        uint64_t len = ReadVarint();
        if (valid) 
        {
            return ReadBytes(len);
        }
    }
    break;
    default:
        break;
    }
    return std::vector<uint8_t>{};
}

std::string_view WireDecoder::DecodeSubmessage() {
    switch (wt) {
    case WT_LEN:
    {
        uint64_t len = ReadVarint();
        if (valid) 
        {
            return ReadLength(len);
        }
    }
    break;
    default:
        break;
    }
    return std::string_view{};
}

void WireDecoder::DecodeUnknown() {
    switch (wt) {
    case WT_VARINT:
        ReadVarint();
        break;
    case WT_LEN:
        {
            uint64_t len = ReadVarint();
            if (valid) 
            {
                ReadLength(len);
            }
        }
        break;
    case WT_I64:
        ReadFixed64();
        break;
    case WT_I32:
        ReadFixed32();
    default:
        break;
    }
}