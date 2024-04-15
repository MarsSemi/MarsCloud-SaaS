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
public class MarsCloudClient extends AbstractObject.ITunnel
{
    //---------------------------------------------------------------------------------------
    private class ClientMQTTCallback implements MqttCallback
    {
        //---------------------------------------------------------------------------------------
        public ClientMQTTCallback(){}
        //---------------------------------------------------------------------------------------
        public void connectionLost(Throwable cause)
        {
            try
            {
                if(_AuthToken.length() > 0) MQTTReconnect();
                for(int i=0;i<_MessageCallback.size();i++) _MessageCallback.get(i).OnDisconnect();
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        }
        //---------------------------------------------------------------------------------------
        public void messageArrived(String topic, MqttMessage message)
        {
            try
            {
                //Tools.Log.Print(Tools.LogLevel.ll_Info, message.toString());
                String _msg = message.toString();
                for(int i=0;i<_MessageCallback.size();i++) _MessageCallback.get(i).OnMessage(_msg);
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
    public class P2PStreamer
    {
        //---------------------------------------------------------------------------------------
        private MarsCloudClient _Parent = null;
        private DatagramSocket _UDPSocket = null;
        private Socket _TCPSocket = null;

        private String _ServerName = "test.mars-cloud.com";
        private String _ServerIP = "";
        private int _ServerPort = 30001;

        private String _RemoteIP_UDP = "";
        private int _RemotePort_UDP = 0;

        private long _updTimeHeartbit = System.currentTimeMillis();
        private int _updConcatOffset = 0;
        //---------------------------------------------------------------------------------------
        P2PStreamer(MarsCloudClient _parent, String _url)
        {
            try
            {
                InetAddress[] _server_address = InetAddress.getAllByName(_url);

                _Parent = _parent;
                _ServerName = _url;
                _ServerIP = _server_address[0].getHostAddress();
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        }
        //---------------------------------------------------------------------------------------
        private void UDPProcessor(DatagramPacket _gPacket, byte[] _temp_buffer, byte[] _concat_buffer, long _t)
        {
            try
            {
                if(_t - _updTimeHeartbit > 5000)
                {
                    SendUDPData(_UDPSocket, _ServerIP, _ServerPort, "heartbit".getBytes(), 0);
                    _updTimeHeartbit = _t;
                }

                DatagramPacket _packet = _gPacket;

                try { if(_UDPSocket != null) _UDPSocket.receive(_packet); } catch(Exception _e){ _packet = null; };
                if (_packet == null || _packet.getLength() <= 0)
                    Thread.sleep(1);
                else
                {
                    int _len = _packet.getLength();
                    byte[] _payload = _packet.getData();

                    if(_packet.getLength() < 512 && _payload[0] == '{') cmdReceived(0, _payload, _len);

                    System.arraycopy(_payload, 0, _concat_buffer, _updConcatOffset, _len);
                    _updConcatOffset += _len;

                    if(_updConcatOffset >= _concat_buffer.length/2) { dataReceived(0, _concat_buffer, _updConcatOffset); _updConcatOffset = 0; }
                    SendUDPData(_UDPSocket, _RemoteIP_UDP, _RemotePort_UDP, "ack".getBytes(), 0);
                }
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        }
        //---------------------------------------------------------------------------------------
        private void TCPProcessor(byte[] _data, long _t)
        {
            try
            {
                InputStream _is = _TCPSocket.getInputStream();
                int _len = _data.length > _is.available() ? _is.available() : _data.length;
                try { _is.read(_data, 0, _len); } catch(Exception _e){ _len = -1; };

                Tools.Log.Print(Tools.LogLevel.ll_Warning, "TCPProcessor : " + _len);

                if (_len <= 0)
                    Thread.sleep(1);
                else
                {

                    dataReceived(0, _data, _len);
                }
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        }
        //---------------------------------------------------------------------------------------
        public boolean Start(final String _srcID)
        {
            try
            {
                if(_UDPSocket == null)
                {
                    _UDPSocket = new DatagramSocket();
                    _UDPSocket.setReceiveBufferSize(512*1024);

                    new Thread(new Runnable()
                    {
                        public void run()
                        {
                            try
                            {
                                JSONObject _regInfo = new JSONObject();

                                _regInfo.put("cmd", "connect");
                                _regInfo.put("target", _srcID);
                                _regInfo.put("ip_lan", Tools.GetLocalIPv4Address());
                                _regInfo.put("max_connection", 1);

                                SendUDPData(_UDPSocket, _ServerIP, _ServerPort, _regInfo.toString().getBytes("utf-8"), 0);

                                _updTimeHeartbit = System.currentTimeMillis();
                                _updConcatOffset = 0;

                                byte[] _temp_buffer = new byte[512*1024];
                                byte[] _concat_buffer = new byte[4*1024];

                                DatagramPacket _gPacket =  new DatagramPacket(_temp_buffer, _temp_buffer.length);
                                Tools.Log.Print(Tools.LogLevel.ll_Info, "Local UDP Host : "+Tools.GetLocalIPv4Address()+" / "+_UDPSocket.getLocalPort());

                                while(_UDPSocket != null || _TCPSocket != null)
                                {
                                    long _t = System.currentTimeMillis();

                                    if(_UDPSocket != null) UDPProcessor(_gPacket, _temp_buffer, _concat_buffer, _t);
                                    if(_TCPSocket != null) TCPProcessor(_temp_buffer, _t);
                                }
                            }
                            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
                        }
                    }).start();
                }

                return true;
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
            return false;
        }
        //-------------------------------------------------------------------------------------
        private void SendUDPData(DatagramSocket _socket, String _ip, int _port, byte[] _data, int _len)
        {
            try
            {
                if(_UDPSocket == null) return;
                if(_socket == null || _ip.length() < 0 || _port <= 0) return;

                DatagramPacket _packet = new DatagramPacket(_data, _len > 0 ? _len : _data.length, InetAddress.getByName(_ip), _port);
                if(_socket != null) _socket.send(_packet);
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        }
        //-------------------------------------------------------------------------------------
        public void SendUDPData(byte[] _data, int _len)
        {
            try
            {
                SendUDPData(_UDPSocket, _RemoteIP_UDP, _RemotePort_UDP, _data, _len);
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        }
        //---------------------------------------------------------------------------------------
        public boolean cmdReceived(int _id, byte[] _data, int _len)
        {
            try
            {
                if(_UDPSocket == null) return false;
                if(_data[0] != '{')  return false;

                String _cmd_string = new String(_data, 0, _len);
                if(_cmd_string.contains(":") == false) return false;

                JSONObject _cmd = new JSONObject(_cmd_string);
                switch(_cmd.optString("cmd", ""))
                {
                    case "dev_pair":
                        String _mode = _cmd.optString("mode", "udp");
                        if(_mode.equals("udp"))
                        {
                            _RemoteIP_UDP = _cmd.optBoolean("is_lan", false) ? _cmd.getString("ip_lan") : _cmd.getString("ip");
                            _RemotePort_UDP = _cmd.getInt("port");

                            SendUDPData(_UDPSocket, _RemoteIP_UDP, _RemotePort_UDP, "hello".getBytes("UTF-8"), 0); Thread.sleep(100);
                            SendUDPData(_UDPSocket, _RemoteIP_UDP, _RemotePort_UDP, "hello".getBytes("UTF-8"), 0); Thread.sleep(100);
                            SendUDPData(_UDPSocket, _RemoteIP_UDP, _RemotePort_UDP, "hello".getBytes("UTF-8"), 0); Thread.sleep(100);
                        }
                        else
                        if(_mode.equals("tcp"))
                        {
                            try
                            {
                                _UDPSocket.disconnect();
                                _UDPSocket.close();
                            }
                            catch(Exception _e){}

                            _cmd.put("target", _cmd.optString("uuid", ""));

                            _UDPSocket = null;
                            _TCPSocket = new Socket();
                            _TCPSocket.connect(new InetSocketAddress(_cmd.optString("tcp_server_ip", _ServerName), _cmd.optInt("tcp_server_port", 31002)));
                            _TCPSocket.setTcpNoDelay(true);
                            _TCPSocket.getOutputStream().write(_cmd.toString().getBytes());
                        }
                        break;
                }

                Tools.Log.Print(Tools.LogLevel.ll_Info, "cmdReceived : "+_cmd_string);
                return true;
            }
            catch(Exception _e){};
            return false;
        }
        //---------------------------------------------------------------------------------------
        public boolean Stop()
        {
            try
            {
                if(_TCPSocket != null)
                {
                    try
                    {
                        _TCPSocket.close();
                    }
                    catch(Exception _e){}
                    _TCPSocket = null;
                }

                if(_UDPSocket != null)
                {
                    try
                    {
                        SendUDPData(_UDPSocket, _RemoteIP_UDP, _RemotePort_UDP, "bye".getBytes("UTF-8"), 0); Thread.sleep(100);
                        SendUDPData(_UDPSocket, _RemoteIP_UDP, _RemotePort_UDP, "bye".getBytes("UTF-8"), 0); Thread.sleep(100);
                        SendUDPData(_UDPSocket, _RemoteIP_UDP, _RemotePort_UDP, "bye".getBytes("UTF-8"), 0); Thread.sleep(100);

                        Tools.Log.Print(Tools.LogLevel.ll_Info, "Sending P2P ByeBye");
                    }
                    catch(Exception _e){}

                    try
                    {
                        if(_UDPSocket.getPort() > 0) { _UDPSocket.disconnect(); Tools.Log.Print(Tools.LogLevel.ll_Info, "P2P Socket disconnect"); }
                    }
                    catch(Exception _e){}

                    try
                    {
                        if(_UDPSocket.isClosed()) { _UDPSocket.close(); Tools.Log.Print(Tools.LogLevel.ll_Info, "P2P Socket close"); }
                    }
                    catch(Exception _e){}

                    _UDPSocket = null;
                    return true;
                }
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
            return false;
        }
        //---------------------------------------------------------------------------------------
        public void dataReceived(int _id, byte[] _data, int _len)
        {
            try
            {
                if(_Parent != null) _Parent.OnStreamData(_id, AVCodec.VCodecID_FLAG, _data, _len, 0);
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        }
        //---------------------------------------------------------------------------------------
        public long RemainBufferSize()
        {
            try
            {
                if(_UDPSocket != null) return _UDPSocket.getReceiveBufferSize();
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
            return 0;
        }
        //---------------------------------------------------------------------------------------
    }
    //---------------------------------------------------------------------------------------
    //
    //---------------------------------------------------------------------------------------
    private static String _MarsCloudHost = "test.mars-cloud.com";
    public static String _MarsCloudURL = "https://"+_MarsCloudHost;
    public static String _MarsCloudMQTTURL = "ssl://"+_MarsCloudHost+":8883";
    public static int _DefaultConnectTimeOut = 10000;
    public static String _DefaultCharset = "UTF-8";
    //---------------------------------------------------------------------------------------
    public MarsCloudClient _thisIterator = null;
    public P2PStreamer _P2PStreamer = null;
    public MqttClient _MQTTClient = null;
    //---------------------------------------------------------------------------------------
    private AbstractObject.IVideoDataCallback _VideoCallback = null;
    private ClientMQTTCallback _MQTTCallback = null;
    private ArrayList<AbstractObject.IMessageCallback> _MessageCallback = new ArrayList<AbstractObject.IMessageCallback>();
    //---------------------------------------------------------------------------------------
    public AbstractObject.IDemuxer _VideoDemuxer = null;
    public AbstractObject.IDemuxer _AudioDemuxer = null;
    //---------------------------------------------------------------------------------------
    private String _AuthToken = "";
    private String _Account = "";
    private String _Password = "";
    //---------------------------------------------------------------------------------------
    public MarsCloudClient(String _account, String _password)
    {
        try
        {
            _thisIterator = this;
            _Account = _account;
            _Password = _password;
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
    public String HttpGet(final String _cmd, final AbstractObject.IStringDataCallback _callback)
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
    public String HttpGet(final String _cmd, final int _timeout , final AbstractObject.IStringDataCallback _callback)
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
    public String HttpGet(final String _cmd, String _auth)
    {
        try
        {
            return Tools.HttpGet(_MarsCloudURL+_cmd, _auth, _DefaultConnectTimeOut);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public String HttpPost(final String _cmd, final String _payload, final AbstractObject.IStringDataCallback _callback)
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
    public boolean StartStream(String _srcID)
    {
        try
        {
            return StartStream(_srcID, _MarsCloudHost);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //---------------------------------------------------------------------------------------
    public boolean StartStream(String _srcID, String _stream_host)
    {
        try
        {
            if(_stream_host.startsWith("tcp") || _stream_host.startsWith("http"))
            {
                if(_VideoDemuxer == null) _VideoDemuxer = new FFMpegDemuxer(0, _stream_host);
            }
            else
            {
                if(_VideoDemuxer == null) _VideoDemuxer = new FFMpegDemuxer(0, 64*1024);
                if(_P2PStreamer == null) _P2PStreamer = new P2PStreamer(this, _stream_host != null ? _stream_host : _MarsCloudHost);
            }

            if(_VideoDemuxer != null) _VideoDemuxer.BindVideoCallback(_VideoCallback);
            if(_P2PStreamer != null) _P2PStreamer.Start(_srcID);

            Tools.Log.Print(Tools.LogLevel.ll_Info, "StartStream Success");
            return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        Tools.Log.Print(Tools.LogLevel.ll_Warning, "StartStream Fail");
        return false;
    }
    //---------------------------------------------------------------------------------------
    public boolean StopStream()
    {
        try
        {
            Tools.Log.Print(Tools.LogLevel.ll_Info, "Stop VideoDemuxer ...");
            if(_VideoDemuxer != null) _VideoDemuxer.Release();

            Tools.Log.Print(Tools.LogLevel.ll_Info, "Stop P2PStreamer ...");
            if(_P2PStreamer != null) _P2PStreamer.Stop();

            _VideoDemuxer = null;
            _P2PStreamer = null;

            Tools.Log.Print(Tools.LogLevel.ll_Info, "StopStream Success");
            return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        Tools.Log.Print(Tools.LogLevel.ll_Warning, "StopStream Fail");
        return false;
    }
    //---------------------------------------------------------------------------------------
    public boolean BindVideoCallback(AbstractObject.IVideoDataCallback _callback)
    {
        try
        {
            _VideoCallback = _callback;
            if(_VideoDemuxer != null) _VideoDemuxer.BindVideoCallback(_VideoCallback);

            Tools.Log.Print(Tools.LogLevel.ll_Info, _VideoCallback != null ? "Bind VideoCallback Success" : "Unbind VideoCallback Success");
            return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //---------------------------------------------------------------------------------------
    public boolean BindAudioCallback(AbstractObject.IAudioDataCallback _callback)
    {
        try
        {
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //---------------------------------------------------------------------------------------
    public boolean AddMessageCallback(AbstractObject.IMessageCallback _callback)
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
    public int GetMessageLength()
    {
        try
        {
            if(_MessageCallback != null) return _MessageCallback.size();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return 0;
    }
    //---------------------------------------------------------------------------------------
    public boolean MQTTReconnect()
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

                _MQTTClient = new MqttClient(_MarsCloudMQTTURL, _AuthToken + '@'+new Date().getTime()%1000, new MemoryPersistence());
                _MQTTClient.setCallback(_MQTTCallback);
                _MQTTClient.connect(_options);
                _MQTTClient.subscribe(_Account + "/+/#");

                Tools.Log.Print(Tools.LogLevel.ll_Info, "MQTT connect success ~");

                return true;
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        Tools.Log.Print(Tools.LogLevel.ll_Warning, "MQTT connect Fail ~");
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
    //--------------------------------------------------------------------------------------
    public void UpdateDevice(JSONObject _payload)
    {
        try
        {
            if(_payload != null)
            {
                Tools.HttpPost(_MarsCloudURL+"/api/usrinfo?method=updatedatasrc", _AuthToken, _payload.toString(), _DefaultConnectTimeOut);
                Tools.Log.Print(Tools.LogLevel.ll_Debug, "UpdateDevice : "+_payload.toString());
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //--------------------------------------------------------------------------------------
    public boolean CertificatDevice(String _uuid, String _suid, String _stamp)
    {
        try
        {
            JSONObject _payload = new JSONObject();

            _payload.put("uuid", _uuid);
            _payload.put("suid", _suid);
            _payload.put("stamp", _stamp);
            _payload.put("req", "auth_token");

            String _res = Tools.HttpPost(_MarsCloudURL+"/auth/registry?target=device", _AuthToken, _payload.toString(), _DefaultConnectTimeOut);

            if(_res != null)
            {
                Tools.Log.Print(Tools.LogLevel.ll_Info, "CertificatDevice Success : "+_uuid);
                return true;
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        Tools.Log.Print(Tools.LogLevel.ll_Info, "CertificatDevice Fail : "+_uuid);
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
    public String PutData(String _uuid , String _suid, String _ukey ,JSONArray _values)
    {
        try
        {
            JSONObject _obj = new JSONObject();
            _obj.put("uuid"	, _uuid);
            _obj.put("values"	, _values);

            if(_suid != null && _suid.length() > 0)	_obj.put("suid"	, _suid);
            if(_ukey != null && _ukey.length() > 0)	_obj.put("ukey", _ukey);

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
    public void CloseCallback()
    {
        try
        {
            //_callback= null;
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
    public void Start(final AbstractObject.IStringDataCallback _callback)
    {
        try
        {
            Tools.Log.Print(Tools.LogLevel.ll_Info, "Try to Login : "+_MarsCloudURL);
            HttpGet("/auth/login?usr="+_Account+"&pwd="+_Password , new AbstractObject.IStringDataCallback()
            {
                public void OnData(String _payload)
                {
                    try
                    {
                        _AuthToken = _payload;
                        MQTTReconnect();

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
            _VideoCallback = null;
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
    @Override
    public void CloseTCPService() {

    }

    //---------------------------------------------------------------------------------------
    private void OnStreamData(int _id, int _type, Object _data, int _size, long _pts)
    {
        try
        {
            Tools.Log.Print(Tools.LogLevel.ll_Debug, "OnStreamData : "+_size);

            if(_VideoCallback != null && _size > 0) _VideoCallback.OnRawData(_id, _type, _data, _size, _pts);
            if(_VideoDemuxer != null && _type == AVCodec.VCodecID_FLAG) _VideoDemuxer.PushRawData(_type, (byte[] )_data, _size, _pts);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
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
    public boolean BindOAuth_Line(boolean _is_binding, String _id, String _token)
    {
        try
        {
            String _method = _is_binding ? "bind" : "unbind";
            JSONObject _payload = new JSONObject();

            _payload.put("line_oauth_id", _id);
            _payload.put("line_oauth_token", _token);

            String _resp = HttpPost("/auth/oauth_bind?method="+_method, _payload.toString(), null);
            if(_resp != null && _resp.length() >= 0) return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
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
}
