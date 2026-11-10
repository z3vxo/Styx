#include "net/net.h"
#include "agent/core.h"

NetMgr* net = NULL;



BOOL NetMgr::init() {
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

		return FALSE;
	}


	return TRUE;
}
