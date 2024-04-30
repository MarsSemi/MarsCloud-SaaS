
import MarsClient as marsClient
import MarsMQTT as marsMQTT
import json
import time
#--------------------------------------------------------------
#
#--------------------------------------------------------------
_Host = "test.mars-cloud.com"
_Client = marsClient.MarsClient()
_MQTT = marsMQTT.MarsMQTT()
#--------------------------------------------------------------
def MQTTCallback(_topic, _msg): 
    try:
        print("MQTT Receive : "+_topic)
        print("MQTT Msg : "+_msg)
    except:
        return
#--------------------------------------------------------------
def TestAPIs(): 
    try:
        if _Client.Login("https://"+_Host, "test", "test", "justtest"):

            TestMQTT()
            
            _resp = _Client.RegDevice("dev", "test", "both.tempmeter", "VirtualDevice", "com.test")
            print("RegDevice : "+str(_resp))

            _resp = _Client.PutData("dev", "test", { "temp": 23.5, "humi": 87 })
            print("PutData : "+str(_resp))

            _resp = json.loads(_Client.GetLastData("dev", "test", 1));
            _resp = _resp['results'][0]        
            print("Get Last Data : "+str(_resp))

            _resp = _Client.RemoveData("dev", "test", _resp['ukey'])
            print("Remove Data : "+str(_resp))

            _resp = _Client.CallService("service.myService", "/api/hello", None)
            print("Call Service : "+str(_resp))
    except:
        return

#--------------------------------------------------------------
def TestMQTT(): 
    try:
        if _MQTT.Connect(_Host, 8884, _Client.User, _Client.Token, MQTTCallback):
            print("MQTT Connect OK !")

            _MQTT.Subscribe("test/+/#")
            _Client.PushMessage('test/my/msg', { "test": 12345 })
    except:
        return
#--------------------------------------------------------------
def main(): 
    try:
        TestAPIs()

        while True:
            time.sleep(0.1)
    except:
        return
#--------------------------------------------------------------
#
#--------------------------------------------------------------
if __name__=="__main__": 
    main() 
#--------------------------------------------------------------
