#pragma once
#include <windows.h>



#define BASE_PACKER_SIZE 256
#define METADATA_MAGIC   0xf3f2f1ff


class Packager {
public:
	static void* operator new(SIZE_T size) {
		void* ptr = kratos->apis.LocalAlloc(ALLOC_ARGS, size);
		if ( ! ptr ) { return NULL; }
		return ptr;
	}
	

	PBYTE ReadBuf;
	SIZE_T ReadBufSize;
	INT32 ReadIndex;
	BOOL  NewUnPacker(PBYTE Data, SIZE_T size);
	INT32 UnPack4( );
	INT32 UnPackBytesReadInto( PBYTE out, INT32 N );
	PBYTE UnPackBytes( INT32 n );
	BOOL  NewPacker( );

	INT32 WriteIndex;
	SIZE_T WriteSize;
	PBYTE WriteBuf;
    BOOL PrepareMetaData( INT32 AgentID );
	BOOL Pack4( INT32 val );
	BOOL Pack8( INT64 val );
	BOOL PackBytes( PBYTE in, INT32 n );
	BOOL PackWideBytes( PWSTR str );
	BOOL EnsurePacker( SIZE_T needed );
	PBYTE PackerPointer( );
	INT32 PackerSize( );
};

extern Packager* packager;