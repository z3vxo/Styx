#include <windows.h>


#define SEC_B __attribute__((section(".text$B")))

#define HASH_STRA( x ) HasherA( ( x ) )
#define HASH_STRW( x ) HasherW( ( x ) )

typedef struct {
	INT32( WINAPI* ModuleUnpack4 )( );
	PBYTE( WINAPI* ModuleUnpackBytes )( INT32 Length );

	void ( WINAPI* ModulePack4 )( INT32 Value );
	void ( WINAPI* ModulePack8 )( INT64 Value );
	void ( WINAPI* ModulePackBytes )( PBYTE Data, INT32 length );

	FARPROC ( WINAPI* ModuleGetProcAddress )( HMODULE dll, DWORD Hash );
	HMODULE ( WINAPI* ModuleGetModule )( DWORD Hash );

	HLOCAL ( WINAPI* ModuleAllocate )( SIZE_T size );
	HLOCAL ( WINAPI* ModuleFreeMem ) ( HLOCAL buf  );
}Module, *PModule;



#define END_SIG 0xFFFFFFFF

#define OUTPUT_ERROR  0xc4


#define CAT_ERROR_NOT_FILE 20

#define OUTPUT_NEEDS_PARSE 0xc2
#define OUTPUT_NO_PARSE    0xc1
#define PARSER_LS 1

constexpr DWORD HasherA(PCHAR str) {
	UINT32 hash = 5381;
	while (*str) {
		hash = ((hash << 5) + hash) + (unsigned char)*str++;
	}
	return hash;
}

constexpr DWORD HasherW(PWCHAR str) {
	UINT32 hash = 5381;
	while (*str) {
		unsigned char c = (unsigned char)*str++;
		if (c >= 'A' && c <= 'Z') {
			c += 0x20;
		}
		hash = ((hash << 5) + hash) + c;
	}
	return hash;
}


static inline int pStrcmp(const char* s1, const char* s2) {
	while (*s1 && (*s1 == *s2)) {
		s1++;
		s2++;
	}
	return (unsigned char)*s1 - (unsigned char)*s2;
}

static inline SIZE_T pStrlen(const char* str) {
	const volatile char* s = (const volatile char*)str;
	SIZE_T n = 0;
	while (*s++) n++;
	return n;
}