#pragma once
#include <windows.h>



#define BASE_PACKER_SIZE 256

class Packager {
public:
	INT32 WriteIndex;
	SIZE_T WriteSize;
	PBYTE WriteBuf;
	INT32 UnPack4();
	INT32 UnPackBytesReadInto( PBYTE out, INT32 N );
	PBYTE UnPackBytes( INT32 n );
	BOOL  NewPacker();



	BOOL Pack4( INT32 val );
	BOOL PackBytes( PBYTE in, INT32 n );
	BOOL EnsurePacker( SIZE_T needed );
	PBYTE PackerPointer( );
	INT32 PackerSize( );
};

extern Packager* packager;