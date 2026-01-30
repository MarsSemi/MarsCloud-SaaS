package com.mars.cloudapp.server;
import java.io.OutputStream;
import java.io.PrintStream;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.util.Arrays;

import com.mars.cloud.MarsService;
import com.mars.cloud.Tools;
//--------------------------------------------------------------------------------------
//
//--------------------------------------------------------------------------------------
public class MainServer
{	

	//--------------------------------------------------------------------------------------
	//
	//--------------------------------------------------------------------------------------
	public static class Service extends MarsService
	{
		//--------------------------------------------------------------------------------------
		public Service(String _propertyFileName)
		{
			super(_propertyFileName);
		}			
		//--------------------------------------------------------------------------------------
		public void OnMQTTMessage(String _topic, String _payload)
		{
			try
			{
				Global.OnMQTTMessage(_topic, _payload);
			}
			catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }				
		}
		//--------------------------------------------------------------------------------------
		public void BeforeServiceStop()
		{
			try
			{
				//Do something berfore service close
			}
			catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }				
		}
		//--------------------------------------------------------------------------------------
		public void Process()
		{
			try
			{
				//Service looping process
			}
			catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }			
		}
		//--------------------------------------------------------------------------------------
	}
	//--------------------------------------------------------------------------------------
	//
	//--------------------------------------------------------------------------------------
    public static void main( String[] _args ) throws ClassNotFoundException
    {    	
    	try
	{   		        		
    		System.gc();
    		System.err.close();
    		System.setErr(new PrintStream(new OutputStream() { public void write(int _b) {}; }));  
    		
    		Tools.Log.SetRemoteOutput(true);
    		Tools.EnableUncaughtExceptionHandler(MainServer.class.getName(), 32, null);
    		
    		String _initFile = _args.length > 0 ? _args[0] : "agent.properties";
    		
		Global.InitAll(MainServer.class, new Service(_initFile));	
	}
	catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }	
    	finally { System.gc(); }
    }
    //--------------------------------------------------------------------------------------
}
