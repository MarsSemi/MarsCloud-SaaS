#ifndef __MARSMQTT__
#define __MARSMQTT__
//--------------------------------------------------------------
//
//--------------------------------------------------------------
#pragma once
//--------------------------------------------------------------
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <curl/curl.h>
//--------------------------------------------------------------
#include "MarsClient.h"
#include "MQTTClient/include/MQTTClient.h"
//--------------------------------------------------------------
//
//--------------------------------------------------------------
typedef void (*fConnectLostMQTTCallback)(char *, char *);
typedef void (*fRecevieMsgMQTTCallback)(char *, int, MQTTClient_message *);
//--------------------------------------------------------------
//
//--------------------------------------------------------------
class MarsMQTT
{
private:
	const static int _DefaultTimeOut_Sec = 10; //sec
	const static int _DefaultAcceptTimeOut_MSec = 5000; //msec
private:
	MarsClient *_MarsClient;
	MQTTClient _MQTTClient;
public:
	fConnectLostMQTTCallback _LostCallback;
	fRecevieMsgMQTTCallback _RecvCallback;
private:
	void ResetClient(void);
	int IndexOf(const char *_str, char _c);
public:
	MarsMQTT(MarsClient *_client);
	~MarsMQTT();

	bool Connect(fConnectLostMQTTCallback _lostHandler, fRecevieMsgMQTTCallback _recvHandler);
	bool Disconnect(void);
	bool Receive(void);
	bool Subscribe(const char *_topic);
};
//--------------------------------------------------------------
#endif