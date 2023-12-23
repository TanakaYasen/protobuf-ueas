// Copyright Epic Games, Inc. All Rights Reserved.

using UnrealBuildTool;

public class Protobuf : ModuleRules
{
	public Protobuf(ReadOnlyTargetRules Target) : base(Target)
	{
		PrivateDependencyModuleNames.AddRange(
			new string[] {
				"Core",
				"CoreUObject"
			});
	}
}