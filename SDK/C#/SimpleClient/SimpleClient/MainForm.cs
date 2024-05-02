using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Drawing;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Forms;
//---------------------------------------------------------------
//
//---------------------------------------------------------------
namespace SimpleClient
{
    public partial class MainForm : Form
    {
        //---------------------------------------------------------------
        //
        //---------------------------------------------------------------
        private MarsClient _Client = null;
        private MarsMQTT _Mqtt = null;
        //---------------------------------------------------------------
        public MainForm()
        {
            try
            {
                InitializeComponent();

                _Client = new MarsClient();
                _Mqtt = new MarsMQTT(_Client);
            }
            catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        }
        //---------------------------------------------------------------
        private void MainForm_Leave(object sender, EventArgs e)
        {
            try
            {
                
            }
            catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        }
        //---------------------------------------------------------------
        private void MainForm_FormClosed(object sender, FormClosedEventArgs e)
        {
            try
            {
                Console.WriteLine("Exit APP");
                
                _Mqtt.Disconnect();

                Application.ExitThread();
                Application.Exit();
            }
            catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        }
        //---------------------------------------------------------------
        private void btn_Connect_Click(object sender, EventArgs e)
        {
            try
            {
                if(_Client.Login("https://test.mars-cloud.com", "test", "test", "justtest"))
                {
                    Console.WriteLine("Login Success");

                    if(_Mqtt.Connect())
                    {
                        Console.WriteLine("MQTT Client connect success");

                        _Mqtt.MsgReceiver += MQTTMsgHandler;
                        _Mqtt.Subscribe("test/+/#");

                        btn_Test.Enabled = true;
                    }
                    else
                        Console.WriteLine("MQTT Client connect fail");
                }
            }
            catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        }
        //---------------------------------------------------------------
        private void MQTTMsgHandler(object sender, String _topic, String _msg)
        {
            try
            {
                Console.WriteLine("-------------- Get MQTT Msg --------------");
                Console.WriteLine(_topic);
                Console.WriteLine(_msg);
                Console.WriteLine("------------------------------------------");
            }
            catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        }
        //---------------------------------------------------------------
        private void btn_Test_Click(object sender, EventArgs e)
        {
            try
            {
                String _resp;
                JObject _data = new JObject();

                _data.Add("temp", 24.7);
                _data.Add("humi", 80);

                if(_Client.RegDevice("dev", "test", "both.meter", "Virtual Meter", "com.test")) Console.WriteLine("RegDevice Success");
                if (_Client.PutMessage("test/my/msg", _data.ToString())) Console.WriteLine("PutMessage Success");
                if (_Client.PutData("dev", "test", _data)) Console.WriteLine("PutData Success");

                _resp = _Client.GetLastData("dev", "test", 1);

                if(_resp != null)
                {
                    _data = JObject.Parse(_resp);
                    _data = (JObject )_data["results"][0];

                    Console.WriteLine(_data.ToString());

                    if (_Client.RemoveData("dev", "test", _data["ukey"].ToString())) Console.WriteLine("RemoveData Success");
                }

                if (_Client.CallService("service.myservice", "/api/hello", null) != null) Console.WriteLine("CallService Success");
            }
            catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        }
        //---------------------------------------------------------------
    }
}
