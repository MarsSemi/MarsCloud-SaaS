//--------------------------------------------------------------
//
//--------------------------------------------------------------
#include "MarsClient.h"
//--------------------------------------------------------------
int main(void)
{
	InitNetwork();

	char _resp[512];
	char _data[] = "{ \"temp\": 24.5, \"humi\": 82 }";

	MarsClient _client("test", "test", "justtest");

	if(_client.DoLogin("https://test.mars-cloud.com")) printf("Login SUCCESS\n");
	if(_client.RegDevice("test.com", "test", "metter01", "both.temp")) printf("Reg Device SUCCESS\n");
	if(_client.PutData("test", "metter01", _data)) printf("Put Data SUCCESS\n");
	if(_client.CallService("service.databroker", "/api/hello", NULL, _resp)) printf("Call Service SUCCESS\n");

	return 0;
}
//--------------------------------------------------------------
