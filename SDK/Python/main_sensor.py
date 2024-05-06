
import json
import time
import asyncio
import struct
import os.path

from bleak import BleakScanner
from bleak import BleakClient, BleakGATTCharacteristic
#--------------------------------------------------------------
import MarsClient as marsClient
import json
#--------------------------------------------------------------
#
#--------------------------------------------------------------
class MiDevice:
    #--------------------------------------------------------------
    def __init__(self, _devUUID):

        self.bleDevice = _devUUID
        self.bleClient = None
        
        self.FilePath = 'dev.txt'
        self.GET_DATA_UUID = 'EBE0CCC1-7A0A-4B0C-8A1A-6FF2997DA3A6'
    #--------------------------------------------------------------
    def SaveDev(self, _devID): 
        try:

            _f = open(self.FilePath, 'w')
            _f.write(_devID)
            _f.close()

        except Exception as _e:
            print(_e)
            return
    #--------------------------------------------------------------
    def LoadDev(self): 
        try:

            if os.path.isfile(self.FilePath) :

                _f = open(self.FilePath, 'r')

                if _f != None :
                    _devID = _f.read()
                    _f.close()
                    return _devID

            return None

        except Exception as _e:
            print(_e)
            return None
    #--------------------------------------------------------------
    async def ScanDev(self): 
        try:
            print("")
            print("---------- Rescan BLE Dev ----------")
            
            _devices = await BleakScanner.discover()

            if _devices != None:
                for _d in _devices:
                    if "LYWSD03MMC" in _d.name:
                        print("Found : "+_d.name+ " / "+_d.address)

                        self.bleDevice = _d.address
                        self.SaveDev(self.bleDevice)
                        return

        except Exception as _e:
            #print(_e)
            return
    #--------------------------------------------------------------
    async def ConnectDev(self) -> bool: 
        try:
            
            if self.bleDevice == None:
                self.bleDevice = self.LoadDev()

            if self.bleDevice == None:
                await self.ScanDev()
                if self.bleDevice == None:
                    return False

            print("Try to connect : "+self.bleDevice)

            if self.bleClient == None:
                self.bleClient = BleakClient(self.bleDevice)
            
            if await self.bleClient.connect():
                print("Connect Success : "+self.bleDevice)
                return True

            return False
        except Exception as _e:
            print(_e)
            self.bleClient = None
            return False
    #--------------------------------------------------------------
    def IsConnected(self) -> bool: 
        try:
            
            if self.bleClient != None:
                return self.bleClient.is_connected

            return False
        except Exception as _e:
            #print(_e)
            return False
    #--------------------------------------------------------------
    async def ReadData(self) -> bool: 
        try:

            if self.IsConnected():
                _byteArray = await self.bleClient.read_gatt_char(self.GET_DATA_UUID)
                #print(_byteArray.hex())

                if _byteArray != None and len(_byteArray) >= 5 :

                    self.temp = int.from_bytes(_byteArray[0:2], byteorder='little', signed=True) / 100
                    self.humi = int.from_bytes(_byteArray[2:3], byteorder='little')
                    self.volt = int.from_bytes(_byteArray[3:5], byteorder='little') / 1000
                    self.battery = round((_volt - 2) / (3 - 2) * 100, 2)

                    if self.battery > 100 :
                        self.battery = 100

                    if self.battery < 0 :
                        self.battery = 0

                    print("-")
                    print(str(self.temp)+"ºC | "+str(self.humi)+"% | "+str(self.battery)+"%")

                    return True

            return False
        except Exception as _e:
            print(_e)
            return False
#--------------------------------------------------------------
_Host = "test.mars-cloud.com"
_Client = marsClient.MarsClient()
#--------------------------------------------------------------
def main(): 
    try:
        
        _startTime = time.time()

        while True :

            _startTime = time.time()
            _Client.Token = None

            if _Client.Login("https://"+_Host, "test", "test", "justtest") == False :
                time.sleep(15)

            if _Client.Token != None:

                print("Login Success")

                _dev = MiDevice(None)
                _loop = asyncio.new_event_loop()

                while time.time() - _startTime < 86400 :
                    
                    if _dev.IsConnected() == False:
                        _loop.run_until_complete(_dev.ConnectDev())

                    if _dev.IsConnected():
                        if _loop.run_until_complete(_dev.ReadData()) :
                            _Client.PutData("dev", "test", { "temp": _dev.temp, "humi": _dev.humi, "battery": _dev.battery })

                        time.sleep(30)
                    else:
                        time.sleep(0.1)

    except Exception as _e:
            print(_e)
#--------------------------------------------------------------
#
#--------------------------------------------------------------
if __name__=="__main__": 
    main()
#--------------------------------------------------------------
