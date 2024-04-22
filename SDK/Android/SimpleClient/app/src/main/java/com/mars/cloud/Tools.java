package com.mars.cloud;

import android.app.Activity;
import android.app.ActivityManager;
import android.app.AlertDialog;
import android.app.ProgressDialog;
import android.content.Context;
import android.content.DialogInterface;
import android.graphics.Bitmap;
import android.graphics.BitmapFactory;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import android.net.wifi.ScanResult;
import android.net.wifi.SupplicantState;
import android.net.wifi.WifiInfo;
import android.net.wifi.WifiManager;
import android.os.Build;
import android.os.Looper;
import android.os.StrictMode;
import android.provider.Settings;
import android.util.Base64;
import android.view.View;
import android.view.inputmethod.InputMethodManager;
import android.widget.Toast;

import org.apache.commons.httpclient.HttpClient;
import org.apache.commons.httpclient.HttpStatus;
import org.apache.commons.httpclient.methods.ByteArrayRequestEntity;
import org.apache.commons.httpclient.methods.GetMethod;
import org.apache.commons.httpclient.methods.PostMethod;
import org.json.JSONArray;
import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.PrintWriter;
import java.io.StringWriter;
import java.lang.Thread.UncaughtExceptionHandler;
import java.net.InetAddress;
import java.net.NetworkInterface;
import java.nio.ByteBuffer;
import java.sql.Date;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Enumeration;
import java.util.List;
import java.util.Locale;

//----------------------------------------------------------------------------------------
//
//----------------------------------------------------------------------------------------
public class Tools
{
    //-------------------------------------------------------------------------------------
    private final static Tools _thisIterator = new Tools();
    private final static String _DateFormatString = "MM/dd HH:mm:ss";
    public final static String _DefaultCharset = "UTF-8";
    private final static android.text.format.DateFormat _DateFormat = new android.text.format.DateFormat();
    //-------------------------------------------------------------------------------------
    //public final static long _TimeZoneOffset=new Date().getTimezoneOffset()*60*1000;
    //-------------------------------------------------------------------------------------
    private Tools()
    {
        try
        {
            StrictMode.setThreadPolicy(new StrictMode.ThreadPolicy.Builder().permitNetwork().build());
        }
        catch(Exception _e){ if(_e != null) _e.printStackTrace();}
    }
    //-------------------------------------------------------------------------------------
    //
    //-------------------------------------------------------------------------------------
    private static ProgressDialog _ProgressDialog = null;
    //-------------------------------------------------------------------------------------
    public static final LogOut Log=_thisIterator.new LogOut();
    //-------------------------------------------------------------------------------------
    public class LogLevel
    {
        public final static int ll_Debug = 0;
        public final static int ll_Normal = 1;
        public final static int ll_Info = 2;
        public final static int ll_Warning = 3;
        public final static int ll_Error = 4;
        public final static int ll_Serious = 5;
    }
    //-------------------------------------------------------------------------------------
    public static String GetCurrentDateTimeString()
    {
        try
        {
            return _DateFormat.format(_DateFormatString, new java.util.Date()).toString();
        }
        catch(Exception _e){ if(_e != null) _e.printStackTrace();}
        return "";
    }
    //-------------------------------------------------------------------------------------
    public static String GetFunctionName(Class<? extends Object> _class)
    {
        try
        {
            if(_class != null)
            {
                String _class_name = (_class!= null && _class.getEnclosingClass() != null) ? _class.getEnclosingClass().getName() : "<???>";
                String _method_name = (_class != null && _class.getEnclosingMethod() != null) ? _class.getEnclosingMethod().getName() : null;

                if(_method_name == null) _method_name = Thread.currentThread().getStackTrace()[1].getMethodName();
                return (_class_name + "." + _method_name);
            }
        }
        catch(Exception _e){ if(_e != null) _e.printStackTrace(); }
        return "Unknown Function";
    }
    //-------------------------------------------------------------------------------------
    public static String ExceptionToString(Exception _e)
    {
        try
        {
            if(_e != null)
            {
                StringWriter _sw = new StringWriter();
                _e.printStackTrace(new PrintWriter(_sw));
                return _sw.toString();
            }
        }
        catch(Exception _local_e){ _local_e.printStackTrace(); }
        return "ExceptionToString FAIL";
    }
    //-------------------------------------------------------------------------------------
    static public String ExceptionMsgPrintOut(Exception _e, Class<? extends Object> _class)
    {
        String _msg = "Unknow exception";

        try
        {
            String _class_name = (_class!= null && _class.getEnclosingClass() != null) ? _class.getEnclosingClass().getName() : "<???>";
            String _method_name = (_class != null && _class.getEnclosingMethod() != null) ? _class.getEnclosingMethod().getName() : "<???>";
            String full_name = _class_name + "." + _method_name;
            boolean _isUnknownSrc = full_name.contains("???");

            _msg =  full_name+" --> "+_msg;
            if(_e != null)
            {
                int _line = -1;
                if (_e.getStackTrace().length > 0 && _isUnknownSrc == false)
                {
                    StackTraceElement[] _trackers = _e.getStackTrace();
                    for (int i = 0; i < _trackers.length; i++)
                        if (_trackers[i].getClassName().contains(_class_name) && _trackers[i].getMethodName().contains(_method_name))
                        {
                            _line = _trackers[i].getLineNumber();
                            _msg = _trackers[i].toString() + " -> " + _e.getMessage();
                            break;
                        }
                }

                if (_isUnknownSrc || _line < 0) _msg = ExceptionToString(_e);
                _msg = String.format("%s", _msg);
            }
        }
        catch(Exception _e_local)
        {
            if(_e != null && _e.getStackTrace().length > 0)
                _msg = ExceptionToString(_e);
            else
                _e_local.printStackTrace();
        }
        finally
        {
            Log.Print(LogLevel.ll_Error, _msg);
            return _msg;
        }
    }
    //-------------------------------------------------------------------------------------
    static public String ExceptionMsgPrintOut(Context _parent, Exception _e, Class<? extends Object> _class)
    {
        String _msg = ExceptionMsgPrintOut(_e, _class);
        try
        {
            if(_parent != null) Toast.makeText(_parent, _msg, Toast.LENGTH_LONG);
        }
        catch(Exception _e_local){};
        return _msg;
    }
    //-------------------------------------------------------------------------------------
    private static class SpecailExceptionHandler implements UncaughtExceptionHandler
    {
        private Activity _Activity = null;
        public SpecailExceptionHandler(Activity _parent){ _Activity = _parent; }
        public void uncaughtException(Thread _t, Throwable _e)
        {
            try
            {
                String _text = "";
                if (_e.getStackTrace().length > 0 && _Activity != null)
                {
                    String _pktName = _Activity.getPackageName();

                    StackTraceElement[] _trackers = _e.getStackTrace();
                    for (int i = 0; i < _trackers.length ; i++)
                        if (_trackers[i].getClassName().contains(_pktName))
                            _text = _text + _trackers[i].toString()+"\n";
                }

                String _msg = String.format("Unexpected Error ---> %s", _text.length() > 0 ? _text : _e.toString());
                _thisIterator.Log.Print(LogLevel.ll_Serious, _msg);

                if(_msg.contains("StackOverflowError")) return;
                if(_Activity != null)
                {
                    _msg = _e.toString();

                    Toast.makeText(_Activity, _msg, Toast.LENGTH_SHORT).show();
                    new AlertDialog.Builder(_Activity)
                            .setTitle("Unexpected Error")
                            .setMessage(_msg)
                            .show();
                }
            }
            catch(Exception _e_local){}
        }
    }
    //-------------------------------------------------------------------------------------
    static public void EnableUncaughtExceptionHandler(Activity _activity)
    {
        try
        {
            Thread.setDefaultUncaughtExceptionHandler (new SpecailExceptionHandler(_activity));
        }
        catch(Exception _e_local){};
    }
    //-------------------------------------------------------------------------------------
    public class LogOut
    {
        //-------------------------------------------------------------------------------------
        private final String _logSender = "MetaLog";
        //-------------------------------------------------------------------------------------
        private int _Log_ShowLevel = LogLevel.ll_Debug;

