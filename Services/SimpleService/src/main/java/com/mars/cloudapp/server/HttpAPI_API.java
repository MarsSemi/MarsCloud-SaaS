package com.mars.cloudapp.server;
import java.net.HttpURLConnection;
import java.util.List;

import org.apache.commons.io.IOUtils;
import org.json.JSONArray;
import org.json.JSONObject;
import com.sun.net.httpserver.*;
import com.mars.cloud.MarsService;
import com.mars.cloud.ServiceData;
import com.mars.cloud.Tools;
//-------------------------------------------------------------------------------------
//
//-------------------------------------------------------------------------------------
@SuppressWarnings("restriction")
public class HttpAPI_API implements HttpHandler
{		
	//--------------------------------------------------------------------------------------
	//
	//-------------------------------------------------------------------------------------
	public MarsService _Service = null;	
	public int _ExtFuncID = -1;
	//-------------------------------------------------------------------------------------
	public HttpAPI_API(MarsService _service)
	{
		try
		{		
			_Service = _service;
			_ExtFuncID = Global._JSExcuter.LoadFromFile("./extJS/common.js", true, 300);
		}
		catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }  
	}
	//-------------------------------------------------------------------------------------
	public String SayHello(HttpExchange _http, JSONObject _params, String _data)
    {
    	try
    	{  			
			return "hello";
    	}
    	catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }  
    	return null;
    }    
	//-------------------------------------------------------------------------------------
	public String AddData(HttpExchange _http, JSONObject _params, JSONObject _data)
    {
    	try
    	{  			
    		String _id = _data.optString("user_id");
    		
    		if(_id != null)
    		{
    			if(_Service._MarsClient.PutData(Global._UUID_Company, Global._UUID_Company, _id, _data))
        			return "ok";
    		}
    	}
    	catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }  
    	return null;
    }    
	//-------------------------------------------------------------------------------------
	public String DelData(HttpExchange _http, JSONObject _params, JSONObject _data)
    {
    	try
    	{  			
    		String _id = _data.optString("user_id");
    		
    		if(_id != null)
    		{
    			if(_Service._MarsClient.DeleteDataByKey(Global._UUID_Company, Global._UUID_Company, _id))
        			return "ok";
    		}
    	}
    	catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }  
    	return null;
    }    
	//-------------------------------------------------------------------------------------
	public String ListData(HttpExchange _http, JSONObject _params, JSONObject _data)
    {
    	try
    	{  			
    		JSONObject _resp = _Service._MarsClient.GetLastData(Global._UUID_Company, Global._UUID_Company, 0);
    		
    		if(_resp != null)
    			return _resp.toString();
    	}
    	catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }  
    	return null;
    }    
	//-------------------------------------------------------------------------------------
	public String TestExtJS(HttpExchange _http, JSONObject _params, String _data)
    {
    	try
    	{  							    		
			return Global._JSExcuter.Call(_ExtFuncID, "MyFirstExtJS", _data.toString()).toString();
    	}
    	catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }  
    	return null;
    }    
	//-------------------------------------------------------------------------------------
	public JSONObject ParseParams(String _req)
	{
		try
		{
			String[] _params = _req.split("\\&");
			JSONObject _payload = new JSONObject();
			
			for(String _item : _params)
			{
				String[] _texts = _item.split("=", 2);
				if(_texts.length == 2)
					_payload.put(_texts[0], _texts[1]);
			}
			
			return _payload;
		}
		catch (Exception _e) {Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass());}
		return null;
	}
	//-------------------------------------------------------------------------------------	
	public void Process(HttpExchange _http)
	{
		try
		{  	        					
			if(_Service != null)
			{
				String[] _request = _http.getRequestURI().toString().split("\\?");
	    		String[] _items = _request[0].split("/");
	    		String _last_item =_items[_items.length-1];
	    		String _RContent = null;
	    		
	    		List<String> _auth_list=_http.getRequestHeaders().get("Authentication");
	    		String _authToken = "";
	    		JSONObject _params = _request.length > 1 ? ParseParams(_request[1]) : new JSONObject();
	    		JSONObject _jwt_payload = null;
			String _bodyText = IOUtils.toString(_http.getRequestBody(), ServiceData._DefaultCharset);
			JSONObject _body = _bodyText.startsWith("{") ? new JSONObject(_bodyText) : new JSONObject();
	    		        			        			        		
	    		try
	    		{
	    			if(_auth_list != null && _auth_list.size() > 0)
	        		{
	        			_authToken = _auth_list.get(0).replace("Bearer ", "");
	        			_jwt_payload = ServiceData.VerifyToken(_authToken, null, null);  
	        		}
	    		}
	    		catch(Exception _e){ _jwt_payload = null; }      		
	    		
	    		if(_jwt_payload != null)
	    		{	
	    			switch(_last_item)
	    			{
					case "add_data": _RContent = AddData(_http, _params, _body); break;
					case "del_data": _RContent = DelData(_http, _params, _body); break;
					case "list_data": _RContent = ListData(_http, _params, _body); break;
				}
	    		}
	    		
	    		if(_RContent == null)
		    		switch(_last_item)
				{
					case "hello": _RContent = SayHello(_http, _params, _bodyText); break;
					case "test_ext_js": _RContent = TestExtJS(_http, _params, IOUtils.toString(_http.getRequestBody(), ServiceData._DefaultCharset)); break;
				}
	    		
	    		if(_RContent != null && _RContent.length() > 0)
	    		{
	    			ServiceData.SendRespone(_http, HttpURLConnection.HTTP_OK, "application/json; charset=UTF-8", _RContent.getBytes(ServiceData._DefaultCharset));	
	    			return;
	    		}
			}			
		}
    	catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
		ServiceData.SendRespone(_http, HttpURLConnection.HTTP_NOT_FOUND, null, null);		
	}
	//-------------------------------------------------------------------------------------	
	public void handle(HttpExchange _http)
	{
		try
		{  			
			Process(_http);
		}
    	catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }
	}
	//-------------------------------------------------------------------------------------
}
