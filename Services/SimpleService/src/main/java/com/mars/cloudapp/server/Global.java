package com.mars.cloudapp.server;

import java.util.Timer;
import java.util.TimerTask;

import org.apache.http.HttpResponse;
import org.apache.http.HttpStatus;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.config.SocketConfig;
import org.apache.http.entity.ByteArrayEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.json.JSONObject;

import com.mars.cloud.ExtScriptManager;
import com.mars.cloud.HttpAPI_Basic;
import com.mars.cloud.HttpAPI_System;
import com.mars.cloud.JavaScriptExecutor;
import com.mars.cloud.MarsClient;
import com.mars.cloud.MarsService;
import com.mars.cloud.Tools;
//-------------------------------------------------------------------------------------
public class Global
{		
	//-------------------------------------------------------------------------------------
	//
	//-------------------------------------------------------------------------------------
	public static MarsService _Service = null;
	public static ExtScriptManager.IScriptExecutor _JSExcuter = new JavaScriptExecutor();
	//-------------------------------------------------------------------------------------
	public static String _UUID_Company = "company";
	public static String _SUID_Employee = "employee";
	//-------------------------------------------------------------------------------------
	public static void InitAll(Class<?> _class, MarsService _service)
	{
		try
		{	
			_Service = _service;	
			_Service.AddRestfulAPI("/", new HttpAPI_Basic());
			_Service.AddRestfulAPI("/system", new HttpAPI_System(_Service));
			_Service.AddRestfulAPI("/api", new HttpAPI_API(_Service));
			
			_Service.RegistryServerInfo(Tools.GetPackageVersion(_class), "myService", true, true);
			_Service.start();		
		}
		catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }  
	}
	//--------------------------------------------------------------------------------------
	public static void OnMQTTMessage(String _topic, String _payload)
	{
		try
		{
			
		}
		catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }				
	}	
	//-------------------------------------------------------------------------------------
}
//-------------------------------------------------------------------------------------