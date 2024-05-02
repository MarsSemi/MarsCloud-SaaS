using Newtonsoft.Json.Linq;
using System;
using System.Net;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using uPLibrary.Networking.M2Mqtt;
using uPLibrary.Networking.M2Mqtt.Messages;
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
public class MarsMQTT
{
    //-----------------------------------------------------------------------
    private MarsClient _MarsClient = null;
    private MqttClient _MQTTClient = null;
    //-----------------------------------------------------------------------
    public delegate void MessageHandler(object sender, String _topic, String _msg);
    public event MessageHandler MsgReceiver;
    //-----------------------------------------------------------------------
    public MarsMQTT(MarsClient _client)
    {
        try
        {
            _MarsClient = _client;
            _MQTTClient = null;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
    }
    //-----------------------------------------------------------------------
    public void Disconnect()
    {
        try
        {
            _MQTTClient.Disconnect();
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
    }
    //-----------------------------------------------------------------------
    public bool Connect()
    {
        try
        {
            if(_MarsClient != null)
            {
                String _host = _MarsClient._Host;
                String _user = _MarsClient._Account;
                String _id = _MarsClient._Account + "@" + new Random().Next() % 10000;
                String _token = _MarsClient._Token;

                if (_host.LastIndexOf(":") > 8)
                    _host = _host.Substring(0, _host.LastIndexOf(":"));
                
                if (_host.StartsWith("http://")) _host = _host.Replace("http://", "");
                if (_host.StartsWith("https://")) _host = _host.Replace("https://", "");
                
                _MQTTClient = new MqttClient(_host, 8883, true, MqttSslProtocols.TLSv1_2, null, null);
                _MQTTClient.ProtocolVersion = MqttProtocolVersion.Version_3_1_1;
                _MQTTClient.MqttMsgPublishReceived += MsgReceived;
                _MQTTClient.Connect(_id, _user, _token);

                if (_MQTTClient.IsConnected)
                    return true;
            }
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return false;
    }
    //-----------------------------------------------------------------------
    public bool Subscribe(String _topic)
    {
        try
        {
            if (_MQTTClient != null && _MQTTClient.IsConnected)
                if (_MQTTClient.Subscribe(new String[] { _topic }, new byte[] { MqttMsgBase.QOS_LEVEL_AT_LEAST_ONCE }) == 0)
                    return true;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return false;
    }
    //-----------------------------------------------------------------------
    void MsgReceived(object _sender, MqttMsgPublishEventArgs _event)
    {
        try
        {
            MsgReceiver(this, _event.Topic, System.Text.Encoding.UTF8.GetString(_event.Message));
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
    }
    //-----------------------------------------------------------------------
}
//-----------------------------------------------------------------------
