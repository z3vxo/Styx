#pragma once
#include <windows.h>
#include <stdio.h>
#include <wchar.h>

#include "nt.h"
#include "../api/api.h"


#ifdef DEBUG
#define DEBUG_LOG( msg , ... )  printf( msg "\n", ##__VA_ARGS__)
#define DEBUG_LOGW( msg, ... )  wprintf( msg L"\n", ##__VA_ARGS__)
#else
#define DEBUG_LOG(msg, ...)  
#define DEBUG_LOGW(msg, ...) 
#endif



#define DECL( x ) decltype( x ) * x

#define HASH_STRA( x ) api::HasherA( ( x ) )
#define HASH_STRW( x ) api::HasherW( ( x ) )

class Kratos {
public:
	struct Apis {
		DECL( GetUserNameW );
		DECL( GetComputerNameExW );

		DECL( LoadLibraryA );
		DECL( GetProcAddress );
		DECL( LocalReAlloc );
		DECL( LocalAlloc );
		DECL( LocalFree );
		DECL( GetUserGeoID );
	};

	struct Modules {
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
	Modules modules;
	void Register( );
	BOOL LoadConfig( );
	BOOL Init( );
	BOOL LoadApis( );
	BOOL Run( );
};

extern Kratos* kratos;