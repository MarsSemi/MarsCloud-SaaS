//--------------------------------------------------------------
#include "libraries/MarsClient.h"
//--------------------------------------------------------------
MarsClient *_Client = NULL;
//--------------------------------------------------------------
void setup()
{
  // http is no available for test.mars-cloud.com.
  // So this code is just a sample to show how it works.
  // You can find https solution bu searching with "Https Cleint ESP32"

  _Client = new MarsClient(F("http://test.mars-cloud.com"), 80);

  if(_Client->Login("test", "test", "justtest"))
  {
    StaticJsonDocument<512> _item;
    JsonArray _array;

    _item[F("temp")] = 24.8;
    _item[F("humi")] = 74.9;
    _array.add(_item);

    _Client->RegDevice("test", "metter01", "both.temp");
    _Client->PutData("test", "metter01", _array);
  }
}
//--------------------------------------------------------------
void loop()
{

}
//--------------------------------------------------------------
