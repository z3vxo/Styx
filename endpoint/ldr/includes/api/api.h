#pragma once
#include <windows.h>




namespace api {
	constexpr DWORD HasherA( PCHAR str ) {
		UINT32 hash = 5381;
		while ( *str ) {
			hash = ( ( hash << 5 ) + hash ) + ( unsigned char )*str++;
		}
		return hash;
	}

	constexpr DWORD HasherW( PWCHAR str ) {
		UINT32 hash = 5381;
		while ( *str ) {
			unsigned char c = ( unsigned char )*str++;
			if ( c >= 'A' && c <= 'Z' ) {
				c += 0x20;
			}
			hash = ( ( hash << 5 ) + hash ) + c;
		}
		return hash;
	}
	FARPROC GetProcAddressInternal( HMODULE dll, DWORD hash, INT depth );
	FARPROC GetProcAddress( HMODULE dll, DWORD hash );
	HMODULE GetModule( DWORD HASH);
}