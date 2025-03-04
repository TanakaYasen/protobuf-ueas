# Build
make as


# project structures
protobuf comprises two parts:
1. plugin to generate lang-specific codes
2. lang-specific runtime library


# attentions
Don't use wried name such as C++ keyword "new"

```
message LevelUpReq {
	uint32 old = 1;
	uint32 new = 2;		// new is a keyword, rename it.
}
```

# usage

1. generate cpp code
2. link runtime (mainly wire.cpp)
3. write your angelscript code like following
```
LevelUpReq req;
req.old = 3;
req._new = 4;
ustring encoded = req.Serialize();
send(encoded);
```


# considerations

## HasXxx

Protobuf v3, 天然是支持可选字段的。所以一个字段要区分是否存在要有带外数据，不能靠零值故有方法`bool HasXxx() const`

Protobuf v3, is by default field-optional, 
so it has a Method called `bool HasXxx() const` to indicate whether the filed exists.
for instance, you can't tell a integer field "Age" exsistence by comparing it to zero. You should use `HasAge()` instead.

## Network-agnostic
本库只负责序列化反序列化，没有提供网络传输功能，那是应用层的事情，example里有send的示例

This lib is only in charge of seralization rathan than network transmision.

There is a demo in example folder to call rpc.

# About Unreal Engine AngelScript
https://angelscript.hazelight.se/
