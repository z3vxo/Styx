#include "../../includes/core/core.h"
#include "../../includes/parser/parser.h"

PBYTE GetConfig() {
	return (PBYTE)"\x24\x00\x00\x00\x61\x35\x63\x37\x32\x39\x65\x61\x2d\x64\x39\x39\x30\x2d\x34\x38\x65\x38\x2d\x61\x34\x39\x37\x2d\x36\x62\x35\x39\x61\x33\x66\x38\x34\x30\x38\x61\x24\x00\x00\x00\x36\x32\x61\x31\x63\x34\x32\x65\x2d\x62\x30\x32\x30\x2d\x34\x63\x38\x39\x2d\x61\x33\x31\x38\x2d\x65\x38\x36\x35\x34\x32\x64\x66\x38\x37\x31\x34";
}

UINT32 ConfigSize() {
	return 80;
}

BOOL ParseConfig() {
	ParserRead *p = ParserInitRead(GetConfig(), ConfigSize());
	INT bytesRead = 0;
	bytesRead = ParserReadStringInto(p, ldr->config->CampaignID, 37);
	if(bytesRead == 0) 
		return FALSE;

	bytesRead = ParserReadStringInto(p, ldr->config->AgentId, 37);
	if(bytesRead == 0) 
		return FALSE;
	
	return TRUE;

}


