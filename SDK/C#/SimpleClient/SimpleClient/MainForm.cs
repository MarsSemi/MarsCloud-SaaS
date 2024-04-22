using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Data;
using System.Drawing;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Forms;

namespace SimpleClient
{
    public partial class MainForm : Form
    {
        //---------------------------------------------------------------
        //
        //---------------------------------------------------------------
        private MarsClient _Client = new MarsClient();
        //---------------------------------------------------------------
        public MainForm()
        {
            InitializeComponent();
        }
        //---------------------------------------------------------------
        private void btn_Connect_Click(object sender, EventArgs e)
        {
            try
            {
                if(_Client.Login("https://test.mars-cloud.com", "test", "test", "justtest"))
                {
                    Console.WriteLine("Login Success");
                    btn_Test.Enabled = true;
                }
               
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
                if(_Client.PutData("dev", "test", _data)) Console.WriteLine("PutData Success");

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
