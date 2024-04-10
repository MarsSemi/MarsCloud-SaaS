//--------------------------------------------------------------
//
//--------------------------------------------------------------
#include "MarsClient.h"
//--------------------------------------------------------------
int main(void)
{
	InitNetwork();

	char _resp[512];
	void *_user = CreateUser("test", "test", "justtest");

	if(DoLogin(_user, "https://test.mars-cloud.com"))
		printf("Login SUCCESS");

	CloseUser(_user);
	return 0;
}
//--------------------------------------------------------------
