#pragma once
#include <windows.h>
#include <winhttp.h>

#define DECL( x ) decltype( x ) * x


class NetMgr {
private:
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
	BOOL init( );
	BOOL Post( );
};


extern NetMgr* net;