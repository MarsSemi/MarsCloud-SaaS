using Newtonsoft.Json.Linq;
using System;
using System.Net;
using System.Text;
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
public class MarsClient
{
    //-----------------------------------------------------------------------
    public String _Host = "";
    public String _Account = "";
    public String _Password = "";
    public String _Token = "";

    public WebClient _Http = new WebClient();
    //-----------------------------------------------------------------------
    public MarsClient()
    {
        try
        {
            _Http.Encoding = Encoding.UTF8;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
    }
    //-----------------------------------------------------------------------
    private String HttpGet(String _url)
    {
        try
        {
            _Http.Headers.Clear();

            if (_Token != null && _Token.Length > 64) _Http.Headers.Add("authentication", "Bearer " + _Token);

            //Console.WriteLine(_url);

            return _Http.DownloadString(_url);
        }
        catch (Exception _e) {}
        return null;
    }
    //-----------------------------------------------------------------------
    private String HttpPost(String _url, String _payload)
    {
        try
        {
            _Http.Headers.Clear();

            if (_Token != null && _Token.Length > 64) _Http.Headers.Add("authentication", "Bearer " + _Token);

            return _Http.UploadString(_url, _payload);
        }
        catch (Exception _e) {}
        return null;
    }
    //-----------------------------------------------------------------------
    public bool IsLogin()
    {
        try
        {
            if (_Token != null && _Token.Length > 64) return true;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return false;
    }
    //-----------------------------------------------------------------------
    public bool Login(string _host, string _account, string _pwd, string _proj)
    {
        try
        {
            _Host = _host;
            _Account = _account;
            _Password = _pwd;

            String _resp = HttpGet(_Host + "/auth/login?usr=" + _Account + "&pwd=" + _Password + "&proj=" + _proj);
            if (_resp != null)
            {
                _Token = _resp;
                return true;
            }
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return false;
    }
    //-----------------------------------------------------------------------
    public bool PutMessage(String _topic, String _msg)
    {
        try
        {
            String _resp = HttpPost(_Host + "/api/put?message&topic="+ _topic, _msg);
            if (_resp != null)
                return true;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return false;
    }
    //-----------------------------------------------------------------------
    public bool RegDevice(String _uuid, String _suid, String _profile, String _name, String _vender)
    {
        try
        {
            JObject _jsonObj = new JObject();

            _jsonObj.Add("uuid", _uuid);
            _jsonObj.Add("suid", _suid);
            _jsonObj.Add("data_profile", _profile);
            _jsonObj.Add("name", _name);
            _jsonObj.Add("vender", _vender);

            String _resp = HttpPost(_Host + "/api/usrinfo?method=adddatasrc", _jsonObj.ToString());

            if (_resp.ToLower() == "ok") return true;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return false;
    }
    //-----------------------------------------------------------------------
    public bool PutData(String _uuid, String _suid, JObject _value)
    {
        try
        {
            JObject _jsonObj = new JObject();
            JArray _values = new JArray();

            _values.Add(_value);

            _jsonObj.Add("uuid", _uuid);
            _jsonObj.Add("suid", _suid);
            _jsonObj.Add("values", _values);

            String _resp = HttpPost(_Host + "/api/put?data", _jsonObj.ToString());
            if (_resp != null)
                return true;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return false;
    }
    //-----------------------------------------------------------------------
    public bool PutData(String _uuid, String _suid, JArray _values)
    {
        try
        {
            JObject _jsonObj = new JObject();

            _jsonObj.Add("uuid", _uuid);
            _jsonObj.Add("suid", _suid);
            _jsonObj.Add("values", _values);

            String _resp = HttpPost(_Host + "/api/put?data", _jsonObj.ToString());
            if (_resp != null)
                return true;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return false;
    }
    //-----------------------------------------------------------------------
    public string GetLastData(String _uuid, String _suid, int _Count)
    {
        try
        {
            JObject _jsonObj = new JObject();
            _jsonObj.Add("uuid", _uuid);
            _jsonObj.Add("suid", _suid);
            _jsonObj.Add("count", _Count);

            String _resp = HttpPost(_Host + "/api/lastdata?method=read", _jsonObj.ToString());
            if (_resp != null)
                return _resp;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return null;
    }
    //-----------------------------------------------------------------------
    public string GetData(String _uuid, String _suid, long _StartTime, long _EndTime)
    {
        try
        {
            JObject _jsonObj = new JObject();
            _jsonObj.Add("uuid", _uuid);
            _jsonObj.Add("suid", _suid);
            _jsonObj.Add("timestamp", _StartTime + "~" + _EndTime);

            String _resp = HttpPost(_Host + "/api/get?data", _jsonObj.ToString());
            if (_resp != null)
                return _resp;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return null;
    }
    //-----------------------------------------------------------------------
    public Boolean RemoveData(String _uuid, String _suid, String _ukey)
    {
        try
        {
            JObject _jsonObj = new JObject();

            _jsonObj.Add("uuid", _uuid);
            _jsonObj.Add("suid", _suid);
            _jsonObj.Add("ukey", _ukey);

            String _resp = HttpPost(_Host + "/api/del?data", _jsonObj.ToString());
            if (_resp != null)
                return true;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return false;
    }
    //-----------------------------------------------------------------------
    public string CreateDataSource(String _uuid, String _suid, String _name)
    {
        try
        {
            JObject _jsonObj = new JObject();
            _jsonObj.Add("uuid", _uuid);
            _jsonObj.Add("suid", _suid);
            _jsonObj.Add("name", _name);

            String _resp = HttpPost(_Host + "/api/usrinfo?method=adddatasrc", _jsonObj.ToString());
            if (_resp != null)
                return _resp;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return null;
    }
    //-----------------------------------------------------------------------
    public string DelDataSource(String _uuid, String _suid)
    {
        try
        {
            JObject _jsonObj = new JObject();
            _jsonObj.Add("uuid", _uuid);
            _jsonObj.Add("suid", _suid);

            String _resp = HttpPost(_Host + "/api/usrinfo?method=deldatasrc", _jsonObj.ToString());
            if (_resp != null)
                return _resp;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return null;
    }
    //-----------------------------------------------------------------------
    public string CallService(String _service, String _api, String _payload)
    {
        try
        {
            String _resp = _payload == null ? HttpGet(_Host + "/services/" + _service + _api) : HttpPost(_Host + "/services/"+ _service+ _api, _payload);
            if (_resp != null)
                return _resp;
        }
        catch (Exception _e) { Console.WriteLine(_e.ToString()); }
        return null;
    }
    //-----------------------------------------------------------------------
}
//-----------------------------------------------------------------------
