#include "agent/core.h"
#include "utils/utils.h"
#include "packager/packager.h"



BOOL Packager::NewPacker( ) {
	if ( this->WriteBuf )
		kratos->apis.LocalFree( this->WriteBuf );

	this->WriteBuf = ( PBYTE )kratos->apis.LocalAlloc( LMEM_FIXED | LMEM_ZEROINIT, 256 );
	if ( ! this->WriteBuf )
		return FALSE;

	this->WriteIndex = 0;
	this->WriteSize = 256;
	return TRUE;
}

INT32 Packager::UnPack4( ) {
	return 0;
}

INT32 Packager::UnPackBytesReadInto( PBYTE out, INT32 n ) {
	return 0;
}

PBYTE Packager::UnPackBytes( INT32 n ) {
	return NULL;
}




BOOL Packager::EnsurePacker( SIZE_T needed ) {
	if ( ! this->WriteBuf ) { return FALSE; }
	if ( needed <= this->WriteSize ) { return TRUE; }

	SIZE_T NewSize = this->WriteSize ? this->WriteSize : BASE_PACKER_SIZE;
	do {
		if  (NewSize > ( ( SIZE_T ) -1 ) / 2) { return FALSE; }
		NewSize *= 2;
	} while ( NewSize < needed );

	PBYTE NewBuf = ( PBYTE )kratos->apis.LocalReAlloc( this->WriteBuf, NewSize, LMEM_FIXED );
	if (!NewBuf) { return FALSE; }

	this->WriteBuf = NewBuf;
	this->WriteSize = NewSize;
	return TRUE;
}

BOOL Packager::Pack4( INT32 val ) {
	if ( ! this->EnsurePacker( ( SIZE_T ) this->WriteIndex + sizeof ( val ) ) ) { return FALSE; }
	utils::mem::memcpy( this->WriteBuf + this->WriteIndex, &val, sizeof( val ) );

	this->WriteIndex += sizeof( val );
	return TRUE;
}

BOOL Packager::PackBytes( PBYTE in, INT32 n ) {
	if (!this->EnsurePacker( ( SIZE_T ) this->WriteIndex + ( SIZE_T )n ) ) { return FALSE; }
	utils::mem::memcpy( this->WriteBuf + this->WriteIndex, in, n );
	this->WriteIndex += n;
	return TRUE;
}

BOOL Packager::PackWideBytes( PWSTR str ) {
	if ( ! str ) { return FALSE; }
	INT needed = WideCharToMultiByte( CP_UTF8, WC_ERR_INVALID_CHARS, str, -1, NULL, 0, NULL, NULL );
	if ( needed == 0 ) { return FALSE; }
	PBYTE buf = (PBYTE)kratos->apis.LocalAlloc( LMEM_FIXED | LMEM_ZEROINIT, needed );
	if ( !buf ) { return FALSE; }
	INT Written = WideCharToMultiByte( CP_UTF8, WC_ERR_INVALID_CHARS, str, -1, ( PSTR )buf, needed, NULL, NULL );
	BOOL res = FALSE;
	if (Written != 0) {
		this->Pack4(Written);
		res = this->PackBytes( buf, Written );
	}
	kratos->apis.LocalFree( buf );
	return res;
}



PBYTE Packager::PackerPointer() {
	return this->WriteBuf;
}

INT32 Packager::PackerSize() {
	return this->WriteIndex;
}

BOOL Packager::PrepareMetaData(INT32 AgentID) {
	if ( ! this->Pack4( METADATA_MAGIC ) ) { return FALSE; }
	if ( ! this->Pack4( AgentID ) ) { return FALSE; }

	return TRUE;
}




Packager* packager = NULL;