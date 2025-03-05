#pragma once

#include <Containers/StringFwd.h>
#include <Containers/Array.h>
#include <ProtobufUEAS.generated.h>

//the first letter 'u' is 'unreal' 
using ustring = TArray<uint8>;
using ubinary = ustring;
using usview = TStringView<uint8>;

USTRUCT(BlueprintType)
struct FBinary
{
	GENERATED_BODY()
	ustring _BinaryData;

	FBinary() = default;
	FBinary(const TArray<uint8> &o): _BinaryData(o) {}
	FBinary(TArray<uint8> &&o):_BinaryData(MoveTemp(o)) {}
};


UCLASS(Meta = (ScriptMixin = "FString"))
class UFStringScriptMixinLibrary : public UObject
{
	GENERATED_BODY()
public:
	UFUNCTION(ScriptCallable)
	static FBinary ToBinary(const FString &SourceString)
	{
		const FTCHARToUTF8 Convert(*SourceString);
		
		// TODO:这里可以无拷贝构造
		return FBinary(
			ustring(reinterpret_cast<const uint8*>(Convert.Get()),
			Convert.Length()));
	}
	
	UFUNCTION(ScriptCallable)
	static FBinary AsBinary(const FString &SourceString)
	{
		return FBinary(ustring(reinterpret_cast<const uint8*>(GetData(SourceString)),
			sizeof(TCHAR)*GetNum(SourceString)));
	}
};


UCLASS(Meta = (ScriptMixin = "FBinary"))
class UFBinaryScriptMixinLibrary : public UObject
{
	GENERATED_BODY()
public:
	UFUNCTION(ScriptCallable)
	static FString ToString(const FBinary& InBinary, bool& bOk)
	{
		const TArray<uint8> &SourceData = InBinary._BinaryData;
		
		const FUTF8ToTCHAR Convert(reinterpret_cast<const ANSICHAR*>(SourceData.GetData())
			, SourceData.Num());

		bOk = true;	// 暂时假定必成功
		// TODO:这里也可以无拷贝构造的
		return  FString {Convert.Length(), Convert.Get()};
	}

	// 重载版本，忽略
	UFUNCTION(ScriptCallable)
	static FString ToStringUnchecked(const FBinary& InBinary)
	{
		const TArray<uint8> &SourceData = InBinary._BinaryData;
		const FUTF8ToTCHAR Convert(reinterpret_cast<const ANSICHAR*>(SourceData.GetData())
			, SourceData.Num());

		// TODO:这里也可以无拷贝构造的
		return  FString {Convert.Length(), Convert.Get()};
	}
	
	UFUNCTION(ScriptCallable)
	static FString AsString(const FBinary& InBinary)
	{
		const TArray<uint8> &SourceData = InBinary._BinaryData;
		// assert(InBinary._BinaryData.Num() %2 == 0)
		return FString{SourceData.Num()/2,
			reinterpret_cast<const TCHAR*>(SourceData.GetData())};
	}
	
	UFUNCTION(ScriptCallable)
	static int32 Size(const FBinary& InBinary)
	{
		//TIsSame is deprecated.
		static_assert(std::is_same_v<decltype(InBinary._BinaryData.Num()), int32>);
		return InBinary._BinaryData.Num();
	}

	static void DumpHex(const uint8* Data, int32 Size)
	{
		FString Output;
#define TO_HS(idx) (Data[i+idx] >= ' ' && Data[i+idx] <= '~' ? (TCHAR)Data[i+idx] : L'.')
		for (int32 i = 0; i < Size; i += 16) {
			if (i + 16 <= Size)
			{
				Output += FString::Printf(TEXT("%04x:  %02x %02x %02x %02x %02x %02x %02x %02x  %02x %02x %02x %02x %02x %02x %02x %02x"),
					i,
					Data[i+0], Data[i+1], Data[i+2], Data[i+3],
					Data[i+4], Data[i+5], Data[i+6], Data[i+7],
					Data[i+8], Data[i+9], Data[i+10], Data[i+11],
					Data[i+12], Data[i+13], Data[i+14], Data[i+15]);
				
				Output += FString::Printf(TEXT("\t|  %c%c%c%c%c%c%c%c%c%c%c%c%c%c%c%c\n"),
					TO_HS(0), TO_HS(1), TO_HS(2), TO_HS(3), 
					TO_HS(4), TO_HS(5), TO_HS(6), TO_HS(7),
					TO_HS(8), TO_HS(9), TO_HS(10), TO_HS(11),
					TO_HS(12), TO_HS(13), TO_HS(14), TO_HS(15));
			}
			else
			{
				Output += FString::Printf(TEXT("%04x:  "), i);
				
				for (int32 j = 0; j < 16; ++j)
				{
					if (j+1 % 8 == 0)
						Output += TEXT(" ");
					Output += (i+j >= Size) ? FString(TEXT("   ")) : FString::Printf(TEXT("%02x "), Data[i+j]);
				}
				Output += TEXT("\t|  ");
				
				for (int32 j = 0; j < 16; ++j)
				{
					Output += (i+j >= Size) ? TEXT(".") : FString::Printf(TEXT("%c"), TO_HS(j));
				}
				Output += TEXT("\n");
			}
		}
#undef TO_HS
		UE_LOG(LogTemp, Display, TEXT("%s"), *Output);
	}
	
	UFUNCTION(ScriptCallable)
	static void DumpHex(const FBinary &InBinary)
	{
		const uint8 *Data = InBinary._BinaryData.GetData();
		const int32 Size = InBinary._BinaryData.Num();
		DumpHex(Data, Size);
	}
};



USTRUCT(BlueprintType)
struct FPbMessage
{
	GENERATED_BODY()
	
	virtual ~FPbMessage() {}

	virtual FString GetMessageName() const { return FString{}; };
	
	virtual ustring Encode() const { return ustring{}; };

	// true if succeed to decode
	virtual bool Decode(usview) {return false;};
	virtual void DumpJson(TSharedPtr<FJsonObject> obj) const { return; };
};
