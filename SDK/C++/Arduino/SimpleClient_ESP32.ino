//--------------------------------------------------------------
#include "libraries/MarsClient.h"
//--------------------------------------------------------------
MarsClient *_Client = NULL;
//--------------------------------------------------------------
void setup()
{
  int _tick = 0;

  Serial.begin(115200);
  Serial.println("--- System Start ---");

  WiFi.mode(WIFI_STA);
  WiFi.begin("mars3_2.4G", "57525670");

  while (WiFi.status() != WL_CONNECTED)
  {
    if(_tick % 20 == 0)
    {
      Serial.println("");
      _tick = 0;
    }

    delay(500);
    Serial.print(".");
    
    _tick++;
  }
  
  if(WiFi.status() == WL_CONNECTED)
  {
    Serial.println("\nWifi Connected : "+WiFi.localIP().toString());

    _Client = new MarsClient("test.mars-cloud.com", 443);

    Serial.println("Try to login");

    if(_Client->Login("test", "test", "justtest"))
    {
      Serial.println(F("[Mars] Login SUCCESS"));  
      
      StaticJsonDocument<512> _item;
      JsonArray _array;

      _item[F("temp")] = 24.8;
      _item[F("humi")] = 74.9;
      _array.add(_item);

      if(_Client->RegDevice("dev", "test", "both.temp")) Serial.println(F("RegDevice SUCCESS"));  
      if(_Client->PutData("dev", "test", _array)) Serial.println(F("PutData SUCCESS"));  
    }  
  }
}
//--------------------------------------------------------------
void loop()
{

}
//--------------------------------------------------------------
