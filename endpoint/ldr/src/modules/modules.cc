#include "modules/modules.h"




BOOL Modules::IsModuleLoaded(INT32 bit) {
	return (this->LoadModulesBitMap >> bit) & 1u;
}


BOOL Modules::LoadModule(PBYTE Module, SIZE_T ModuleSize) {
	if (!this->IsStompedDllLoaded) {
		if ( ! this->LoadStompedDll( ) ) {
			return FALSE;
		}
	}

	return TRUE;
}




Modules* modules = NULL;