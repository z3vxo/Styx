#pragma once
#include <windows.h>



#define BASE_PACKER_SIZE 256
#define METADATA_MAGIC   0xfff1f2f3

class Packager {
public:
	INT32 WriteIndex;
	SIZE_T WriteSize;
	PBYTE WriteBuf;
	INT32 UnPack4();
	INT32 UnPackBytesReadInto(PBYTE out, INT32 N);
	PBYTE UnPackBytes(INT32 n);
	BOOL  NewPacker();


    BOOL PrepareMetaData(INT32 AgentID);
	BOOL Pack4(INT32 val);
	BOOL PackBytes(PBYTE in, INT32 n);
	BOOL PackWideBytes( PWSTR str );
	BOOL EnsurePacker(SIZE_T needed);
	PBYTE PackerPointer();
	INT32 PackerSize();
};

extern Packager* packager;