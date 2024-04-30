package com.mars.cloud;

import android.app.Activity;
import android.os.StrictMode;
import android.provider.Settings;

import org.eclipse.paho.client.mqttv3.IMqttDeliveryToken;
import org.eclipse.paho.client.mqttv3.MqttCallback;
import org.eclipse.paho.client.mqttv3.MqttClient;
import org.eclipse.paho.client.mqttv3.MqttConnectOptions;
import org.eclipse.paho.client.mqttv3.MqttMessage;
import org.eclipse.paho.client.mqttv3.persist.MemoryPersistence;
import org.json.JSONArray;
import org.json.JSONObject;

import java.io.InputStream;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;

//---------------------------------------------------------------------------------------
//
//---------------------------------------------------------------------------------------
public class MarsClient
{
    //---------------------------------------------------------------------------------------
    private class ClientMQTTCallback implements MqttCallback
    {
        //---------------------------------------------------------------------------------------
        public ClientMQTTCallback(){}
        //---------------------------------------------------------------------------------------
        public void connectionLost(Throwable _cause)
        {
            try
            {
                if(_AuthToken.length() > 0) MQTTReconnect();
                for(int i=0;i<_MessageCallback.size();i++)
                    _MessageCallback.get(i).connectionLost(_cause);
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        }
        //---------------------------------------------------------------------------------------
        public void messageArrived(String _topic, MqttMessage _msg)
        {
            try
            {
                for(int i=0;i<_MessageCallback.size();i++)
                    _MessageCallback.get(i).messageArrived(_topic, _msg);
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        }
        //---------------------------------------------------------------------------------------
        public void deliveryComplete(IMqttDeliveryToken token)
        {
            try
            {
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        }
        //---------------------------------------------------------------------------------------
    };
    //---------------------------------------------------------------------------------------
    //
    //---------------------------------------------------------------------------------------
    private static String _MarsCloudHost = "test.mars-cloud.com";
    public static String _MarsCloudURL = "https://"+_MarsCloudHost;
    public static String _MarsCloudMQTTURL = "ssl://"+_MarsCloudHost+":8883";
    public static int _DefaultConnectTimeOut = 10000;
    public static String _DefaultCharset = "UTF-8";
    //---------------------------------------------------------------------------------------
    public MarsClient _thisIterator = null;
    public MqttClient _MQTTClient = null;
    //---------------------------------------------------------------------------------------
    private ClientMQTTCallback _MQTTCallback = null;
    private ArrayList<MqttCallback> _MessageCallback = new ArrayList<MqttCallback>();
    //---------------------------------------------------------------------------------------
    private String _AuthToken = "";
    private String _Account = "";
    private String _Password = "";
    private String _Project = "";
    //---------------------------------------------------------------------------------------
    public MarsClient(String _account, String _password, String _proj)
    {
        try
        {
            _thisIterator = this;
            _Account = _account;
            _Password = _password;
            _Project = _proj;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //---------------------------------------------------------------------------------------
    public void SetCloudServerURL(String _url)
    {
        try
        {
            if(_url.startsWith("http"))
            {
                _MarsCloudHost = _url.replace("http://", "").replace("https://", "").split(":")[0];
                _MarsCloudURL = _url;
                _MarsCloudMQTTURL = _url.startsWith("https") ? "ssl://"+_MarsCloudHost+":8883" : "tcp://"+_MarsCloudHost+":1883";
            }
            else
            {
                boolean _isIP = true;
                String[] _items = _url.split(":");
                String _tempURL = _items[0].replaceAll("\\.", "");

                for (int i = 0; i < _tempURL.length() && _isIP; i++)
                    _isIP = Character.isDigit(_tempURL.charAt(i));

                _MarsCloudHost = _items[0];
                _MarsCloudURL = (_isIP ? "http://" : "https://")+_url;
                _MarsCloudMQTTURL = _isIP ? "tcp://"+_MarsCloudHost+":1883" : "ssl://"+_MarsCloudHost+":8883";
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //---------------------------------------------------------------------------------------
    public String GetCloudServerURL()
    {
        try
        {
            return _MarsCloudURL;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return "";
    }
    //---------------------------------------------------------------------------------------
    private String HttpGet(final String _cmd, final AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            if(_callback == null) return Tools.HttpGet(_MarsCloudURL+_cmd, _AuthToken, _DefaultConnectTimeOut);
            new Thread(new Runnable()
            {
                public void run()
                {
                    try
                    {
                        String _resp = Tools.HttpGet(_MarsCloudURL+_cmd, _AuthToken, _DefaultConnectTimeOut);
                        if(_callback != null) _callback.OnData(_resp);
                        return;
                    }
                    catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
                    if(_callback != null) _callback.OnData("");
                }
            }).start();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    private String HttpGet(final String _cmd, final int _timeout , final AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            if(_callback == null) return Tools.HttpGet(_MarsCloudURL+_cmd, _AuthToken, _timeout);
            new Thread(new Runnable()
            {
                public void run()
                {
                    try
                    {
                        String _resp = Tools.HttpGet(_MarsCloudURL+_cmd, _AuthToken, _timeout);
                        if(_callback != null) _callback.OnData(_resp);
                        return;
                    }
                    catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
                    if(_callback != null) _callback.OnData("");
                }
            }).start();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    private String HttpGet(final String _cmd, String _auth)
    {
        try
        {
            return Tools.HttpGet(_MarsCloudURL+_cmd, _auth, _DefaultConnectTimeOut);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    private String HttpPost(final String _cmd, final String _payload, final AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            if(_callback == null) return Tools.HttpPost(_MarsCloudURL+_cmd, _AuthToken, _payload, _DefaultConnectTimeOut);
            new Thread(new Runnable()
            {
                public void run()
                {
                    try
                    {
                        String _resp = Tools.HttpPost(_MarsCloudURL+_cmd, _AuthToken, _payload, _DefaultConnectTimeOut);
                        if(_callback != null) _callback.OnData(_resp);
                        return;
                    }
                    catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
                    if(_callback != null) _callback.OnData("");
                }
            }).start();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public boolean AddMQTTCallback(MqttCallback _callback)
    {
        try
        {
            _MessageCallback.add(_callback);
            return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //---------------------------------------------------------------------------------------
    public int GetMQTTCallBackLength()
    {
        try
        {
            if(_MessageCallback != null) return _MessageCallback.size();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return 0;
    }
    //---------------------------------------------------------------------------------------
    private boolean MQTTReconnect()
    {
        try
        {
            if(_MQTTClient != null && _MQTTClient.isConnected()) { _MQTTClient.disconnect(); _MQTTClient.close(); }
            if(_AuthToken != null && _AuthToken.length() > 0)
            {
                if(_MQTTCallback == null) _MQTTCallback = new ClientMQTTCallback();

                MqttConnectOptions _options = new MqttConnectOptions();

                _options.setUserName(_Account);
                _options.setPassword(_AuthToken.toCharArray());
                _options.setConnectionTimeout(_DefaultConnectTimeOut);
                _options.setKeepAliveInterval(_DefaultConnectTimeOut);
                _options.setCleanSession(true);

                _MQTTClient = new MqttClient(_MarsCloudMQTTURL, _Account + '@'+new Date().getTime()%1000, new MemoryPersistence());
                _MQTTClient.setCallback(_MQTTCallback);
                _MQTTClient.connect(_options);

                Tools.Log.Print(Tools.LogLevel.ll_Info, "MQTT connect SUCCESS");
                return true;
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        Tools.Log.Print(Tools.LogLevel.ll_Warning, "MQTT connect FAIL");
        return false;
    }
    //---------------------------------------------------------------------------------------
    public boolean MQTTSubscribe(String _topic)
    {
        try
        {
            if(_MQTTClient != null && _MQTTClient.isConnected())
            {
                _topic = _topic.isEmpty() ? _Account + "+/#" : _topic;
                _MQTTClient.subscribe(_topic);

                Tools.Log.Print(Tools.LogLevel.ll_Info, "MQTT subscribe SUCCESS : "+_topic);
                return true;
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        Tools.Log.Print(Tools.LogLevel.ll_Warning, "MQTT subscribe FAIL : "+_topic);
        return false;
    }
    //---------------------------------------------------------------------------------------
    public static boolean RegistryAccount(String _account, String _password, String _email)
    {
        try
        {
            JSONObject _payload = new JSONObject();
            _payload.put("id", _account);
            _payload.put("name", _account);
            _payload.put("password", _password);
            _payload.put("email", _email);

            if(_MarsCloudURL.length() > 0)
            {
                String _resp = Tools.HttpPost(_MarsCloudURL+"/auth/registry?target=user", "", _payload.toString(), _DefaultConnectTimeOut);
                if(_resp != null && _resp.length() > 0) return true;
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        Tools.Log.Print(Tools.LogLevel.ll_Warning, "Registry Fail ~");
        return false;
    }
    //---------------------------------------------------------------------------------------
    public void RegistryDevice(String _name, String _uuid, String _suid, String _vendor, String _ip, String _profile)
    {
        try
        {
            JSONObject _payload = new JSONObject();
            _payload.put("name", _name);
            _payload.put("uuid", _uuid);
            _payload.put("suid", _suid);
            _payload.put("ext1", _ip);
            _payload.put("vender_id", _vendor);
            _payload.put("data_profile",_profile);

            //Tools.Log.Print(Tools.LogLevel.ll_Debug , "Auth:" + _AuthToken);
            //Tools.Log.Print(Tools.LogLevel.ll_Debug , "Obj:" + _payload.toString());

            Tools.HttpPost(_MarsCloudURL+"/api/usrinfo?method=adddatasrc", _AuthToken, _payload.toString(), _DefaultConnectTimeOut);
            Tools.Log.Print(Tools.LogLevel.ll_Info, "RegistryDevice : "+_name);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    // --------------------------------------------------------------------------------------
    public String PutData(String _uuid , String _suid, JSONArray _values)
    {
        try
        {
            JSONObject _obj = new JSONObject();
            _obj.put("uuid"	, _uuid);
            _obj.put("values"	, _values);

            if(_suid != null && _suid.length() > 0)	_obj.put("suid"	, _suid);

            Tools.Log.Print(Tools.LogLevel.ll_Debug , "Obj :" + _obj.toString());
            return Tools.HttpPost( _MarsCloudURL + "/api/put?data", _AuthToken , _obj.toString() , 3000);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public void SavePassword(JSONObject _userinfo , String _password, final AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            if(_userinfo != null)
            {
                _userinfo.put("password" , _password);
                Tools.Log.Print(Tools.LogLevel.ll_Info, "Try to Login and SavePassword : "+_MarsCloudURL);
                HttpPost("/api/usrinfo?method=write&full_apply=true" , _userinfo.toString(),  new AbstractObject.IStringDataCallback()
                {
                    public void OnData(String _payload)
                    {
                        try
                        {
                            if(_callback != null) _callback.OnData(_AuthToken);
                        }
                        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
                    }
                });
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //---------------------------------------------------------------------------------------
    public boolean Start()
    {
        Start((AbstractObject.IStringDataCallback )null);
        return false;
    }
    //---------------------------------------------------------------------------------------
    public boolean IsConnected()
    {
        return _AuthToken.isEmpty() ? false : true;
    }
    //---------------------------------------------------------------------------------------
    public void Start(final AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            JSONObject _payload = new JSONObject();

            _payload.put("usr", _Account);
            _payload.put("pwd", _Password);
            _payload.put("proj", _Project);

            Tools.Log.Print(Tools.LogLevel.ll_Info, "Try to Login : "+_MarsCloudURL);
            HttpPost("/auth/login?", _payload.toString(), new AbstractObject.IStringDataCallback()
            {
                public void OnData(String _payload)
                {
                    try
                    {
                        _AuthToken = _payload;

                        MQTTReconnect();

                        if(_AuthToken.isEmpty() == false) Tools.Log.Print(Tools.LogLevel.ll_Info, "Login SUCCESS");
                        if(_callback != null) _callback.OnData(_AuthToken);
                    }
                    catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
                }
            });
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //---------------------------------------------------------------------------------------
    public void Start(String _token)
    {
        try
        {
            Tools.Log.Print(Tools.LogLevel.ll_Info, "Try to Login with token : "+_MarsCloudURL);
            _AuthToken = _token;
            MQTTReconnect();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //---------------------------------------------------------------------------------------
    public void Start(String _token , boolean _mqttFlag)
    {
        try
        {
            _AuthToken = _token;
            if(_mqttFlag) MQTTReconnect();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //---------------------------------------------------------------------------------------
    public boolean Stop()
    {
        try
        {
            if(_MessageCallback != null) _MessageCallback.clear();
            if(_MQTTClient != null)
            {
                _MQTTClient.setCallback(null);
                _MQTTClient.close();
            }

            _AuthToken = "";
            _MQTTClient = null;

            return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        _MQTTClient = null;
        return false;
    }
    //---------------------------------------------------------------------------------------
    public String GetAuthToken()
    {
        try
        {
            return _AuthToken;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return "";
    }
    //---------------------------------------------------------------------------------------
    public String GetUserInfo(AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            return HttpPost("/api/usrinfo?method=read", "", _callback);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public String SetUserInfo(JSONObject _payload, AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            return HttpPost("/api/usrinfo?method=write", _payload.toString(), _callback);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public String GetDataSrcList(AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            return HttpPost("/api/usrinfo?method=datasrclist", "", _callback);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public String GetDataSrcInfo(String _src_id, AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            JSONObject _payload = new JSONObject();
            _payload.put("uuid", _src_id.replace('.', '_'));

            return HttpPost("/api/usrinfo?method=datasrcinfo", _payload.toString(), _callback);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public String GetData(String _uuid, String _suid, long _start_time, long _end_time, AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            JSONObject _payload = new JSONObject();

            _payload.put("src", _suid.length() <= 0 ? _uuid : _uuid+"_"+_suid);
            _payload.put("timestamp", ""+_start_time+"~"+_end_time);

            return HttpPost("/api/get?data", _payload.toString(), _callback);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public String GetLastData(String _uuid, String _suid, int count, AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            JSONObject _payload = new JSONObject();

            _payload.put("src", _suid.length() <= 0 ? _uuid : _uuid+"_"+_suid);
            _payload.put("count", count);

            return HttpPost("/api/lastdata?methid=read", _payload.toString(), _callback);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public boolean PutEvent(String _uuid, String _suid, JSONObject _payload)
    {
        try
        {
            JSONObject _event = new JSONObject();
            JSONArray _values = new JSONArray();

            _values.put(_payload);

            _event.put("uuid", _uuid);
            _event.put("suid", _suid);
            _event.put("from", "server");
            _event.put("values", _values);

            String _res = HttpPost("/api/put?event", _event.toString(), null);
            if(_res != null && _res.length() > 0) return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //---------------------------------------------------------------------------------------
    public boolean PutData(String _uuid, String _suid, JSONObject _payload)
    {
        try
        {
            JSONObject _event = new JSONObject();
            JSONArray _values = new JSONArray();

            _values.put(_payload);

            if(_uuid.length() > 0)  _event.put("uuid", _uuid);
            if(_suid.length() >  0)  _event.put("suid", _suid);

            _event.put("from", "server");
            _event.put("values", _values);

            String _res = HttpPost("/api/put?data", _event.toString(), null);
            if(_res != null && _res.length() > 0) return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //---------------------------------------------------------------------------------------
    public boolean PutMessage(String _topic, JSONObject _msg)
    {
        try
        {
            JSONObject _event = new JSONObject();

            _topic = _topic.replaceAll("/", "\\.");

            String _res = HttpPost("/api/put?message&topic="+_topic, _msg.toString(), null);
            if(_res != null && _res.length() > 0) return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //---------------------------------------------------------------------------------------
}
