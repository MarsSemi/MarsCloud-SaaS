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
		_MarsClient = _client;
		_MQTTClient = NULL;
		_LostCallback = NULL;
		_RecvCallback = NULL;
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
int MarsMQTT::IndexOf(const char *_str, char _c)
{
	try
	{
		for(int i=0;i<strlen(_str);i++)
			if(_str[i] == _c)
				return i;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
	return -1;
}
//--------------------------------------------------------------
void MarsMQTT::ResetClient(void)
{
	try
	{	
		if(_MQTTClient != NULL)
		{
			MQTTClient_destroy(&_MQTTClient);
			_MQTTClient = NULL;
		}

		int _hostStartIndex = 0;
		int _hostEndIndex = 0;
		char _host[128];
		char _mqttHost[128];
		char _id[1024];
		const char *_orgHost = _MarsClient->GetHost();
		MQTTClient_createOptions _createOpts = MQTTClient_createOptions_initializer;
		
		_createOpts.MQTTVersion = MQTTVERSION_DEFAULT;

		if(_hostEndIndex <= 0) _hostEndIndex = IndexOf(_orgHost+10, ':');
		if(_hostEndIndex <= 0) _hostEndIndex = IndexOf(_orgHost+10, '/');
		if(_hostEndIndex <= 0) _hostEndIndex = strlen(_orgHost)-1;

		if(strncmp(_orgHost, "http://", 7)) _hostStartIndex = 8;
		if(strncmp(_orgHost, "https://", 8)) _hostStartIndex = 9;

		if(_hostEndIndex > _hostStartIndex)
		{
			strncpy(_host, _MarsClient->GetHost()+_hostStartIndex, _hostEndIndex - _hostStartIndex + 1);
			sprintf(_mqttHost, "wss://%s:8884", _host);	
			sprintf(_id, "%s@%d", _MarsClient->GetAccount(), rand()%10000);	
			
			MQTTClient_createWithOptions(&_MQTTClient, _mqttHost, _id, MQTTCLIENT_PERSISTENCE_NONE, NULL, &_createOpts);
		}
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
}
//--------------------------------------------------------------
bool MarsMQTT::Connect(fConnectLostMQTTCallback _lostHandler, fRecevieMsgMQTTCallback _recvHandler)
{
	try
	{	
		int _rc = MQTTCLIENT_NULL_PARAMETER;

		_LostCallback = _lostHandler;
		_RecvCallback = _recvHandler;
		
		if(_MQTTClient == NULL) ResetClient();
		if(_MQTTClient != NULL)
			if(MQTTClient_setCallbacks(_MQTTClient, this, connectionLost, messageArrived, NULL) == MQTTCLIENT_SUCCESS)
			{
				MQTTClient_connectOptions _connOpts = MQTTClient_connectOptions_initializer;
				MQTTClient_SSLOptions _sslOpts = MQTTClient_SSLOptions_initializer;
				MQTTClient_willOptions _willOpts = MQTTClient_willOptions_initializer;

				_connOpts.keepAliveInterval = 15;
				_connOpts.username = _MarsClient->GetAccount();
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