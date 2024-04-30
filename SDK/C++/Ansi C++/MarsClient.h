#ifndef __MARSCLIENT__
#define __MARSCLIENT__
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
//
//--------------------------------------------------------------
struct HttpPostPayload
{
	const char *_buf;
	int _size;
	int _sentSize;
};
//--------------------------------------------------------------
//
//--------------------------------------------------------------
bool InitNetwork(void);
//--------------------------------------------------------------
size_t HttpWriteBack(void *_content, size_t _size, size_t _nmemb, void *_dataPtr);
//--------------------------------------------------------------
size_t HttpReadBack(void *_dest, size_t _size, size_t _nmemb, void *_dataPtr);
//--------------------------------------------------------------
//
//--------------------------------------------------------------
class MarsClient
{
private:
	const static int _DefaultTimeOut_Sec = 30; //sec
	const static int _DefaultAcceptTimeOut_MSec = 5000; //msec
private:
	char _Account[128];
	char _Password[24];
	char _Proj[64];
	char _Token[513];
	char _Host[128];
private:
	bool HttpGET(char *_respone, const char *_url, const char *_token);
	bool HttpPost(char *_respone, const char *_url, const char *_payload, int _size, const char *_token);

	bool HttpGetData(char *_result, const char *_request);
	bool HttpPostData(char *_result, const char *_request, const char *_payload);
public:
	char *GetAccount(void);
	char *GetToken(void);	
public:
	MarsClient(const char *_account, const char *_password, const char *_proj);
	~MarsClient();

	bool DoLogin(const char *_host);
	bool RegDevice(const char *_uuid, const char *_suid, const char *_type, const char *_name, const char *_vendor);

	bool PutMessage(const char *_topic, const char *_postData);
	bool PutData(const char *_uuid, const char *_suid, const char *_content);
	bool GetLastData(const char *_uuid, const char *_suid, int _count, char *_result);
	bool RemoveData(const char *_uuid, const char *_suid, const char *_ukey, char *_result);

	bool CallService(const char *_service, const char *_api, const char *_postData, char *_result);
};
//--------------------------------------------------------------
#endif