        private String _RemoteURL = "";
        private String _AuthToken = "";
        private String _UDID = "";

        private ArrayList<String> _RemoteLogList = new ArrayList<String>();
        private boolean _UsingRemoteDebugger = false;
        private long _PrevDeliverLogTime = 0;
        //-------------------------------------------------------------------------------------
        public LogOut()
        {

        }
        //-------------------------------------------------------------------------------------
        private void SendLogToRemote(final JSONArray _list)
        {
            try
            {
                new Thread(new Runnable()
                {
                    public void run()
                    {
                        try
                        {
                            if(_UsingRemoteDebugger == false) return;

                            JSONObject _payload = new JSONObject();
                            _payload.put("logs", _list);
                            _thisIterator.HttpPost(_RemoteURL, _AuthToken, _payload.toString(), 3000);
                        }
                        catch(Exception _e){ if(_e != null) _e.printStackTrace();}
                    }
                }).start();
            }
            catch(Exception _e){ if(_e != null) _e.printStackTrace();}
        }
        //-------------------------------------------------------------------------------------
        public void SetRemoteDebugger(String _url, String _token, String _udid)
        {
            try
            {
                _RemoteURL = _url == null ? "" : _url;
                _AuthToken = _token == null ? "" : _token;
                _UDID = _udid == null ? "" : _udid;
                _UsingRemoteDebugger = (_RemoteURL.length() > 0 && _AuthToken.length() > 0 && _UDID.length() > 0);

                if(_UsingRemoteDebugger)
                {
                    _RemoteURL = _RemoteURL + "/api/put?log&topic=log.app." + _udid;
                    Print(LogLevel.ll_Warning, "Remote Debugger is enable : "+ _RemoteURL);
                }
            }
            catch(Exception _e){ if(_e != null) _e.printStackTrace();}
        }
        //-------------------------------------------------------------------------------------
        public void SetLevel(int _level)
        {
            try
            {
                Print(LogLevel.ll_Normal, "Set Log Level : "+ _level);
                _Log_ShowLevel=_level;
            }
            catch(Exception _e){ if(_e != null) _e.printStackTrace();}
        }
        //-------------------------------------------------------------------------------------
        public void Print(int _level, String _msg)
        {
            try
            {
                if(_level < _Log_ShowLevel) return;

                String _out="["+GetCurrentDateTimeString()+"]";
                switch(_level)
                {
                    case LogLevel.ll_Normal:
                        _out = _out+"[Norm] "+_msg;
                        android.util.Log.i(_logSender, _out);
                        break;
                    case LogLevel.ll_Info:
                        _out = _out+"[Info] "+_msg;
                        android.util.Log.i(_logSender, _out);
                        break;
                    case LogLevel.ll_Warning:
                        _out = _out+"[Warn] "+_msg;
                        android.util.Log.w(_logSender, _out);
                        break;
                    case LogLevel.ll_Error:
                        _out = _out+"[Error] "+_msg;
                        android.util.Log.e(_logSender, _out);
                        break;
                    case LogLevel.ll_Serious:
                        _out = _out+"[Serious] "+_msg;
                        android.util.Log.e(_logSender, _out);
                        break;
                    default:
                        _out = _out+"[Debug] "+_msg;
                        android.util.Log.d(_logSender, _out);
                        break;
                }

                //System.out.println(_out);
                if(_UsingRemoteDebugger)
                {
                    _RemoteLogList.add(_out);

                    long _t = System.currentTimeMillis();
                    if(_t - _PrevDeliverLogTime >= 1000)
                    {
                        JSONArray _logs = new JSONArray();

                        for (String _text : _RemoteLogList) _logs.put(_text);
                        SendLogToRemote(_logs);

                        _PrevDeliverLogTime = _t;
                        _RemoteLogList.clear();
                    }
                }
            }
            catch(Exception _e){ if(_e != null) _e.printStackTrace();}
        }
    }
    //-------------------------------------------------------------------------------------
    //
    //-------------------------------------------------------------------------------------
    public static void StopAllThreads(boolean _interrupted)
    {
        StopAllThreads(_interrupted, 5000);
    }
    //-------------------------------------------------------------------------------------
    public static void StopAllThreads(boolean _interrupted, int _timeOut)
    {
        try
        {
            long _t = System.currentTimeMillis();
            while(Thread.activeCount() > 1)
            {
                Thread[] threads = new Thread[Thread.activeCount()];
                Thread.enumerate(threads);
                for (Thread _thread : threads)
                {
                    try
                    {
                        if (_thread.isAlive() && Looper.getMainLooper().getThread() != _thread)
                            if (_interrupted)
                                _thread.interrupt();
                            else
                                _thread.join(100);
                    }
                    catch (Exception _e){}
                }

                if(System.currentTimeMillis() - _t > _timeOut) break;
            }
        }
        catch(Exception _e){}
    }
    //-------------------------------------------------------------------------------------
    public static void Sleep(long _ms){try{Thread.sleep(_ms);}catch(Exception _e){};}
    //-------------------------------------------------------------------------------------
    public static void SleepEx(long _ms)
    {
        try
        {
            long _t = System.currentTimeMillis();
            while(System.currentTimeMillis() - _t < _ms) Thread.sleep(1);
        }
        catch(Exception _e){};
    }
    //-------------------------------------------------------------------------------------
    public abstract static class SafeThread extends Thread implements Runnable
    {
        //-------------------------------------------------------------------------------------
        private String _Name = "MySafeThread_"+System.currentTimeMillis();
        private Activity _Parent = null;
        private boolean _IsRunning = false;
        //-------------------------------------------------------------------------------------
        public SafeThread(){}
        public SafeThread(String _name){ _Name = _name == null ? "" : _name; }
        public SafeThread(Activity _activity, Class<?> _class){ _Parent = _activity; _Name = (_activity == null ? "" : GetFunctionName(_class)); }
        public SafeThread(Activity _activity){ _Parent = _activity; _Name = (_activity == null ? "" : GetFunctionName(_activity.getClass())); }
        public SafeThread(Activity _activity, String _name){ _Parent = _activity; _Name = _name == null ? "" : _name; }
        //-------------------------------------------------------------------------------------
        public abstract void Processor();
        public boolean IsRunning(){ return _IsRunning; };
        //-------------------------------------------------------------------------------------
        public void PostProcessor(){};
        public void PostProcessorUI(){};
        public void Sleep(int _ms){ Tools.Sleep(_ms); }
        public void run()
        {
            try
            {
                _IsRunning = true;
                Processor();
            }
            catch(Exception _e)
            {
                Tools.Log.Print(LogLevel.ll_Error, "SafeThread Crash : "+ _Name);
                Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass());
            }

