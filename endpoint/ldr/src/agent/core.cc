#include "agent/core.h"
#include "packager/packager.h"

Kratos* kratos = NULL;


/*
	kratos->Init() Allocates core NetMgr and other critical stuff
	->
	Init calls, this->LoadApis(), this loads all required apis
	Init calls, this->LoadConfig(), this pulls config, parses and saves values
	Init calls, NetMgr->Init(), this loads winhttp, gets apis, saves internal http info(domain, paths, port etc)
	returns
*/



BOOL Kratos::Run() {




	return TRUE;
}



