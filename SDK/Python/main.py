import requests
import json
#--------------------------------------------------------------
#
#--------------------------------------------------------------
class MarsClient:
#--------------------------------------------------------------
    def HttpRequest(_self, _req, _payload):
        try:
            _headers = { "Content-Type": "application/json", "Cache-Control": "no-cache" }

            if _self.Token != None:
                _headers["Authentication"] = "Bearer "+_self.Token

            if _payload == None:
                _resp = requests.get(_req, headers = _headers, timeout = 30)
            else:
                _resp = requests.post(_req, data = json.dumps(_payload).encode('utf-8'), headers = _headers, timeout = 30)

            if _resp.status_code == 200:
                return _resp.text
        except:
            return None
#--------------------------------------------------------------
    def Login(_self, _host, _usr, _pwd, _proj): 
        try:
            _self.Host = _host
            _self.User = _usr
            _self.Password = _pwd
            _self.Proj = _proj
            _self.Token = None

            _payload = {}
            _payload['usr'] = _usr
            _payload['pwd'] = _pwd
            _payload['proj'] = _proj

            _resp = _self.HttpRequest(_host+"/auth/login?", _payload)

            if _resp != None:
                _self.Token = _resp
                return True
        except:
            return False
#--------------------------------------------------------------
    def PutData(_self, _uuid, _suid, _data): 
        try:
            _payload = {}
            _payload['uuid'] = _uuid
            _payload['suid'] = _suid
            _payload['values'] = [ _data ]

            _resp = _self.HttpRequest(_self.Host+"/api/put?data", _payload)
            
            if _resp != None:
                return True
        except:
            return False
#--------------------------------------------------------------
    def GetLastData(_self, _uuid, _suid, _count): 
        try:
            _payload = {}
            _payload['uuid'] = _uuid
            _payload['suid'] = _suid
            _payload['count'] = _count

            return _self.HttpRequest(_self.Host+"/api/lastdata?method=read", _payload)
        except:
            return False
#--------------------------------------------------------------
    def RemoveData(_self, _uuid, _suid, _ukey): 
        try:
            _payload = {}
            _payload['uuid'] = _uuid
            _payload['suid'] = _suid
            _payload['ukey'] = _ukey

            _resp = _self.HttpRequest(_self.Host+"/api/del?data", _payload)

            if _resp != None:
                return True
        except:
            return False
#--------------------------------------------------------------
    def CallService(_self, _service, _api, _payload): 
        try:
            
            if _service == "service.databroker":
                return _self.HttpRequest(_self.Host+_api, _payload)
            else:
                return _self.HttpRequest(_self.Host+"/services/"+_service+_api, _payload)
        except:
            return None
#--------------------------------------------------------------
    def PushMessage(_self, _topic, _payload): 
        try:            
            _topic = _topic.replace('/', '.')
            if _self.HttpRequest(_self.Host+"/api/put?message&topic="+_topic, _payload) != None:
                return True
        except:
            return False
#--------------------------------------------------------------
