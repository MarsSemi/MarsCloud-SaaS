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
struct HttpPostPayload
{
	const char *_buf;
	int _size;
	int _sentSize;
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
size_t HttpWriteBack(void *_content, size_t _size, size_t _nmemb, void *_dataPtr)
{
	try
	{
		size_t _realSize = _size*_nmemb;
		memcpy(_dataPtr, _content, _realSize);
		return _realSize;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return 0;
}
//--------------------------------------------------------------
size_t HttpReadBack(void *_dest, size_t _size, size_t _nmemb, void *_dataPtr)
{
	try
	{
		HttpPostPayload *_payload = (HttpPostPayload *)_dataPtr;

		if(_payload->_size > 0)
		{
			size_t _destSize = _size*_nmemb;
			size_t _size = _destSize < _payload->_size ? _destSize : _payload->_size;			

			memcpy(_dest, _payload->_buf+_payload->_sentSize, _size);
			
			_payload->_sentSize += _size;
			_payload->_size -= _size;

			return _size;
		}
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return 0;
}
//--------------------------------------------------------------
//
//--------------------------------------------------------------
class MarsClient
{
private:
	const static int _DefaultTimeOut = 30; //sec
	const static int _DefaultAcceptTimeOut = 5000; //msec
private:
	char _Account[256];
	char _Password[16];
	char _Proj[256];
	char _Token[513];
	char _Host[128];
private:
	bool HttpGET(char *_respone, const char *_url, const char *_token);
	bool HttpPost(char *_respone, const char *_url, const char *_payload, int _size, const char *_token);

	bool HttpGetData(char *_result, const char *_request);
	bool HttpPostData(char *_result, const char *_request, const char *_payload);
public:
	MarsClient(const char *_account, const char *_password, const char *_proj);
	~MarsClient();

	bool DoLogin(const char *_host);
	bool RegistryDevice(const char *_vendor, const char *_uuid, const char *_suid, const char *_type);

	bool GetDataSrcList(char *_result);
	bool GetLastData(const char *_uuid, const char *_suid, int _count, char *_result);
};
//--------------------------------------------------------------
//
//--------------------------------------------------------------
MarsClient::MarsClient(const char *_account, const char *_password, const char *_proj)
{
	try
	{
		InitNetwork();

		if (_account == NULL || _password == NULL)
			return;

		strcpy(_Account, _account);
		strcpy(_Password, _password);
		strcpy(_Proj, _proj);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
}
//--------------------------------------------------------------
MarsClient::~MarsClient()
{
	try
	{
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
}
//--------------------------------------------------------------
bool MarsClient::HttpGET(char *_respone, const char *_url, const char *_token)
{
	try
	{
		struct curl_slist* _headers = NULL;
		CURL *_curl = curl_easy_init();
		
		//curl_easy_setopt(_curl, CURLOPT_VERBOSE, 1);
		curl_easy_setopt(_curl, CURLOPT_FAILONERROR, 1);
		curl_easy_setopt(_curl, CURLOPT_URL, _url);
		curl_easy_setopt(_curl, CURLOPT_TIMEOUT, _DefaultTimeOut);
		curl_easy_setopt(_curl, CURLOPT_ACCEPTTIMEOUT_MS, _DefaultAcceptTimeOut);
		curl_easy_setopt(_curl, CURLOPT_WRITEFUNCTION, HttpWriteBack);
		curl_easy_setopt(_curl, CURLOPT_WRITEDATA, _respone);
		
		if(_token != NULL) { char _buf[540]; sprintf(_buf, "Authentication: Bearer %s", _token); _headers = curl_slist_append(_headers, _buf); };
		if(_headers != NULL) curl_easy_setopt(_curl, CURLOPT_HTTPHEADER, _headers);
		if(curl_easy_perform(_curl) == CURLE_OK)
		{
			curl_slist_free_all(_headers);
			curl_easy_cleanup(_curl);
			return true;
		}
		
		curl_slist_free_all(_headers);
		curl_easy_cleanup(_curl);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool MarsClient::HttpPost(char *_respone, const char *_url, const char *_payload, int _size, const char *_token)
{
	try
	{
		struct curl_slist* _headers = NULL;
		struct HttpPostPayload _data;
		CURL *_curl = curl_easy_init();		

		_data._buf = _payload;
		_data._size = _payload == NULL ? 0 : _size;
		_data._sentSize = 0;

		//curl_easy_setopt(_curl, CURLOPT_VERBOSE, 1);
		curl_easy_setopt(_curl, CURLOPT_POST, 1);	
		curl_easy_setopt(_curl, CURLOPT_FAILONERROR, 1);
		curl_easy_setopt(_curl, CURLOPT_URL, _url);
		curl_easy_setopt(_curl, CURLOPT_TIMEOUT, _DefaultTimeOut);	
		curl_easy_setopt(_curl, CURLOPT_ACCEPTTIMEOUT_MS, _DefaultAcceptTimeOut);
		curl_easy_setopt(_curl, CURLOPT_WRITEFUNCTION, HttpWriteBack);
		curl_easy_setopt(_curl, CURLOPT_WRITEDATA, _respone);		
		curl_easy_setopt(_curl, CURLOPT_READFUNCTION, HttpReadBack);
		curl_easy_setopt(_curl, CURLOPT_READDATA, &_data);
		
		_headers = curl_slist_append(_headers, "Transfer-Encoding:");
		_headers = curl_slist_append(_headers, "Expect:");
		_headers = curl_slist_append(_headers, "Content-Type: application/json; charset=utf-8");

		if(_data._size > 0) { char _buf[48]; sprintf(_buf, "Content-Length: %d", _data._size); _headers = curl_slist_append(_headers, _buf); };
		if(_token != NULL) { char _buf[540]; sprintf(_buf, "Authentication: Bearer %s", _token); _headers = curl_slist_append(_headers, _buf); };				
		if(_headers != NULL) curl_easy_setopt(_curl, CURLOPT_HTTPHEADER, _headers);
		if(curl_easy_perform(_curl) == CURLE_OK)
		{
			curl_slist_free_all(_headers);
			curl_easy_cleanup(_curl);
			return true;
		}

		curl_slist_free_all(_headers);
		curl_easy_cleanup(_curl);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool MarsClient::HttpGetData(char *_respone, const char *_request)
{
	try
	{
		if (_respone == NULL || _request == NULL) return false;
		
		bool _status = false;
		char _req[1024] = { EOF };

		sprintf(_req, "%s%s", _Host, _request);

		if (HttpGET(_respone, _req, _Token))
			_status = true;
				
		return _status;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool MarsClient::HttpPostData(char *_respone, const char *_request, const char *_payload)
{
	try
	{
		if (_respone == NULL || _request == NULL) return false;
		
		bool _status = false;
		char _req[1024] = { EOF };

		sprintf(_req, "%s%s", _Host, _request);

		if (HttpPost(_respone, _req, _payload, strlen(_payload), _Token))
			_status = true;
			
		return _status;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool MarsClient::DoLogin(const char *_host)
{
	try
	{	
		char _req[1024];
		sprintf(_req, "%s/auth/login?usr=%s&pwd=%s&proj=%s", _host, _Account, _Password, _Proj);

		bool _status = false;

		if(HttpGET(_Token, _req, NULL))
		{
			strcpy(_Host, _host);

			_Token[512] = EOF;
			_status = true;
		}

		return _status;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
	return false;
}
//--------------------------------------------------------------
bool MarsClient::RegistryDevice(const char *_vendor, const char *_uuid, const char *_suid, const char *_type)
{      	
	try
	{
		char _resp[128] = { EOF };
		char _info[256] = { EOF };

		sprintf(_info, "{ \"vendor\":\"%s\", \"uuid\":\"%s\", \"suid\":\"%s\", \"profile\":\"%s\" }", _vendor, _uuid, _suid, _type);

		return HttpPostData(_resp, "/auth/registry?target=device", _info);    
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
	return false;
}
//--------------------------------------------------------------
bool MarsClient::GetDataSrcList(char *_result)
{
	try
	{
		return HttpGetData(_result, "/api/usrinfo?method=datasrclist");
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool MarsClient::GetLastData(const char *_uuid, const char *_suid, int _count, char *_result)
{
	try
	{
		char _info[256] = { EOF };

		sprintf(_info, "{ \"uuid\":\"%s\", \"suid\":\"%s\", \"count\":%d }", _uuid, _suid, _count);

		return HttpGetData(_result, "/api/lastdata?method=read");
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
#endif
