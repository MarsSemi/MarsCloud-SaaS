from paho.mqtt import client as mqtt_client
import random
import certifi
import time
#--------------------------------------------------------------
#
#--------------------------------------------------------------
def OnConnect(_client, _data, flags, _resultCode): 
    try:
        if _resultCode == 0:
            _client.is_connected = True
    except:
        return
#--------------------------------------------------------------
def OnDisconnect(_client, _data, flags, _resultCode): 
    try:
        time.sleep(1)
        _client.connect(_client.Parent.Host, port = _client.Parent.Port, keepalive=60)
    except:
        return
#--------------------------------------------------------------
def OnMessage(_client, _data, _msg): 
    try:
        if _client.UserCallback != None:
            _client.UserCallback(_msg.topic, _msg.payload.decode())
    except:
        return
#--------------------------------------------------------------
class MarsMQTT:
#--------------------------------------------------------------
    def Connect(_self, _host, _port, _user, _token, _msgCallback): 
        try:
            _tick = 0
            _client = mqtt_client.Client(client_id = _token+'@'+str(random.randint(1000, 2000)), transport = "websockets", protocol = mqtt_client.MQTTv311)

            _self.Host = _host
            _self.Port = _port
            _self.Token = _token
            _self.Client = _client
            _self.Client.Parent = _self
            _self.Client.UserCallback = _msgCallback

            _client.is_connected = False
            _client.on_connect = OnConnect
            _client.on_disconnect = OnDisconnect
            _client.on_message = OnMessage

            _client.tls_set(certifi.where())
            _client.username_pw_set(_user, _token)            
            _client.connect(_host, port = _port, keepalive=60)
            _client.loop_start()

            while _client.is_connected == False:
                time.sleep(0.1)

                _tick += 1 
                if _tick > 600:
                    break

            return True
        except:
            return False
#--------------------------------------------------------------
    def Subscribe(_self, _topic): 
        try:
            _self.Client.Topic = _topic
            _self.Client.subscribe(_topic)
            
            return True
        except:
            return False
#--------------------------------------------------------------
