# Build
make as

# project structures
protobuf comprises two parts:
1. plugin to generate lang-specific codes
2. lang-specific runtime library


# attentions
不要给字段起奇怪的名字，比如c++关键字`new`，如果有请另起名。

> Don't use wried name such as C++ keyword "new"

```
message LevelUpReq {
	uint32 old = 1;
	uint32 new = 2;		// new is a keyword, rename it.
}
```

# usage

1. 生成cpp代码
2. 链接运行时，主要是 wire.cpp
3. 用法如下

1. generate cpp code
2. link runtime (mainly wire.cpp)
3. write your angelscript code like following

```cpp
LevelUpReq req;
req.old = 3;
req._new = 4;
ustring encoded = req.Serialize();
send(encoded);
```


# considerations

## Why C++
1. UHT自动会导给AS
2. 普适性，有可能需要在C++层业务逻辑上发包
3. 特别快
4. 缺限：schema变量时需要重新编译c++，不过不要紧，很快。

1. UHT will automatically export 
2. compatibility, maybe you need to serialize items in c++ layer.(e.g. some logic implementated in c++ because something is only avaliable in c++ rather than AS)
3. for effeciency, c++ is far faster than angelscript
4. cons: recompilation needed if schema changes, but it doesn't matter and it just take a few seconds.

## HasXxx

Protobuf v3, 天然是支持可选字段的。所以一个字段要区分是否存在要有带外数据，不能靠零值故有方法`bool HasXxx() const`，
你是不能千 x==0 来区分到底这个字段就是`0`，还是说这个字段是nil。 AS没有类似的null值。
如果是repeat字段，就不会有`HasXxx` 因为TArray size == 0时就表示无数据了。

> Protobuf v3, is by default field-optional, 
> so it has a Method called `bool HasXxx() const` to indicate whether the filed exists.
> for instance, you can't tell a integer field "Age" exsistence by comparing it to zero. You should use `HasAge()` instead.
> repeat fields is without `HasXxx` since Tarray size == 0 means that it's empty.(not ambigous)

## FBinary

Unreal的FString是编码敏感的，有时候我们需要一个编码中立的结构，所以我们定义了FBinary，其实现就是 `TArray<uint8>`。
所以消息定义里的string并不对应FString 而是对应 FBinary。

> Since Unreal's FString is encoding-sensitive(with ucs16), whereas FBinary provides an encoding-neutual solution. Sometimes, some binary data is needed. so string in protobuf's message schema isn't corresponded with FString. it uses FBinary instead.


## Network-agnostic
本库只负责序列化反序列化，没有提供网络传输功能，那是应用层的事情，example里有send的示例

> This lib is only in charge of seralization rathan than network transmision.
> There is a demo in example folder to call rpc.

# About Unreal Engine AngelScript
https://angelscript.hazelight.se/
