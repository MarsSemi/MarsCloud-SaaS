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
			}
			catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }				
		}
		//--------------------------------------------------------------------------------------
		public void BeforeServiceStop()
		{
			try
			{
			}
			catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); }				
		}
		//--------------------------------------------------------------------------------------
		public void Process()
		{
			try
			{
				final long _oneDaySecond = 86400;
				final long _reloginTimeout = (long) (_oneDaySecond*0.95*1000);
	        	
	        	while(_MarsClient != null)
				{
					Thread.sleep(60*1000);
					if(System.currentTimeMillis() - _SystemStartTime >= _reloginTimeout)
						RestartService();
				}
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
