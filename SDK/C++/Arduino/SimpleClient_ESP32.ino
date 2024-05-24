//--------------------------------------------------------------
#include "libraries/MarsClient.h"
//--------------------------------------------------------------
#include <BLEDevice.h> 
#include <BLEScan.h>
//--------------------------------------------------------------
//
//--------------------------------------------------------------
#define LED 13
//--------------------------------------------------------------
SET_LOOP_TASK_STACK_SIZE(12*1024); // Very important! Without this, ble get service might crash
//--------------------------------------------------------------
BLEAddress _BleTargetAddr("");
BLEClient *_BleClient = NULL;
BLERemoteService *_BleService = NULL;
//--------------------------------------------------------------
//
//--------------------------------------------------------------
class BLEClientCallback : public BLEClientCallbacks
{
  void onConnect(BLEClient *pclient){}
  void onDisconnect(BLEClient *pclient){}
};
//--------------------------------------------------------------
//
//--------------------------------------------------------------
MarsClient *_Client = NULL;

unsigned long _LoopTick = 0;
unsigned char _ErrorTick = 0;

float _Tempture = 0;
float _Humidity = 0;
float _Battery = 0;
//--------------------------------------------------------------
void initENV()
{
  try
  {
    delay(1000);
    pinMode(LED, OUTPUT);

    Serial.begin(115200);
    Serial.println("--- System Start ---");
  }
  catch(...){}
}
//--------------------------------------------------------------
void turnOnLED()
{
  try
  {
    digitalWrite(LED, HIGH);
  }
  catch(...){}
}
//--------------------------------------------------------------
void turnOffLED()
{
  try
  {
    digitalWrite(LED, LOW);
  }
  catch(...){}
}
//--------------------------------------------------------------
void blinkLED(int _interval)
{
  try
  {
    turnOnLED();
    delay(_interval);
    turnOffLED();
  }
  catch(...){}
}
//--------------------------------------------------------------
void resetWiFi()
{
  try
  {
    Serial.println("--- Ty Connect WiFi ---");

    int _tick = 0;

    WiFi.mode(WIFI_STA);
    WiFi.begin("mars3_2.4G", "57525670");

    while (WiFi.status() != WL_CONNECTED)
    {
      _tick++;

      if(_tick % 20 == 0)
      {
        Serial.println("");
        _tick = 0;
      }

      delay(500);
      Serial.print(".");
    }

    Serial.println("");
  }
  catch(...){}
}
//--------------------------------------------------------------
void connetBLEDevice()
{
  try
  {
    Serial.println("--- Try Connect BLE Device ---");
    
    BLEDevice::init("");
    BLEScan *_scanner = BLEDevice::getScan();
    int _targetCount = 0;

    _scanner->setActiveScan(true);

    while(_targetCount <= 0)
    {
      Serial.println("Ble Device Scanning ...");
      BLEScanResults _results = _scanner->start(5, true);

      for(int i=0;i<_results.getCount();i++)
      {
        BLEAdvertisedDevice _dev = _results.getDevice(i);

        if(_dev.getName().compare("LYWSD03MMC") == 0)
        {
          _BleTargetAddr = _dev.getAddress();
          _targetCount++;

          Serial.printf("BLE Device : %s\n", _BleTargetAddr.toString().c_str());
          break;
        }
      }
    }
      
    if(_BleClient == NULL)
    {
      _BleClient = BLEDevice::createClient();

      _BleClient->connect(_BleTargetAddr);
      _BleClient->setMTU(512); // Very important! Without this, ble get service might crash

      if(_BleClient->isConnected())
        Serial.printf("BLE Device connected : %s\n", _BleTargetAddr.toString().c_str());
    }
  }
  catch(...){}
}
//--------------------------------------------------------------
void loginSystem()
{
  try
  {
    Serial.println("--- Try Login Cloud System ---");

    if(WiFi.status() == WL_CONNECTED)
    {
      Serial.println("\nWifi Connected : "+WiFi.localIP().toString());

      _Client = new MarsClient("test.mars-cloud.com", 443);

      do
      {
        Serial.println("Try to login ...");
      }
      while(_Client->Login("test", "test", "justtest") == false);

      Serial.println(F("[Mars] Login SUCCESS"));  

      if(_Client->RegDevice("dev", "test", "both.temp"))
        Serial.println(F("RegDevice SUCCESS"));  
    }
  }
  catch(...){}
}
//--------------------------------------------------------------
void uploadData(float _temp, float _humi, float _battery)
{
  try
  {
    if(_Client == NULL) return;
    if(_Client->IsLogin() == false) return;

    StaticJsonDocument<256> _docItem;
    StaticJsonDocument<8> _docArray;

    JsonObject _doc = _docItem.to<JsonObject>();
    JsonArray _array = _docArray.to<JsonArray>();

    _doc["temp"] = _temp;
    _doc["humi"] = _humi;
    _doc["battery"] = _battery;
    _array.add(_doc);

    turnOnLED();

    if(_Client->PutData("dev", "test", _array))
    {
      Serial.println(F("PutData SUCCESS"));  
      _ErrorTick = 0;
    }
    else
      _ErrorTick++;

    turnOffLED();
  }
  catch(...){}
  
}
//--------------------------------------------------------------
void setup()
{
  try
  {
    initENV();

    resetWiFi();
    blinkLED(100);

    loginSystem();
    blinkLED(100);

    connetBLEDevice();
    blinkLED(100);
  }
  catch(...){}
}
//--------------------------------------------------------------
void loop()
{
  try
  {
    if(WiFi.status() != WL_CONNECTED) resetWiFi();
    if(_BleClient->isConnected())
    {
      std::string _value = _BleClient->getValue(BLEUUID("ebe0ccb0-7a0a-4b0c-8a1a-6ff2997da3a6"), BLEUUID("ebe0ccc1-7a0a-4b0c-8a1a-6ff2997da3a6"));

      if(_value.length() >= 5)
      {
        const char *_pValue = _value.c_str();
        float _volt = ((_pValue[4] * 256) + _pValue[3]) / 1000.0f;

        _Tempture = (_pValue[0] | (_pValue[1] << 8)) * 0.01f;
        _Humidity = _pValue[2];
        _Battery = (_volt - 2.1) * 100.0f;

        if(_Battery > 100) _Battery = 100;

        Serial.printf("Temp : %.02fºC, Humi : %.01f%, Battery : %.01f\n", _Tempture, _Humidity, _Battery);
        blinkLED(50);
      }
    }
    else
      _BleClient->connect(_BleTargetAddr);

    if(_LoopTick%15 == 0) uploadData(_Tempture, _Humidity, _Battery);
    if(_LoopTick >= 7200 || _ErrorTick >= 10) ESP.restart();

    _LoopTick++;

    delay(1000);
  }
  catch(...){}
}
//--------------------------------------------------------------
