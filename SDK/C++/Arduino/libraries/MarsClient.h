#include <ArduinoJson.h>
#include <WiFi.h> 
#include <WiFiClientSecure.h>
//--------------------------------------------------------------
class MarsClient
{
private :
  String _ServerIP;  
  String _Token;  
  
  int _ServerPort;
private :
  String HttpGet(String _url);
  String HttpPost(String _url, String _payload);
public:
  MarsClient(String _ip, int _port);
  
  bool IsLogin();
  bool Login(String _user, String _pwd_or_key, String _pro);
  bool RegDevice(String _uuid, String _suid, String _profile);
  bool PutData(String _uuid, String _suid, JsonArray _data);
};
//--------------------------------------------------------------
MarsClient::MarsClient(String _ip, int _port)
{
  _ServerIP = _ip;
  _ServerPort = _port;
}
//--------------------------------------------------------------
String MarsClient::HttpGet(String _url)
{  
  WiFiClientSecure _client;
  String _resp = "";

  _client.setInsecure();

  if(_client.connect(_ServerIP.c_str(), _ServerPort))
  {
    _client.print( "GET "+_url+" HTTP/1.1\r\n" );
    _client.print( F("Connection: close\r\n") );
    
    if(_Token.length() > 0) 
        _client.print("Authentication: Bearer "+_Token+"\r\n");
        
    _client.print(F("\r\n"));
            
    for(int i=0;i<15000;i++)
    {      
      delay(1);
      if(_client.available())
        _resp += _client.readString();

      if(_resp.indexOf(F("200 OK")) == 9)
      {   
        int _startIndex = _resp.indexOf(F("\r\n\r\n")); 
        if(_startIndex >= 0)                              
          return _resp.substring(_startIndex+4);   
      }
    }
  }

  _client.stop();  

  while(_client.connected())
    delay(1);
         
  return _resp;
}
//--------------------------------------------------------------
String MarsClient::HttpPost(String _url, String _payload)
{   
  WiFiClientSecure _client;
  String _resp = "";

  _client.setInsecure();

  if(_client.connect(_ServerIP.c_str(), _ServerPort))
  {                
    _client.print( "POST "+_url+" HTTP/1.1\r\n" );     
    _client.print( "Content-Length: "+String(_payload.length())+"\r\n" );
    _client.print( F("Connection: close\r\n") );
                
    if(_Token.length() > 0)      
      _client.print("Authentication: Bearer "+_Token+"\r\n");
                  
    _client.print(F("\r\n"));
    _client.print(_payload);
                      
    for(int i=0;i<15000;i++)
    {      
      delay(1);

      if(_client.available()) _resp += _client.readString();
      if(_resp.indexOf(F("200 OK")) == 9)
      {   
        int _startIndex = _resp.indexOf(F("\r\n\r\n")); 
        if(_startIndex >= 0)            
        {
          _client.stop();               
          return _resp.substring(_startIndex+4);
        }             
      }
      else
      if(_resp.indexOf(F("HTTP/")) == 0 && _resp.indexOf(F("\r\n\r\n")) >= 0)
        break;
    }
  }

  _resp = "";
  _client.stop();      
  
  return _resp;
}
//--------------------------------------------------------------
bool MarsClient::Login(String _user, String _pwd_or_key, String _proj)
{
  _Token = HttpGet("/auth/login?usr="+_user+"&pwd="+_pwd_or_key+"&proj="+_proj);   
  
  if(IsLogin())
    return true;
  
  return false;
}
//--------------------------------------------------------------
bool MarsClient::IsLogin()
{
  if(_Token != NULL && _Token.length() >= 128)
    return true;
    
  return false;
}
//--------------------------------------------------------------
bool MarsClient::RegDevice(String _uuid, String _suid, String _profile)
{  
  if(_Token.length() <= 0) return false;
  
  StaticJsonDocument<256> _doc;
  String _json;
  
  _doc["uuid"] = _uuid;
  _doc["suid"] = _suid;
  _doc["data_profile"] = _profile;      
  
  serializeJson(_doc, _json);   
  String _resp = HttpPost(F("/api/usrinfo?method=adddatasrc"), _json); 
                        
  if(_resp.length() > 0) 
    return true;

  return false;
}
//--------------------------------------------------------------
bool MarsClient::PutData(String _uuid, String _suid, JsonArray _data)
{  
  if(_Token.length() <= 0) return false;
  
  StaticJsonDocument<512> _doc;
  String _json;
  
  _doc["uuid"] = _uuid;
  _doc["suid"] = _suid;
  _doc["values"] = _data;      
    
  serializeJson(_doc, _json);

  String _resp = HttpPost(F("/api/put?data"), _json);

  if(_resp.length() > 0)   
    return true;
  
  return false;
}
//--------------------------------------------------------------
