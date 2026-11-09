#include "agent/core.h"
#include <stdio.h>



void Entry( ) {
#ifdef DEBUG
	AllocConsole( );
#endif
	Kratos K = { 0 };

	kratos = &K;

	DEBUG_LOG("[*] Kratos Started");

	if ( ! kratos->Init( ) ) {
		DEBUG_LOG( "[!] Failed With Kratos->Init" );
		return;
	}
	

}

int main( ) {
	Entry( );
	return 0;
}