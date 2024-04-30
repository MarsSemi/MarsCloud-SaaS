//--------------------------------------------------------------
#include "MarsMQTT.h"
//--------------------------------------------------------------
#if !defined(_WIN32)
#include <unistd.h>
#else
#include <windows.h>
#endif

#if defined(_WRS_KERNEL)
#include <OsWrapper.h>
#endif
//--------------------------------------------------------------
//
//--------------------------------------------------------------
int messageArrived(void *_context, char *_topicName, int _topicLen, MQTTClient_message *_m)
{
	try
	{
		fRecevieMsgMQTTCallback _callback = ((MarsMQTT *)_context)->_RecvCallback;
		if(_callback)
			_callback(_topicName, _topicLen, _m);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return 1;
}
//--------------------------------------------------------------
void connectionLost(void *_context, char *_cause)
{
	try
	{
		fConnectLostMQTTCallback _callback = ((MarsMQTT *)_context)->_LostCallback;
		if(_callback)
			_callback(_cause, 0);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
}
//--------------------------------------------------------------
//
//--------------------------------------------------------------
MarsMQTT::MarsMQTT(MarsClient *_client)
{
	try
	{
		char _id[1024];
		MQTTClient_createOptions _createOpts = MQTTClient_createOptions_initializer;
		
		_createOpts.MQTTVersion = MQTTVERSION_DEFAULT;

		_MarsClient = _client;
		_MQTTClient = NULL;
		_LostCallback = NULL;
		_RecvCallback = NULL;

		sprintf(_id, "%s@%d", _MarsClient->GetAccount(), rand()%10000);	

		MQTTClient_createWithOptions(&_MQTTClient, "wss://test.mars-cloud.com:8884", _id, MQTTCLIENT_PERSISTENCE_NONE, NULL, &_createOpts);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
}
//--------------------------------------------------------------
MarsMQTT::~MarsMQTT()
{
	try
	{
		MQTTClient_destroy(&_MQTTClient);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
}
//--------------------------------------------------------------
bool MarsMQTT::Connect(void)
{
	try
	{	
		int _rc = MQTTCLIENT_NULL_PARAMETER;

		if(_MQTTClient != NULL)
		{
			MQTTClient_connectOptions _connOpts = MQTTClient_connectOptions_initializer;
			MQTTClient_SSLOptions _sslOpts = MQTTClient_SSLOptions_initializer;
			MQTTClient_willOptions _willOpts = MQTTClient_willOptions_initializer;

			_connOpts.keepAliveInterval = 15;
			_connOpts.username = "test";
			_connOpts.password = _MarsClient->GetToken();
			_connOpts.MQTTVersion = MQTTVERSION_DEFAULT;
			_connOpts.cleansession = 1;
			_connOpts.ssl = &_sslOpts;

			_sslOpts.enableServerCertAuth = 0;
			_sslOpts.verify = 1;

			if((_rc = MQTTClient_connect(_MQTTClient, &_connOpts)) == MQTTCLIENT_SUCCESS)
				return true;
		}

		printf("Failed to create mqtt, return code %d\n", _rc);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
	return false;
}
//--------------------------------------------------------------
bool MarsMQTT::Receive(void)
{
	try
	{
		if(MQTTClient_disconnect(_MQTTClient, _DefaultAcceptTimeOut_MSec) == MQTTCLIENT_SUCCESS)
			return true;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
	return false;
}
//--------------------------------------------------------------
bool MarsMQTT::Subscribe(const char *_topic)
{
	try
	{
		if(_MQTTClient != NULL)
			if(MQTTClient_subscribe(_MQTTClient, _topic, 0) == MQTTCLIENT_SUCCESS)
				return true;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
	return false;
}
//--------------------------------------------------------------
bool MarsMQTT::SetCallback(fConnectLostMQTTCallback _lostHandler, fRecevieMsgMQTTCallback _recvHandler)
{
	try
	{
		_LostCallback = _lostHandler;
		_RecvCallback = _recvHandler;

		if(_MQTTClient != NULL)
			if(MQTTClient_setCallbacks(_MQTTClient, this, connectionLost, messageArrived, NULL) == MQTTCLIENT_SUCCESS)
				return true;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
	return false;
}
//--------------------------------------------------------------