//--------------------------------------------------------------
//
//--------------------------------------------------------------
#include "MarsClient.h"
//--------------------------------------------------------------
int main(void)
{
	InitNetwork();

	char _resp[512];
	MarsClient _client("test", "test", "justtest");

	if(_client.DoLogin("https://test.mars-cloud.com")) printf("Login SUCCESS\n");
	if(_client.RegistryDevice("test.com", "test", "temperature", "both.temp")) printf("Reg Device SUCCESS\n");
	if(_client.PutData("test", "metter01", _data)) printf("Put Data SUCCESS\n");

	return 0;
}
//--------------------------------------------------------------
