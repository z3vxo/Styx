#include "agent/core.h"


// basic for now
void Kratos::Sleeper() {
	DEBUG_LOG("[*] Sleeping");

	Sleep(this->settings.Sleep * 1000);
}