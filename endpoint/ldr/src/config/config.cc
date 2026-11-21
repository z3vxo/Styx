#include "config/config.h"



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

	return TRUE;
}