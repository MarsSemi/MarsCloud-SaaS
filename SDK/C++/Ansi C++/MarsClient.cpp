//--------------------------------------------------------------
#include "MarsClient.h"
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

		if(_payload->_buf != NULL && _payload->_size > 0)
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
MarsClient::MarsClient(const char *_account, const char *_password, const char *_proj)
{
	try
	{
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
		curl_easy_setopt(_curl, CURLOPT_TIMEOUT, _DefaultTimeOut_Sec);
		curl_easy_setopt(_curl, CURLOPT_ACCEPTTIMEOUT_MS, _DefaultAcceptTimeOut_MSec);
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
		curl_easy_setopt(_curl, CURLOPT_TIMEOUT, _DefaultTimeOut_Sec);	
		curl_easy_setopt(_curl, CURLOPT_ACCEPTTIMEOUT_MS, _DefaultAcceptTimeOut_MSec);
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
		char _req[1024] = { '\0' };

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

		if (HttpPost(_respone, _req, _payload, _payload == NULL ? 0 : strlen(_payload), _Token))
			_status = true;

		return _status;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
const char *MarsClient::GetHost(void)
{
	try
	{
		if(strlen(_Host) > 0)
			return _Host;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return NULL;
}
//--------------------------------------------------------------
const char *MarsClient::GetAccount(void)
{
	try
	{
		if(strlen(_Account) > 0)
			return _Account;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return NULL;
}
//--------------------------------------------------------------
const char *MarsClient::GetToken(void)
{
	try
	{
		if(strlen(_Token) > 0)
			return _Token;
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return NULL;
}
//--------------------------------------------------------------
bool MarsClient::DoLogin(const char *_host)
{
	try
	{	
		char _req[1024];
		char _info[256] = { '\0' };

		sprintf(_info, "{ \"usr\":\"%s\", \"pwd\":\"%s\", \"proj\":\"%s\" }", _Account, _Password, _Proj);
		sprintf(_req, "%s/auth/login?", _host);

		bool _status = false;

		if(HttpPostData(_Token, _req, _info))
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
bool MarsClient::RegDevice(const char *_uuid, const char *_suid, const char *_type, const char *_name, const char *_vender)
{      	
	try
	{
		char _resp[128] = { '\0' };
		char _info[256] = { '\0' };

		sprintf(_info, "{ \"uuid\":\"%s\", \"suid\":\"%s\", \"data_profile\":\"%s\", \"name\":\"%s\", \"vender\":\"%s\" }", _uuid, _suid, _type, _name, _vender);

		return HttpPostData(_resp, "/api/usrinfo?method=adddatasrc", _info);    
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }	
	return false;
}
//--------------------------------------------------------------
bool MarsClient::PutMessage(const char *_topic, const char *_postData)
{
	try
	{
		char _resp[16] = { '\0' };
		char _api[512] = { '\0' };

		sprintf(_api, "/api/put?message&topic=%s", _topic);

		return HttpPostData(_resp, _api, _postData);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool MarsClient::PutData(const char *_uuid, const char *_suid, const char *_content)
{
	try
	{
		char _resp[128] = { '\0' };
		char _info[256] = { '\0' };

		sprintf(_info, "{ \"uuid\":\"%s\", \"suid\":\"%s\", \"values\": [%s] }", _uuid, _suid, _content);

		return HttpPostData(_resp, "/api/put?data", _info);
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool MarsClient::GetLastData(const char *_uuid, const char *_suid, int _count, char *_result)
{
	try
	{
		char _info[256] = { '\0' };

		sprintf(_info, "{ \"uuid\":\"%s\", \"suid\":\"%s\", \"count\":%d }", _uuid, _suid, _count);

		return HttpGetData(_result, "/api/lastdata?method=read");
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool MarsClient::RemoveData(const char *_uuid, const char *_suid, const char *_ukey, char *_result)
{
	try
	{
		char _info[256] = { '\0' };

		sprintf(_info, "{ \"uuid\":\"%s\", \"suid\":\"%s\", \"ukey\":\"%s\" }", _uuid, _suid, _ukey);

		return HttpGetData(_result, "/api/del?data");
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------
bool MarsClient::CallService(const char *_service, const char *_api, const char *_postData, char *_result)
{
	try
	{
		if(strcmp(_service, "service.databroker"))
			return HttpPostData(_result, _api, _postData);
		else
		{
			char _cmd[512];
			sprintf(_cmd, "/services/%s/%s", _service, _api);
			return HttpPostData(_result, _cmd, _postData);
		}
	}
	catch(...){ printf("Func Exception : %s\n", __func__); }
	return false;
}
//--------------------------------------------------------------