            try
            {
                PostProcessor();
            }
            catch(Exception _e)
            {
                Tools.Log.Print(LogLevel.ll_Error, "SafeThread Crash : "+ _Name);
                Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass());
            }

            try
            {
                if(_Parent != null)
                {
                    Runnable _task = new Runnable(){ public void run()
                    {
                        try
                        {
                            PostProcessorUI();
                        }
                        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
                    }};

                    if (Looper.getMainLooper().getThread().equals(Thread.currentThread()))
                        _task.run();
                    else
                        _Parent.runOnUiThread(_task);
                }

            }
            catch(Exception _e)
            {
                Tools.Log.Print(LogLevel.ll_Error, "SafeThread Crash : "+ _Name);
                Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass());
            }

            _IsRunning = false;
        }
    };

    //---------------------------------------------------------------------------------------
    public static void HideSoftKeyboard(View _view)
    {
        try
        {
            InputMethodManager imm = (InputMethodManager) _view.getContext().getSystemService(Activity.INPUT_METHOD_SERVICE);
            imm.hideSoftInputFromWindow(_view.getWindowToken(), 0);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //---------------------------------------------------------------------------------------
    public static NetworkInfo IsNetworkAvailable(Context _context)
    {
        try
        {
            ConnectivityManager connectivityManager = (ConnectivityManager )_context.getSystemService(Context.CONNECTIVITY_SERVICE);
            NetworkInfo[] _nis = connectivityManager.getAllNetworkInfo();
            NetworkInfo _ni = null;

            for(int i=0;i<_nis.length;i++)
                if (_nis[i] != null && _nis[i].isConnected())
                {
                    return _nis[i];
                }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //-------------------------------------------------------------------------------------
    public static byte[] HttpGetBytes(String _url, String _auth_token, int _timeout)
    {
        try
        {
            HttpClient _client = new HttpClient();
            _client.getHttpConnectionManager().getParams().setConnectionTimeout(_timeout);
            _client.getHttpConnectionManager().getParams().setSoTimeout(_timeout);

            GetMethod _method = new GetMethod(_url);

            _method.setRequestHeader("Keep-alive", "false");
            _method.setRequestHeader("Connection", "close");
            _method.setRequestHeader("Content-Type", "application/json; charset=UTF-8");

            if(_auth_token != null && _auth_token.length() > 0) _method.setRequestHeader("Authentication", "Bearer "+_auth_token);

            byte[] _payload = null;
            switch (_client.executeMethod(_method))
            {
                case HttpStatus.SC_NO_CONTENT:
                case HttpStatus.SC_OK:
                    _payload = _method.getResponseBody();
                    break;
                default:
                    break;
            }

            _client.getHttpConnectionManager().closeIdleConnections(0);
            _method.releaseConnection();
            return _payload;
        }
        catch(Exception _e){ ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return null;
    }
    //-------------------------------------------------------------------------------------
    public static String HttpGet(String _url, String _auth_token, int _timeout)
    {
        try
        {
            byte[] _payload = HttpGetBytes(_url, _auth_token, _timeout);
            if(_payload != null) return new String(_payload, "UTF-8");
        }
        catch(Exception _e){ ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return null;
    }
    //-------------------------------------------------------------------------------------
    public static String HttpPost(String _url, String _auth_token, String _content, int _timeout)
    {
        try
        {
            HttpClient _client = new HttpClient();
            _client.getHttpConnectionManager().getParams().setConnectionTimeout(_timeout);
            _client.getHttpConnectionManager().getParams().setSoTimeout(_timeout);

            PostMethod _method = new PostMethod(_url);

            if(_auth_token != null && _auth_token.length() > 0) _method.setRequestHeader("Authentication", "Bearer "+_auth_token);

            _method.setRequestHeader("Keep-alive", "false");
            _method.setRequestHeader("Connection", "close");
            _method.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
            _method.setRequestEntity(new ByteArrayRequestEntity(_content.getBytes()));

            if(_client != null)
            {
                String _payload = null;
                int _res = _client.executeMethod(_method);
                switch (_res)
                {
                    case HttpStatus.SC_NO_CONTENT:
                    case HttpStatus.SC_OK:
                        _payload = new String(_method.getResponseBody(), "UTF-8");
                        break;
                    case HttpStatus.SC_CONFLICT:
                        _payload = "account is register";
                        break;
                    default:
                        break;
                }

                _client.getHttpConnectionManager().closeIdleConnections(0);
                _method.releaseConnection();
                return _payload;
            }
        }
        catch(Exception _e){ ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return null;
    }
    //-------------------------------------------------------------------------------------
    public static String HttpPost(String _url, String _auth_token_type, String _auth_token, String _content, int _timeout)
    {
        try
        {
            HttpClient _client = new HttpClient();
            _client.getHttpConnectionManager().getParams().setConnectionTimeout(_timeout);
            _client.getHttpConnectionManager().getParams().setSoTimeout(_timeout);

            PostMethod _method = new PostMethod(_url);

            if(_auth_token != null && _auth_token.length() > 0) _method.setRequestHeader(_auth_token_type, _auth_token);

            _method.setRequestHeader("Keep-alive", "false");
            _method.setRequestHeader("Connection", "close");

            if(_auth_token_type != null && _auth_token_type.equals("v7idea_token"))
                _method.setRequestHeader("content-type", "application/x-www-form-urlencoded");
            else
                _method.setRequestHeader("Content-Type", "application/json; charset=UTF-8");

            _method.setRequestEntity(new ByteArrayRequestEntity(_content.getBytes()));

            if(_client != null)
            {
                String _payload = "";
                int _res = _client.executeMethod(_method);
                switch (_res)
                {
                    case HttpStatus.SC_NO_CONTENT:
                    case HttpStatus.SC_OK:
                        _payload = new String(_method.getResponseBody(), "UTF-8");
                        break;
                    default:
                        break;
                }

                _client.getHttpConnectionManager().closeIdleConnections(0);
                _method.releaseConnection();
                return _payload;
            }
        }
        catch(Exception _e){ ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return null;
    }
    //-------------------------------------------------------------------------------------
    public static String HttpGet(String _url, String _auth_token_type, String _auth_token, int _timeout)
    {
        try
        {
            HttpClient _client = new HttpClient();
            _client.getHttpConnectionManager().getParams().setConnectionTimeout(_timeout);
            _client.getHttpConnectionManager().getParams().setSoTimeout(_timeout);

            GetMethod _method = new GetMethod(_url);

            if(_auth_token != null && _auth_token.length() > 0) _method.setRequestHeader(_auth_token_type, _auth_token);

            _method.setRequestHeader("Keep-alive", "false");
            _method.setRequestHeader("Connection", "close");

            if(_auth_token_type != null && _auth_token_type.equals("v7idea_token"))
                _method.setRequestHeader("content-type", "application/x-www-form-urlencoded");
            else
                _method.setRequestHeader("Content-Type", "application/json; charset=UTF-8");


            String _payload = "";
            switch (_client.executeMethod(_method))
            {
                case HttpStatus.SC_NO_CONTENT:
                case HttpStatus.SC_OK:
                    _payload = new String(_method.getResponseBody(), "UTF-8");
                    break;
                default:
                    break;
            }

            _client.getHttpConnectionManager().closeIdleConnections(0);
            _method.releaseConnection();
            return _payload;
        }
        catch(Exception _e){ ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return null;
    }
    //---------------------------------------------------------------------------------------
    public static String GetMobileDeviceUDID(Context _context)
    {
        try
        {
            if(_context == null) return "";

            String serial_number = Build.VERSION.SDK_INT >= 9 ? Build.SERIAL : "";
            String mac_address = "";
            WifiManager wifiManager = (WifiManager) _context.getSystemService(Context.WIFI_SERVICE);

            if (wifiManager != null)
            {
                WifiInfo wifiInfo = wifiManager.getConnectionInfo();
                if (wifiInfo != null) mac_address = wifiInfo.getMacAddress();
            }

            if (serial_number.length() <= 0)
            {
                int rendom_number = (int) (Math.random() * 1 + 1) + (int) (Math.random() * 10 + 1) + (int) (Math.random() * 100 + 1) + (int) (Math.random() * 1000 + 1) + (int) (Math.random() * 10000 + 1) + (int) (Math.random() * 100000 + 1) + (int) (Math.random() * 1000000 + 1) + (int) (Math.random() * 10000000 + 1);
                if (rendom_number < 10000000) rendom_number = rendom_number + 10000000;
                serial_number = String.valueOf(rendom_number);

            }
            if ((mac_address == null || mac_address.equals("")))
            {
                SimpleDateFormat formatter = new SimpleDateFormat("yyyy:MM:dd:HH:mm:ss");
                Date curDate = new Date(System.currentTimeMillis());
                mac_address = formatter.format(curDate);
            }

            mac_address = mac_address.replace(":", "");
            String[] new_mac_address = new String[2];
            new_mac_address[0] = mac_address.substring(0, 6);
            new_mac_address[1] = mac_address.substring(6);
            String[] new_serial_number = new String[2];
            new_serial_number[0] = serial_number.substring(0, 4);
            new_serial_number[1] = serial_number.substring(4);

            return "AN" + new_mac_address[0] + new_serial_number[1] + new_mac_address[1] + new_serial_number[0];
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return "";
    }
    //-------------------------------------------------------------------------------------
    public static String BytesToBase64(byte[] _data)
    {
        try
        {
            if(_data != null)
                return Base64.encodeToString(_data, Base64.DEFAULT);
        }
        catch(Exception _e){ }
        return null;
    }
    //-------------------------------------------------------------------------------------
    public static byte[] Base64ToBytes(String _data_string)
    {
        try
        {
            if(_data_string != null && _data_string.length() > 0)
                return Base64.decode(_data_string, Base64.DEFAULT);
        }
        catch(Exception _e){ ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return null;
    }
    //-------------------------------------------------------------------------------------
    public static String unicodeToUtf8(String theString) {
        char aChar;
        int len = theString.length();
        StringBuffer outBuffer = new StringBuffer(len);
        for (int x = 0; x < len;) {
            aChar = theString.charAt(x++);
            if (aChar == '\\') {
                aChar = theString.charAt(x++);
                if (aChar == 'u') {
                    // Read the xxxx
                    int value = 0;
                    for (int i = 0; i < 4; i++) {
                        aChar = theString.charAt(x++);
                        switch (aChar) {
                            case '0':
                            case '1':
                            case '2':
                            case '3':
                            case '4':
                            case '5':
                            case '6':
                            case '7':
                            case '8':
                            case '9':
                                value = (value << 4) + aChar - '0';
                                break;
                            case 'a':
                            case 'b':
                            case 'c':
                            case 'd':
                            case 'e':
                            case 'f':
                                value = (value << 4) + 10 + aChar - 'a';
                                break;
                            case 'A':
                            case 'B':
                            case 'C':
                            case 'D':
                            case 'E':
                            case 'F':
                                value = (value << 4) + 10 + aChar - 'A';
                                break;
                            default:
                                throw new IllegalArgumentException(
                                        "Malformed   \\uxxxx   encoding.");
                        }
                    }
                    outBuffer.append((char) value);
                } else {
                    if (aChar == 't')
                        aChar = '\t';
                    else if (aChar == 'r')
                        aChar = '\r';
                    else if (aChar == 'n')
                        aChar = '\n';
                    else if (aChar == 'f')
                        aChar = '\f';
                    outBuffer.append(aChar);
                }
            } else
                outBuffer.append(aChar);
        }
        return outBuffer.toString();
    }
    //---------------------------------------------------------------------------------------
    public static String ExtTrim(byte[] _buffer)
    {
        try
        {
            if (_buffer == null || _buffer.length <= 0) return "";
            if (new String(_buffer).trim().length() <= 0) return "";

            StringBuilder sBuilder = new StringBuilder();
            for (int i = 0; i < _buffer.length; i++)
            {
                if (_buffer[i] == 0x00) break;
                sBuilder.append((char) _buffer[i]);
            }

            return sBuilder.toString().trim();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return "";
    }
    //---------------------------------------------------------------------------------------
    public static String UnicodeToString(byte[] _buffer)
    {
        try
        {
            if (_buffer == null || _buffer.length <= 0) return "";
            if (new String(_buffer).trim().length() <= 0) return "";

            StringBuilder sBuilder = new StringBuilder();

            String _txt = "";
            for (int i = 0; i < _buffer.length; i++)
            {
                if (i > 2 && _buffer[i] == 0x00) break;
                _txt += (char)_buffer[i];
            }

            return _txt;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return "";
    }
    //---------------------------------------------------------------------------------------
    public static boolean ShowWaitingDialog(final Activity _parent, final String _title, final String _message)
    {
        try
        {
            return ShowWaitingDialog(_parent, _title, _message, true);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return false;
    }
    //---------------------------------------------------------------------------------------
    public static boolean ShowWaitingDialog(final Activity _parent, final String _title, final String _message, final boolean _is_auto_close)
    {
        try
        {
            if(_parent == null || _parent.getWindow() == null) return false;

            Runnable _task = new Runnable(){ public void run()
            {
                try
                {
                    if (_is_auto_close) CloseWaitingDialog();
                    if (_ProgressDialog == null)
                    {
                        _ProgressDialog = new ProgressDialog(_parent);
                        if(_ProgressDialog != null)
                        {
                            _ProgressDialog.setIndeterminate(true);
                            _ProgressDialog.setCanceledOnTouchOutside(false);
                            _ProgressDialog.setOnCancelListener(new DialogInterface.OnCancelListener()
                            {
                                public void onCancel(DialogInterface dialog)
                                {
                                    CloseWaitingDialog();
                                }
                            });

                            _ProgressDialog.setOnDismissListener(new DialogInterface.OnDismissListener()
                            {
                                public void onDismiss(DialogInterface dialog)
                                {
                                    CloseWaitingDialog();
                                }
                            });

                            try
                            {
                                _ProgressDialog.show();
                                _ProgressDialog.getWindow().getAttributes().alpha = 0.8f;
                            }
                            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
                        }
                    }

                    if (IsWaitingDialogVisible())
                    {
                        _ProgressDialog.setTitle(_title);
                        _ProgressDialog.setMessage(_message);
                    }
                }
                catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
            }};


            runOnUiThreadSafe(_parent, _task);
            return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //---------------------------------------------------------------------------------------
    public static void runOnUiThreadSafe(Activity _parent, Runnable _task)
    {
        try
        {
            if(_task == null) return;
            if(_parent == null) return;

            if (Looper.getMainLooper().getThread().equals(Thread.currentThread()))
                _task.run();
            else
                _parent.runOnUiThread(_task);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //---------------------------------------------------------------------------------------
    public static boolean IsWaitingDialogVisible()
    {
        try
        {
            return (_ProgressDialog != null && _ProgressDialog.isShowing());
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //---------------------------------------------------------------------------------------
    public static boolean CloseWaitingDialog()
    {
        try
        {
            while (_ProgressDialog != null && _ProgressDialog.isShowing()) _ProgressDialog.dismiss();
            _ProgressDialog = null;
            return true;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        finally
        {
            if(_ProgressDialog != null) _ProgressDialog.dismiss();
            _ProgressDialog = null;
        }
        return false;
    }
    //---------------------------------------------------------------------------------------
    public static Bitmap CreateBitmapARGB(byte[] _buffer, int _w, int _h)
    {
        try
        {
            if(_buffer == null || _w <= 0 || _h <= 0) return null;

            Bitmap _bitmap = Bitmap.createBitmap(_w, _h, Bitmap.Config.ARGB_8888);
            _bitmap.copyPixelsFromBuffer(ByteBuffer.wrap(_buffer, 0, _bitmap.getRowBytes() * _bitmap.getHeight()));
            return _bitmap;
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public static Bitmap DecodeImage(byte[] _buffer)
    {
        try
        {
            if(_buffer == null ) return null;
            return BitmapFactory.decodeByteArray(_buffer, 0, _buffer.length);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public static byte[] EncodePNG(Bitmap _bitmap)
    {
        try
        {
            if(_bitmap == null ) return null;
            ByteArrayOutputStream _baos = new ByteArrayOutputStream();
            _bitmap.compress(Bitmap.CompressFormat.PNG, 100, _baos);
            return _baos.toByteArray();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return null;
    }
    //---------------------------------------------------------------------------------------
    public static String GetDeviceUDID(Context _context)
    {
        try
        {
            if(_context != null)
                return Settings.Secure.getString(_context.getContentResolver(), Settings.Secure.ANDROID_ID);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return "";
    }
    //---------------------------------------------------------------------------------------
    public static String CheckPasswordValid(String Value)
    {
        try
        {
            boolean[] passwordType = new boolean[3];
            int countType = 0;

            String _allowSubCHAR1 = "0123456789";
            String _allowSubCHAR2 = "abcdefghijklmnopqrstuvwxyz";
            String _allowSubCHAR3 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
            String _allowSubCHAR4 = "-~`!@#%^*()_-+=|{}[].?/";
            String _allowCHAR = _allowSubCHAR1+_allowSubCHAR2+_allowSubCHAR3+_allowSubCHAR4;

            for(int i=0; i<passwordType.length;i++)
                passwordType[i] = false;

            for (int i = 0; i < Value.length(); i++)
            {
                int _charValue = Value.charAt(i);
                if(_allowCHAR.indexOf(_charValue) < 0 && _charValue != '$') return "fail";

                if(passwordType[0] == false && _allowCHAR.indexOf(_charValue) >= 0) passwordType[0] = true;
                if(passwordType[1] == false && _allowCHAR.indexOf(_charValue) >= 0) passwordType[1] = true;
                if(passwordType[1] == false && _allowCHAR.indexOf(_charValue) >= 0) passwordType[1] = true;
                if(passwordType[2] == false && _allowCHAR.indexOf(_charValue) >= 0) passwordType[2] = true;
                if(passwordType[2] == false && _charValue == '$') passwordType[2] = true;
            }

            for(int i=0; i<passwordType.length;i++)
                if(passwordType[i] == true)
                    countType++;

            if(countType < 2)
            {
                Tools.Log.Print(LogLevel.ll_Error, "Password type less 2 : "+Value);
                return "fail";
            }

            for(int i = 0; i < _allowCHAR.length(); i++)
            {
                String _e = "0000";
                _e = _e.replaceAll("0", _allowCHAR.substring(i, i+1));
                if(Value.indexOf(_e) > -1)
                {
                    Tools.Log.Print(LogLevel.ll_Error, "Password char continue : "+_e);
                    return _e;
                }
            }

            if(Value.indexOf("$$$$") > -1)
            {
                Tools.Log.Print(LogLevel.ll_Error, "Password char continue : $");
                return "fail";
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return "success";
    }
    //---------------------------------------------------------------------------------------
    public static String CheckContinueNumber(int startNum, int endNum, String Value)
    {
        try
        {
            int count = 0;
            for (int i = startNum; i < endNum; i++)
            {
                StringBuffer sTemp2 = new StringBuffer();

                if (i <= endNum - 3)
                {
                    for (int j = i; j < i + 4; j++)
                        sTemp2.append(String.valueOf((char) j));
                }
                else
                {
                    count++;
                    for (int k = 1; k <= (4 - count); k++)
                        sTemp2.append(String.valueOf((char) (i + k)));

                    for (int k = 0; k < (4 - (4 - count)); k++)
                        sTemp2.append(String.valueOf((char) (startNum + k)));
                }


                if (Value.indexOf(sTemp2.toString()) > -1)
                    return sTemp2.toString();
            }
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return "success";
    }
    //---------------------------------------------------------------------------------------
    public static String GetLocaleLanguage()
    {
        try
        {
            Locale l = Locale.getDefault();
            return String.format("%s-%s", l.getLanguage(), l.getCountry());
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return "";
    }
    //-------------------------------------------------------------------------------------
    public static byte[] File2Bytes(String _fn)
    {
        try
        {
            if(_fn != null && _fn.length() > 0)
            {
                FileInputStream _is = new FileInputStream(_fn);
                byte[] _data = null;

                if(_is.available() > 0)
                {
                    _data = new byte[_is.available()];
                    _is.read(_data);
                }

                _is.close();
                return _data;
            }
        }
        catch(Exception _e){}
        return null;
    }
    //-------------------------------------------------------------------------------------
    public static boolean Bytes2File(String _fn, byte[] _payload)
    {
        try
        {
            if(_fn != null && _fn.length() > 0)
            {
                FileOutputStream _os = new FileOutputStream(_fn);

                _os.write(_payload);
                _os.flush();
                _os.close();

                return true;
            }
        }
        catch(Exception _e){}
        return false;
    }
    //-------------------------------------------------------------------------------------
    public static String GetLocalIPv4Address()
    {
        try
        {
            String _ip = "";

            try
            {
                if(_ip.length() <= 0)
                {
                    Enumeration<NetworkInterface> _networks = NetworkInterface.getNetworkInterfaces();
                    while(_networks != null && _networks.hasMoreElements())
                    {
                        _ip = "";

                        Enumeration<InetAddress> _ips = _networks.nextElement().getInetAddresses();
                        while(_ips.hasMoreElements())
                        {
                            _ip = _ips.nextElement().getHostAddress();
                            if(_ip.length() > 0 && _ip.equals("127.0.0.1") == false && _ip.split("\\.").length == 4)
                                return _ip;
                        }
                    }
                }
            }
            catch(Exception _e){}

            try
            {
                return InetAddress.getLocalHost().getHostAddress();
            }
            catch(Exception _e){}
        }
        catch(Exception _e){ ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return "";
    }
    //-------------------------------------------------------------------------------------
    public static boolean IsWifi5Ghz(Context _context)
    {
        try
        {
            WifiManager _wifiManager = (WifiManager) _context.getApplicationContext().getSystemService(Context.WIFI_SERVICE);
            WifiInfo _info = _wifiManager.getConnectionInfo();
            String _freq = "";

            if(_info != null && _info.getSupplicantState() == SupplicantState.COMPLETED)
            {
                if (Build.VERSION.SDK_INT >= 21)
                {
                    _freq = ""+_info.getFrequency();
                }
                else
                {
                    String _ssid = _info.getSSID();
                    List<ScanResult> scanResults = _wifiManager.getScanResults();
                    for(ScanResult _sr : scanResults)
                        if(_sr.SSID.equals(_ssid))
                        {
                            _freq = ""+_sr.frequency;
                            break;
                        }
                }

                if(_freq.startsWith("5")) return true;
            }
        }
        catch(Exception _e){ ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return false;
    }
    //---------------------------------------------------------------------------------------
    /*
     * Chris + 190705.
     * 判斷應用程式是否在前景.
     */
    public static boolean isAppForeground(Context _context)
    {
        try
        {
            ActivityManager _manager = (ActivityManager) _context.getSystemService(Context.ACTIVITY_SERVICE);
            List<ActivityManager.RunningAppProcessInfo> appProcessInfos = _manager.getRunningAppProcesses();

            if (appProcessInfos == null || appProcessInfos.isEmpty()) return false;

            for (ActivityManager.RunningAppProcessInfo info : appProcessInfos) {
                // 應用程式運行中, 並且正在前台.
                if (info.processName.endsWith(_context.getPackageName()) && info.importance == ActivityManager.RunningAppProcessInfo.IMPORTANCE_FOREGROUND)
                    return true;
            }
        }
        catch(Exception _e){ ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
        return false;
    }
    //---------------------------------------------------------------------------------------
}
