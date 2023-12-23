
#pragma once

#include "Modules/ModuleManager.h"

class FProtobufUEAS : public IModuleInterface
{
public:
	/** IModuleInterface implementation */
	virtual void StartupModule() override;
	virtual void ShutdownModule() override;

private:
};