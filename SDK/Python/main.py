import MarsClient as Mars
import json
#--------------------------------------------------------------
#
#--------------------------------------------------------------
_Client = Mars.MarsClient()
#--------------------------------------------------------------
def main(): 
    if _Client.Login("https://test.mars-cloud.com", "test", "test", "justtest"):

        _resp = _Client.PutData("dev", "test", { "temp": 23.5, "humi": 87 })
        print("PutData : "+str(_resp))

        _resp = json.loads(_Client.GetLastData("dev", "test", 1));
        _resp = _resp['results'][0]
        
        print("Get Last Data : "+str(_resp))

        _resp = _Client.RemoveData("dev", "test", _resp['ukey'])
        print("Remove Data : "+str(_resp))
#--------------------------------------------------------------
#
#--------------------------------------------------------------
if __name__=="__main__": 
    main() 
#--------------------------------------------------------------
