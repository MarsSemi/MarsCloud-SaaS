//--------------------------------------------------------------
//
//--------------------------------------------------------------
#include <stdio.h>
#include <stdlib.h>
#include <conio.h>
#include <string.h>
#include <net.h>
//--------------------------------------------------------------
struct _User
{
	char _Account[256];
	char _Password[16];
	char _Proj[256];
	char _Token[512];

	SOCKET _SubscribeSocket = NULL;
};
//--------------------------------------------------------------
void* CreateUser(char *_account, char *_password, char *_proj)
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
	catch (...) { printf("\nfunc : CreateUser exception"); }

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
			CloseSocket(_user->_SubscribeSocket);
			memset(_user, 0, sizeof(_User));
			free(_user);
		}
	}
	catch (...) { printf("\nfunc : CloseUser exception"); }
}
//--------------------------------------------------------------
bool ParseHttpResult(char *_result, char *_respone)
{
	try
	{
		if (_result == NULL)
			return false;
		
		if (_respone != NULL && strstr(_respone, "200 OK") != NULL)
		{
			const char *_seperator = "\r\n\r\n";
			const int _seperator_len = strlen(_seperator);
			char *_pch = strstr(_respone, _seperator);

			if (_pch != NULL)
			{
				int _pch_len = strlen(_pch);
				memcpy(_result, _pch + _seperator_len, strlen(_pch) - _seperator_len);
				return true;
			}
		}
	}
	catch (...) { printf("\nfunc : GetHttpResult exception"); }
	return false;
}
//--------------------------------------------------------------
bool HttpGetData(char *_result, void *_handle, char *_host, int _port, char *_request)
{
	try
	{
		if (_result == NULL || _request == NULL)
			return false;

		_User *_user = (_User *)_handle;
		if (_user == NULL || _host == NULL || strlen(_user->_Token) <= 0)
			return false;

		bool _status = false;
		char *_respone = (char *)malloc(512 * 1024);

		if (HttpGET(_respone, _host, _port, _request, _user->_Token) > 0)
			if (ParseHttpResult(_result, _respone))
			{
				printf("\n\nGetUserDataSrcList : %s", _result);
				_status = true;
			}

		free(_respone);
		return _status;
	}
	catch (...) { printf("\nfunc : GetHttpResult exception"); }
	return false;
}
//--------------------------------------------------------------
bool DoLogin(void *_handle, char *_host, int _port)
{
	try
	{
		_User *_user = (_User *)_handle;
		if (_user == NULL || _host == NULL)
			return false;
		
		char _req[1024];
		sprintf(_req, "/auth/login?usr=%s&pwd=%s&proj=%s", _user->_Account, _user->_Password, _user->_Proj);

		bool _status = false;
		char *_respone = (char *)malloc(512 * 1024);

		if(HttpGET(_respone, _host, _port, _req, NULL) > 0)
			if (ParseHttpResult(_user->_Token, _respone))
			{
				printf("\n\nLogin Token : %s", _user->_Token);
				_status = true;
			}

		free(_respone);
		return _status;
	}
	catch (...) { printf("\nfunc : DoLogin exception"); }
	return false;
}
//--------------------------------------------------------------
bool GetUserDataSrcList(char *_result, void *_handle, char *_host, int _port)
{
	try
	{
		return HttpGetData(_result, _handle, _host, _port, "/api/usrinfo?method=datasrclist");
	}
	catch (...) { printf("\nfunc : GetUserDataSrcList exception"); }
	return false;
}
//--------------------------------------------------------------
