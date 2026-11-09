#include "agent/core.h"
#include "packager/packager.h"

BOOL Kratos::Init() {

	this->modules.kernel32 = api::GetModule( HASH_STRW ( L "KERNEL32.DLL" ) );
	auto pLocalAlloc = ( decltype ( this->apis.LocalAlloc ) ) api::GetProcAddress( this->modules.kernel32, HASH_STRA ( "LocalAlloc" ) );
	if (!pLocalAlloc)
		return FALSE;


	packager = ( Packager * )pLocalAlloc( LMEM_FIXED | LMEM_ZEROINIT, sizeof( Packager ) );
	DEBUG_LOG("[+] Allocated packager");

	//net = (Net)pLocalAlloc(LMEM_FIXED | LMEM_ZEROINIT, sizeof(Net));

	if ( ! this->LoadApis( ) )
		return FALSE;
	if ( ! this->LoadConfig( ) )
		return FALSE;

	return TRUE;
}

BOOL Kratos::LoadApis( ) {
	DEBUG_LOG("[+] Loading APIS");
	this->modules.kernel32 = api::GetModule (HASH_STRW ( L "KERNEL32.DLL" ) );
	this->modules.ntdll = api::GetModule( HASH_STRW ( L "NTDLL.DLL" ) );


	if ( this->modules.kernel32 ) {
		this->apis.LoadLibraryA   = ( decltype ( this->apis.LoadLibraryA ) ) api::GetProcAddress( this->modules.kernel32, HASH_STRA ( "LoadLibraryA" ) );
		this->apis.GetProcAddress = ( decltype ( this->apis.GetProcAddress ) )api::GetProcAddress( this->modules.kernel32, HASH_STRA ( "GetProcAddress" ) );

		this->apis.LocalAlloc   = ( decltype ( this->apis.LocalAlloc ) ) api::GetProcAddress( this->modules.kernel32, HASH_STRA ( "LocalAlloc" ) );
		this->apis.LocalFree    = ( decltype( this->apis.LocalFree ) ) api::GetProcAddress( this->modules.kernel32, HASH_STRA( "LocalFree" ) );
		this->apis.LocalReAlloc = ( decltype ( this->apis.LocalReAlloc ) ) api::GetProcAddress( this->modules.kernel32, HASH_STRA ( "LocalReAlloc" ) );

		this->apis.GetUserGeoID = ( decltype ( this->apis.GetUserGeoID ) ) api::GetProcAddress( this->modules.kernel32, HASH_STRA ( "GetUserGeoID" ) );
	}
	else { return FALSE; }
	DEBUG_LOG( "[+] Loaded Kernel32" );




	CHAR buf[ 12 ];
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
	this->modules.Advapi32 = this->apis.LoadLibraryA( buf );

	if ( this->modules.Advapi32 ) {
		this->apis.GetComputerNameExW = ( decltype ( this->apis.GetComputerNameExW ) ) api::GetProcAddress( this->modules.Advapi32, HASH_STRA ( "GetComputerNameExW" ) );
		this->apis.GetUserNameW = ( decltype( this->apis.GetUserNameW ) ) api::GetProcAddress( this->modules.Advapi32, HASH_STRA ( "GetUserNameW" ) );
	}
	else { return FALSE; }
	DEBUG_LOG("[+] Loaded advapi32");

	return TRUE;
}

BOOL Kratos::LoadConfig( ) {
	return TRUE;
}