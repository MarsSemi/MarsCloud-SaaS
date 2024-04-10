#ifndef __MARSCLIENT__
#define __MARSCLIENT__
//--------------------------------------------------------------
//
//--------------------------------------------------------------
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <curl/curl.h>
//--------------------------------------------------------------
//
//--------------------------------------------------------------
struct _User
{
	char _Account[256];
	char _Password[16];
	char _Proj[256];
	char _Token[512];
	char _Host[128];
};
//--------------------------------------------------------------
//
//--------------------------------------------------------------
bool InitNetwork(void)
{
	try
	{
		curl_global_init(CURL_GLOBAL_ALL);
		return true;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
size_t HttpWriteBack(void *_content, size_t _size, size_t _block, void *_dataPtr)
{
	try
	{
		size_t _realSize = _size*_block;
		memcpy(_dataPtr, _content, _realSize);
		return _realSize;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return 0;
}
//--------------------------------------------------------------
bool HttpGET(char *_respone, const char *_request, const char *_token)
{
	try
	{
		CURL *_curl = curl_easy_init();

		curl_easy_setopt(_curl, CURLOPT_URL, _request);
		curl_easy_setopt(_curl, CURLOPT_WRITEDATA, _respone);
		curl_easy_setopt(_curl, CURLOPT_WRITEFUNCTION, HttpWriteBack);

		if(curl_easy_perform(_curl) == CURLE_OK)
		{
			curl_easy_cleanup(_curl);
			return true;
		}

		curl_easy_cleanup(_curl);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool HttpGetData(char *_result, void *_handle, const char *_request)
{
	try
	{
		if (_result == NULL || _request == NULL) return false;
		if (_handle == NULL || _request == NULL) return false;
		
		bool _status = false;
		char *_respone = (char *)malloc(512 * 1024);
		_User *_user = (_User *)_handle;

		char _req[1024] = { '\0' };
		sprintf(_req, "%s%s", _user->_Host, _request);

		if (HttpGET(_respone, _req, _user->_Token))
			_status = true;
				
		free(_respone);
		return _status;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
//
//--------------------------------------------------------------
void* CreateUser(const char *_account, const char *_password, const char *_proj)
{
	try
	{
		InitNetwork();

		if (_account == NULL || _password == NULL)
			return NULL;

		_User *_user = (_User *)malloc(sizeof(_User));

		memset(_user, 0, sizeof(_User));

		strcpy(_user->_Account, _account);
		strcpy(_user->_Password, _password);
		strcpy(_user->_Proj, _proj);

		return _user;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return NULL;	
}
//--------------------------------------------------------------
void CloseUser(void *_handle)
{
	try
	{
		_User *_user = (_User *)_handle;
		if (_user != NULL)
		{
			memset(_user, 0, sizeof(_User));
			free(_user);
		}
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
}
//--------------------------------------------------------------
bool DoLogin(void *_handle, const char *_host)
{
	try
	{
		if (_handle == NULL || _host == NULL)
		return false;
	
		_User *_user = (_User *)_handle;
		char _req[1024];
		sprintf(_req, "%s/auth/login?usr=%s&pwd=%s&proj=%s", _host, _user->_Account, _user->_Password, _user->_Proj);

		bool _status = false;
		char *_respone = (char *)malloc(512 * 1024);

		strcpy(_user->_Host, _host);

		if(HttpGET(_respone, _req, NULL))
		{
			memcpy(_user->_Token, _respone, sizeof(_user->_Token));
			_status = true;
		}
		
		free(_respone);
		return _status;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
	return false;
}
//--------------------------------------------------------------
bool GetUserDataSrcList(void *_handle, char *_result)
{
	try
	{
		return HttpGetData(_result, _handle, "/api/usrinfo?method=datasrclist");
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
#endif
