//--------------------------------------------------------------
//
//--------------------------------------------------------------
#include <unistd.h>
//--------------------------------------------------------------
#include "MarsClient.h"
#include "MarsMQTT.h"
//--------------------------------------------------------------
void LostMQTTCallback(char *_topic, char *_cause)
{
	printf("---\n");
	printf("MQTT Disconnect : %s\n", _cause);
}
//--------------------------------------------------------------
void ReceiveMQTTCallback(char *_topic, int _len, MQTTClient_message *_msg)
{
	char _payload[4096];

	strcpy(_payload, (char *)_msg->payload);

	printf("---\n");
	printf("MQTT : %s\n", _topic);
	printf("MQTT : %s\n", _payload);
}
//--------------------------------------------------------------
int main(void)
{
	InitNetwork();

	char _resp[512];
	char _data[] = "{ \"temp\": 24.5, \"humi\": 82 }";

	MarsClient _client("test", "test", "justtest");
	MarsMQTT _mqtt(&_client);

	if(_client.DoLogin("https://test.mars-cloud.com"))
	{
		printf("Login SUCCESS\n");
		printf("---\n");

		if(_mqtt.Connect(&LostMQTTCallback, &ReceiveMQTTCallback)) printf("MQTT Connect SUCCESS\n");		
		if(_mqtt.Subscribe("test/+/#")) printf("MQTT Subscribe SUCCESS\n");	

		printf("---\n");
		if(_client.RegDevice("test", "metter01", "both.temp", "virtual_device", "test.com")) printf("Reg Device SUCCESS\n");
		if(_client.PutData("test", "metter01", _data)) printf("Put Data SUCCESS\n");
		if(_client.CallService("service.databroker", "/api/hello", NULL, _resp)) printf("Call Service SUCCESS\n");
		if(_client.PutMessage("test/my/msg", _data)) printf("Put Message SUCCESS\n");

		while (true)
			sleep(100);
	}

	return 0;
}
//--------------------------------------------------------------
