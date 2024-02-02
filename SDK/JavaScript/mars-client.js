//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
new function()
{
	let _HadJQuery = false;
	let _HadMQTT = false;	
	
	try
	{
		let _head = document.getElementsByTagName('head');
				
		if(_head != null)
			for(let i=0;i<_head.length;i++)
			{
				let _children = _head[i].children;
				if(_children != null)
					for(let j=0;j<_children.length;j++)
					{
						if(_children[j].src != null && _children[j].src.indexOf('jquery') > 0) _HadJQuery = true;
						if(_children[j].src != null && _children[j].src.indexOf('paho-mqtt') > 0) _HadMQTT = true;
					}
			}
	}
	catch(_e){}
	
	if(_HadJQuery == false)
	{
		//let _script = document.createElement('script');
		//_script.src = '/portal/app/jquery/jquery.min.js';
		//document.head.appendChild(_script);
	}
	
	if(_HadMQTT == false)
	{
		let _script = document.createElement('script');
		_script.src = '/apps/paho-mqtt/paho-mqtt-min.js';
		document.head.appendChild(_script);
	}
}
//-----------------------------------------------------------------------
function GetURLParams()
{
	let _params = [];
	if (location.search != null)
	{
	    let _parts = location.search.substring(1).split('&');
	    for(let i = 0; i < _parts.length; i++)
	    {
	    	let _nv = _parts[i].split('=');
	    	if(_nv.length != 2) continue;
	    	_params[ _nv[0] ] = _nv[1];
	    }
	}
	
	return _params;
}
//-----------------------------------------------------------------------
function DownloadExcelFile(_fileName , _result)
{   
  try
  {
    if(_result!=null && _result.length>0)
    {        
      _result = JSON.parse(_result);
      if(_result != null)_result = _result.content
      if(_result != null)
      {
        let bin = window.atob(_result);
        let ab = s2ab(bin); 
        let blob = new Blob([ab], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;' });
        let a = document.createElement("a");

        document.body.appendChild(a);
        a.style = "display: none";
        a.href = window.URL.createObjectURL(blob);
        a.download = _fileName;
        a.click();      
        return;
      }
    }
  }   
  catch(_e){console.log(_e);}
 alert("查無資料"); 
}   
//-----------------------------------------------------------------------
function s2ab(s) 
{
  let buf = new ArrayBuffer(s.length);
  let view = new Uint8Array(buf);
  for (let i=0; i!=s.length; ++i) view[i] = s.charCodeAt(i) & 0xFF;
  return buf;
}
//-----------------------------------------------------------------------
function ShowWaitDialog(_id)
{
	
}
//-----------------------------------------------------------------------
function CloseWaitDialog(_id)
{
}
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
//
//-----
// Login & Registry
//-----
class CloudUser
{		
	constructor()
	{
		this._account = localStorage.getItem('mars_cloud_auth_account');
		this._id =  new Date().getTime();
		this._SubscribeClients = [];
		
		if(_User != null && _User._id != this._id)
		{
			_User.Close();
			_User = this;
		}
		
		let _CurrentUser = this;
		let _prev_unload = window.onunload;
		let _prev_unbeforeload = window.onbeforeunload;
		
		window.onbeforeunload  = 	
		window.onunload = function()
		{ 
			if(_prev_unbeforeload != null) _prev_unbeforeload();
			if(_prev_unload != null) _prev_unload();
			
			_CurrentUser.Close();
			return null;
		};
	}
}
//-----
var _User = new CloudUser();
let _GlobalDataOwner = null;
let _Client = _User;
//-----
_User._OneDayMS=3600000;
_User._10DaysMS=10*_User._OneDayMS;
_User._60DaysMS=60*_User._OneDayMS;
_User._HTTPTimeOut = 3000;
_User.TimezoneOffset = new Date().getTimezoneOffset()*60000;
_User.ClientType = 'server';
//-----
_User.Init = function()
{
	this.Close();
	this._SubscribeClients = [];
};
//-----
_User.Close = function()
{
	if(this.CloseSubscribClient != null)
		this.CloseSubscribClient(null);
};
//-----
//
//-----
_User.http = _User.http_get = _User.http_get_unauth = function(_req)
{
	try
	{
		let xmlHttp = new XMLHttpRequest();
	    
	    xmlHttp.open("GET", _req, false);
	    //xmlHttp.timeout = _User._HTTPTimeOut;
	    xmlHttp.ontimeout = xmlHttp.onerror = function(){};
	    xmlHttp.onreadystatechange = function(){};
	    xmlHttp.send();
	    
	    if(xmlHttp.readyState == 4 && xmlHttp.status == 200)	    	
	    	return xmlHttp.responseText;
	}
	catch(_e){};  
	return null;
}
//-----
_User.http_get_sync = function(_req)
{	
	try
	{
		let _token = this.WasLogin();
		let xmlHttp = new XMLHttpRequest();		
		
	    xmlHttp.open("GET", _req, false);	
	    xmlHttp.setRequestHeader("Authentication", _token ? "Bearer "+_token : "");
	    xmlHttp.setRequestHeader("Cache-Control", "no-cache");
	    
	    xmlHttp.send();		    
	    		    
	    if(xmlHttp.readyState == 4 && xmlHttp.status == 200) return xmlHttp.responseText;			    	
	}
	catch(_e){};
	
	return null;
}
//-----
_User.http_get_async = function(_req, _callback, obj)
{
	try
	{		
		if(_callback == null) return null;
		
		let _token=this.WasLogin();
		let xmlHttp = new XMLHttpRequest();
				
		xmlHttp.open("GET", _req, _callback == null ? false : true);	
		xmlHttp.setRequestHeader("Authentication", _token ? "Bearer "+_token : "");
		xmlHttp.setRequestHeader("Cache-Control", "no-cache");
		
		xmlHttp.onreadystatechange = function()
		{								
			if(xmlHttp.readyState == 4)
		    {
				try
	    		{						
	    			if(xmlHttp.status == 200)
			      		  _callback(xmlHttp.responseText, obj);
			      	  else
			      		  _callback(null, obj);
	    		}
	    		catch(_e){}    			
		    }
		}
    
		xmlHttp.send();    
		if(xmlHttp.readyState == 4 && xmlHttp.status == 200)
			return xmlHttp.responseText;
	}
	catch(_e){}  	
	return null;
}
//-----
_User.http_get_progress = function(_req, _callback)
{	
	try
	{
		let _token = this.WasLogin();
		let xmlHttp = new XMLHttpRequest();		
		
	    xmlHttp.open("GET", _req, false);	
	    xmlHttp.setRequestHeader("Authentication", _token ? "Bearer "+_token : "");
	    xmlHttp.setRequestHeader("Cache-Control", "no-cache");

		if(_callback != null)
            xmlHttp.download.addEventListener('progress', _callback);
	    
	    xmlHttp.send();		    
	    		    
	    if(xmlHttp.readyState == 4 && xmlHttp.status == 200)
			return xmlHttp.responseText;			    	
	}
	catch(_e){};	
	return null;
}
//-----
_User.http_get_stream = function(_req, _callback, obj)
{	
	let _token=this.WasLogin();
	if(_token != null && _callback != null)   
	{    
		let xmlHttp = new XMLHttpRequest();
		
		xmlHttp.open("GET", _req, true);
		xmlHttp.multipart = true; 
		 //xmlHttp.timeout = _User._HTTPTimeOut;
		xmlHttp.overrideMimeType('text/plain; charset=ISO-8859-1');
		xmlHttp.setRequestHeader("Authentication", "Bearer "+_token); 	
		
		if(_callback != null)
			xmlHttp.download.addEventListener('progress', function(_e){ _callback(xmlHttp, obj, _e); });
		
		xmlHttp.send();
	}
}
//-----
_User.http_post_unauth = function(_req)
{
	try
	{
		let xmlHttp = new XMLHttpRequest();
	    
		xmlHttp.open("POST", _req, false);
		xmlHttp.send(_payload);   
	            
	    if(xmlHttp.readyState == 4 && xmlHttp.status == 200) return xmlHttp.responseText;
	}
	catch(_e){};  
	return null;
}
//-----
_User.http_post_sync = function(_req, _payload)
{
	try
	{
		let _token = this.WasLogin();  
		let xmlHttp = new XMLHttpRequest();
	    
		xmlHttp.open("POST", _req, false);
		//xmlHttp.timeout = _User._HTTPTimeOut;
		xmlHttp.setRequestHeader("Authentication", _token != null ? "Bearer "+_token : ""); 
		xmlHttp.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
	    xmlHttp.setRequestHeader("Cache-Control", "no-cache");
		xmlHttp.send(_payload);   
	            
	    if(xmlHttp.readyState == 4 && xmlHttp.status == 200) return xmlHttp.responseText;
  }
  catch(_e){}    
  return null;
}   
//-----
_User.http_post_sync_stream = function(_req, _payload)
{
  try
  {
	  let _token=this.WasLogin();     
      
	  if(_token != null)   
	  {    
	    let xmlHttp = new XMLHttpRequest();
	    	    
	    xmlHttp.open("POST", _req, false);
	    xmlHttp.setRequestHeader("Authentication", "Bearer "+_token); 
	    xmlHttp.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
	    xmlHttp.setRequestHeader("Cache-Control", "no-cache");
	    xmlHttp.send(_payload);   
	            
	    if(xmlHttp.readyState == 4 && xmlHttp.status == 200)
	    	return xmlHttp.response;
	  }
  }
  catch(_e){}    
  return null;
}
//-----
_User.http_post_async = function(_req, _payload, _callback, obj)
{
	try
	{
		let _token = this.WasLogin();	      
		let xmlHttp = new XMLHttpRequest();
	    
	    xmlHttp.open("POST", _req, _callback == null ? false : true);
	    xmlHttp.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
	    xmlHttp.setRequestHeader("Cache-Control", "no-cache");
	    xmlHttp.setRequestHeader("Authentication", _token ? "Bearer "+_token : "");
	    
	    if(_callback != null)
		    xmlHttp.onreadystatechange = function()
		    {
		    	if(_callback != null && xmlHttp.readyState == 4)
		        {
		    		try
		    		{
		    			if(xmlHttp.status == 200)
				      		_callback(xmlHttp.responseText, obj);
				      	else
				      		_callback(null, obj);
		    		}
		    		catch(_e){}    			      	  
		        }
		    };
	        		
	    xmlHttp.send(_payload);
	    if(xmlHttp.readyState == 4 && (xmlHttp.status == 200))
	    	return xmlHttp.responseText;
		  
		return null;
	}
	catch(_e){}    
	
	if(_callback != null) _callback(null);
	return null;
}
//-----
_User.http_post_progress = function(_req, _payload, _callback)
{
	try
	{
		let _token = this.WasLogin();	      
		let xmlHttp = new XMLHttpRequest();
	    
	    xmlHttp.open("POST", _req, _callback == null ? false : true);
	    xmlHttp.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
	    xmlHttp.setRequestHeader("Cache-Control", "no-cache");
	    xmlHttp.setRequestHeader("Authentication", _token ? "Bearer "+_token : "");
	    
	    if(_callback != null)
            xmlHttp.upload.addEventListener('progress', _callback);
	        		
	    xmlHttp.send(_payload);
	    if(xmlHttp.readyState == 4 && (xmlHttp.status == 200))
	    	return xmlHttp.responseText;
		  
		return null;
	}
	catch(_e){}    
	return null;
}
//-----
_User.TestWebStorage = function()
{ 
	if(typeof(Storage) === "undefined")
		return false;
  
	return true;
}
//-----
_User.LoginRemainTime = function()
{
	try
	{
		let _time = localStorage.getItem("mars_cloud_auth_token_time");		
		if(_time != null)
		{
			_time = _time - (new Date().getTime()/1000) - 5*60;			
			if(_time > 0) return _time;
		}
	}
	catch(_e){}	
	return 0; 
};
//-----
_User.WasLogin = function(_auto_logout)
{
	try
	{
		let _token = localStorage.getItem("mars_cloud_auth_token");
		if(_token != null && this.LoginRemainTime() > 0 && this.ReadRegLoginURL() != null) return _token;
	}
	catch(_e){}	
	      
	if(_auto_logout) this.DoLogout();
	return null; 
}
//-----
_User.VerifyAuthToken = function()
{
	try
	{
		let _token = this.http_get_sync("/auth/waslogin?");
		if(_token != null && _token.length > 30) return _token;
	}
	catch(_e){}
	return null; 
}
//-----
_User.UpdateToken = function(_account, _token, _time, _simulatedUser)
{
	try
	{

		if(_time <= 0) _time = 86400;
		
		 _time = new Date().getTime()/1000 + _time;
		
		localStorage.setItem("mars_cloud_auth_account", _account);
		localStorage.setItem("mars_cloud_auth_token", _token);
		localStorage.setItem("mars_cloud_auth_token_time", _time);		
		localStorage.removeItem("mars_cloud_auth_simulated_user");
		
		if(_simulatedUser != null) localStorage.setItem("mars_cloud_auth_simulated_user", _simulatedUser);		
	}
	catch(_e){ console.log("UpdateToken :" + _e); }		
};
//-----
_User.GetLoginDetail = function()
{
	try
	{
		let _token = this.WasLogin();
		if(_token)
			return this.http_get_sync("/api/waslogin?");
	}
	catch(_e){}	
	      
  localStorage.setItem("mars_cloud_auth_account", null);
  localStorage.setItem("mars_cloud_user_name", null);
  return null; 
};
//-----
_User.SaveRegLoginURL = function(_url)
{
	localStorage.setItem("mars_cloud_login_url", _url);
}
//-----
_User.ReadRegLoginURL = function()
{
	return localStorage.getItem("mars_cloud_login_url");
}
//-----
_User.GetCurrentUserID = function()
{
	return localStorage.getItem("mars_cloud_auth_account");
}
//-----
_User.GetCurrentUserName = function()
{
	return localStorage.getItem("mars_cloud_user_name");
}
//-----
_User.GetCurrentUserGroup = function()
{
	return localStorage.getItem("mars_cloud_user_group");
}
//-----
_User.GetCurrentUserType = function()
{
	return localStorage.getItem("mars_cloud_user_type");
}
//-----
_User.GetCurrentUserProj = function()
{
	let _proj = localStorage.getItem("mars_cloud_user_proj");
	return (_proj && _proj.length > 0) ? (''+_proj) : 'none';
}
//-----
_User.SetCurrentUserProjName = function(_name)
{
	localStorage.setItem("mars_cloud_user_proj_name", _name);
}
//-----
_User.GetCurrentUserProjName = function()
{
	let _name = localStorage.getItem("mars_cloud_user_proj_name");
	return (_name && _name.length > 0) ? _name : '標準雲端服務';
}
//-----
_User.DoLogin = function(_usr, _pwd, _proj)
{
	return this.DoLoginAdv(_usr, _pwd, _proj, false, null, null);
}
//-----
_User.DoLoginAdv = function(_usr, _pwd_or_token, _proj, _isUseLongTerm, _simulatedUser, _relogin_url)
{
	try
	{
		if(this.TestWebStorage() === false)
	    {
  			alert('This browser does not support login function.');
  			return false;
	    }
		
		_usr = _usr.replace("@","-");
				    
		let _org_account = localStorage.getItem("mars_cloud_auth_account");
		let _curr_proj = localStorage.getItem("mars_cloud_user_proj");

	    if(this.WasLogin() != null && _org_account == _usr && _curr_proj == _proj) return true;    
	            
	    let xmlHttp = new XMLHttpRequest();   
	    let _url = "/auth/login?";
	    let _payload = {};
	    
	    _payload.usr = _usr;
	    _payload.pwd = _pwd_or_token;
	    
	    if(_proj && _proj.length > 0)
	    	_payload.proj = _proj;
	    
	    xmlHttp.open("POST", _url, false);
	    xmlHttp.send(JSON.stringify(_payload));         
	    	    	        
	    if(xmlHttp.readyState == 4 && xmlHttp.status == 200)
	    {   	    	
	    	let _org_token = localStorage.getItem("mars_cloud_auth_token");
	    	let _token = xmlHttp.responseText;
	    	let _login_url = _relogin_url ? _relogin_url : (window.location == null ? window.location : window.location.href);
	    		    	
	    	this.UpdateToken(_usr, _token, 0, _simulatedUser);
	    	this.SaveRegLoginURL(_login_url);

	    	if(_isUseLongTerm) _token = this.ReqLongTermToken(_usr, _token);	

	    	if(_usr != _simulatedUser && _simulatedUser != null)
	    	{
		    	let _payload =JSON.parse(this.http_get_sync("/api/waslogin?"));
		    	let _group = _payload.group.toLowerCase();
		    	if(_group != 'administrator' && _group != 'manager') _simulatedUser = null;		
	    	}	   
	    	
	        let _Info = this.GetUserInfo(null); 
	        	        
        	if(_Info != null) _Info = JSON.parse(_Info);           
	        if(_Info != null && _pwd_or_token == _org_token && _proj == null)
	        {
	        	localStorage.setItem("mars_cloud_auth_org_account", _org_account);
	        	localStorage.setItem("mars_cloud_auth_org_token", _org_token);
	        	localStorage.setItem("mars_cloud_auth_org_time", localStorage.getItem("mars_cloud_auth_token_time"));
	        	localStorage.setItem("mars_cloud_auth_org_name", localStorage.getItem("mars_cloud_user_name"));
	        	localStorage.setItem("mars_cloud_auth_org_group", localStorage.getItem("mars_cloud_user_group"));
	        	localStorage.setItem("mars_cloud_auth_org_type", localStorage.getItem("mars_cloud_user_type"));
	        	//localStorage.removeItem("mars_cloud_auth_simulated_user");
	        }
	       
	        if(_Info != null) localStorage.setItem("mars_cloud_user_name", _Info.name);
	        if(_Info != null) localStorage.setItem("mars_cloud_user_group", _Info.group);
	        if(_Info != null) localStorage.setItem("mars_cloud_user_type", _Info.user_type);	        
	        if(_Info != null) localStorage.setItem("mars_cloud_user_proj", _proj);
	        
	    	return true;
	    }
	}
	catch(_e){ alert('User information is incorrect or non-actived.'); }	
    
	this.DoLogout();
	
	alert('Login fail ~');
	return false;
};
//-----
_User.ReqAuthTokenByDay = function(_usr, _token, _ttl_day)
{    
    if(_usr == null) _usr = localStorage.getItem("mars_cloud_auth_account");
    if(_token == null) _token = localStorage.getItem("mars_cloud_auth_token");
    
    return this.http_post_sync("/auth/get_longterm_auth?usr="+_usr+"&ttl="+(_ttl_day*86400), _token);
};
//-----
_User.ReqLongTermToken = function(_usr, _token)
{         
    let _new_token = this.ReqAuthTokenByDay(_usr, _token, 7);
    
    if(_new_token != null && _new_token.length > 30)
    {
    	this.UpdateToken(_usr, _new_token, 7*86400, _simulatedUser);    	  		    	
    	return _new_token;
    }
    	
    return null;
};
//-----
_User.GetOAuthByKey = function(_key)
{    
    if(_key == null) return;    
    return this.http_post_sync("/auth/get_auth_by_key?", _key);
};
//-----
_User.ResetDevelopKey = function(_usr, _token, _permission)
{    
	try
	{
		if(_usr == null) _usr = localStorage.getItem("mars_cloud_auth_account");
	    if(_token == null) _token = localStorage.getItem("mars_cloud_auth_token");
	    if(_permission == null) return;
	    
	    return this.http_post_sync("/auth/reset_developer_key?usr="+_usr+"&permission="+_permission, _token);
	}
	catch(_e){}
	return null;
};
//-----
_User.ResetProjectDevelopKey = function(_proj, _token, _permission)
{    
	try
	{
	    if(_token == null) _token = localStorage.getItem("mars_cloud_auth_token");
	    if(_proj == null) return;
	    if(_permission == null) return;
	    
	    return this.http_post_sync("/auth/reset_developer_key?proj="+_proj+"&permission="+_permission, _token);
	}
	catch(_e){}
	return null;
};
//-----
_User.WasSimulatedUser = function()
{                 
	 return (localStorage.getItem("mars_cloud_auth_org_token") != null);
};
//-----
_User.GetSimulatedUser = function()
{                 
    let _simulatd_user = localStorage.getItem("mars_cloud_auth_simulated_user");    
    if(_simulatd_user == "null") _simulatd_user = null;
    
    return _simulatd_user;
};
//-----
_User.DoLogout = function(_token)
{                 
    try
    {    	
    	if(this.WasSimulatedUser())
		{    		
    		this.DoSimulateLogout();
    		return;
		}
    	
        localStorage.removeItem("mars_cloud_auth_account");
        localStorage.removeItem("mars_cloud_auth_token");    
        localStorage.removeItem("mars_cloud_auth_token_time");
        localStorage.removeItem("mars_cloud_auth_simulated_user");
        
        localStorage.removeItem("mars_cloud_auth_org_account");
        localStorage.removeItem("mars_cloud_auth_org_token");
        localStorage.removeItem("mars_cloud_auth_org_time");
        localStorage.removeItem("mars_cloud_auth_org_name");
        localStorage.removeItem("mars_cloud_auth_org_group");
        localStorage.removeItem("mars_cloud_auth_org_type");
        
        localStorage.removeItem("mars_cloud_user_name");
        localStorage.removeItem("mars_cloud_user_group");
        localStorage.removeItem("mars_cloud_user_type");
        localStorage.removeItem("mars_cloud_user_proj");
        localStorage.removeItem("mars_cloud_user_proj_name");        
    }
   catch(_e){}
    
    let _url = this.ReadRegLoginURL();
        
    if(_url != null) window.location = _url;
	//window.location.reload();
};
//-----
_User.DoSimulateLogout = function()
{                 
    try
    {    	    	
    	localStorage.setItem("mars_cloud_auth_account", localStorage.getItem("mars_cloud_auth_org_account"));
		localStorage.setItem("mars_cloud_auth_token", localStorage.getItem("mars_cloud_auth_org_token"));
		localStorage.setItem("mars_cloud_auth_token_time", localStorage.getItem("mars_cloud_auth_org_time"));				
        localStorage.setItem("mars_cloud_user_name", localStorage.getItem("mars_cloud_auth_org_name"));
        localStorage.setItem("mars_cloud_user_group", localStorage.getItem("mars_cloud_auth_org_group"));
        localStorage.setItem("mars_cloud_user_type", localStorage.getItem("mars_cloud_auth_org_type"));
        
        localStorage.removeItem("mars_cloud_auth_simulated_user");        
        localStorage.removeItem("mars_cloud_auth_org_account");
        localStorage.removeItem("mars_cloud_auth_org_token");
        localStorage.removeItem("mars_cloud_auth_org_time");
        localStorage.removeItem("mars_cloud_auth_org_name");
        localStorage.removeItem("mars_cloud_auth_org_group");
        localStorage.removeItem("mars_cloud_auth_org_type");
        
        location.reload();
    }
    catch(_e){}
    
    let _url = this.ReadRegLoginURL();
    if(_url != null) window.location = _url;
};
//-----
_User.ApplyForNewUser = function(_data)
{        	
  let xmlHttp = new XMLHttpRequest();
    
  xmlHttp.open("POST", "/auth/registry?target=user", false);
  xmlHttp.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
  xmlHttp.send(JSON.stringify(_data));   
      
  if(xmlHttp.readyState == 4 && xmlHttp.status == 200)        
    return true;
  else
  if(xmlHttp.responseText)
    alert(xmlHttp.responseText);
};
//-----
_User.callService = _User.CallService = function(_service, _api, _data, _callback)
{        
	_api = _api.substring(_api.startsWith('/') ? 1 : 0, 1000).replaceAll('/', '+');
	
	return this.http_post_async("/services/"+_service+'/'+_api, JSON.stringify(_data), _callback);
};
//---
// Old Method, dont use it
_User.callLocalService = _User.CallLocalService = function(_service, _api, _payload, _callback, _encode)
{
    let _data = {};

    _data.service = _service;
    _data.api = _api;
    _data.encode = _encode;
    _data.payload = _payload;

    return this.CallService("service.proxy-cloud", "/api/req_task", _data, _callback);
}
//---
// Old Method, dont use it
_User.callLocalURL = _User.CallLocalURL = function(_url, _api, _payload, _callback, _encode)
{
    let _data = {};

    _data.url = _url;
    _data.api = _api;
    _data.encode = _encode;
    _data.payload = _payload;

    return _User.CallService("service.proxy-cloud", "/api/req_task", _data, _callback);
}
//-----
_User.RegistryUser = function(_data)
{        
	return this.ApplyForNewUser(_data);          
};
//-----
_User.UnRegistryUser = function(_data)
{        
	_data = this.http_post_sync("/auth/unregistry?target=user", JSON.stringify(_data));
    return (_data != null);
};
//-----
_User.RegistryService = function(_data)
{        
	_data = this.http_post_sync("/auth/registry?target=server", JSON.stringify(_data));          
    return (_data != null);
};
//-----
_User.UnRegistryService = function(_data)
{        
	_data = this.http_post_sync("/auth/unregistry?target=server", JSON.stringify(_data));          
    return (_data != null);
};
//-----
// User profile
//-----
_User.GetAllDataSrc = function()
{      	
	return this.http_get_sync("/sys/get_all_datasrc"); 
}
//-----
_User.GetUserDataSrcList = function()
{      
	let _simUser = this.GetSimulatedUser();
	
	if(_simUser == null) _simUser = _GlobalDataOwner;
	if(_simUser != null) return this.http_post_sync("/api/usrinfo?method=datasrclist", JSON.stringify({user: _simUser}));   
	
	return this.http_get_sync("/api/usrinfo?method=datasrclist"); 
}
//-----
_User.GetUserDataSrcInfo = function(_src, _index, _callback)
{   			
	return this.GetUserDataSrcInfoAsync(_src, _index, _callback);
}
//-----
_User.GetUserDataSrcInfoAsync = function(_src, _index, _callback)
{   	
	if(_callback == null) return this.GetUserDataSrcInfoByUser(this.GetSimulatedUser(), _src, _index);
	return this.GetUserDataSrcInfoByUserAsync(this.GetSimulatedUser(), _src, _index, _callback);
}
//-----
_User.GetUserDataSrcInfoByUser = function(_user, _src, _index)
{   
	let _cmd = _user == null ?  { user: this.GetSimulatedUser(), uuid: _src.replace('.', '_') } : { user: _user,  uuid: _src.replace('.', '_') }; 
	return this.http_post_sync("/api/usrinfo?method=datasrcinfo", JSON.stringify(_cmd));
}
//-----
_User.GetUserDataSrcInfoByUserAsync = function(_user, _src, _index, _callback)
{   
	let _cmd = _user == null ?  { uuid: _src.replace('.', '_') } : { user: _user,  uuid: _src.replace('.', '_') }; 
	let _respone = this.http_post_async("/api/usrinfo?method=datasrcinfo", JSON.stringify(_cmd), function(_result)
	{
		if(_callback != null) _callback(_index, _result);
	});
	
    return _respone;
}
//-----
_User.AddUserDevice = function(_dev_info)
{      	
	if(this.http_post_sync("/auth/registry?target=device", JSON.stringify(_dev_info)) != null)    
		return true;
	
  	return false;
}
//-----
_User.AddUserDataSrc = _User.addUserDataSrc = _User.AddUserDataSrc = _User.AddUserDatasrc = function(_dev_info)
{      
	if(_dev_info == null) return false;
	
	_dev_info.from = this.ClientType;
	
    let _data = this.http_post_sync("/api/usrinfo?method=adddatasrc", JSON.stringify(_dev_info));    
    return (_data != null);
} 
//-----
_User.UpdateUserDataSrc = _User.updateUserDataSrc = _User.UpdateUserDevice = _User.updateUserDevice = function(_dev_info)
{      
	if(_dev_info == null) return false;
	
	_dev_info.from = this.ClientType;
	
    let _data=this.http_post_sync("/api/usrinfo?method=updatedatasrc", JSON.stringify(_dev_info));    
    return (_data != null);
}
//-----
_User.DelUserDataSrc = _User.delUserDataSrc = _User.DelUserDevice = function(_dev_info)
{      
	let _respone = this.http_post_sync("/api/usrinfo?method=deldatasrc", JSON.stringify(_dev_info));
    return (_respone != null);
}
//-----
_User.ModifyDeviceStatus = function(_dev_info)
{      	
	_dev_info.target = 'command';
	_dev_info.cmd = 'set';	
	_dev_info.type = 'status';
		
    let _data=this.http_post_sync("/api/put?command", JSON.stringify(_dev_info));    
    return (_data != null);
}
//-----
_User.SyncDeviceStatus = function(_dev_info)
{      	
	_dev_info.target='command';
	_dev_info.cmd = 'sync';	
	_dev_info.type = 'status';
		
    let _data=this.http_post_sync("/api/put?command", JSON.stringify(_dev_info));    
    return (_data != null);
}
//-----
_User.RebootDevice = function(_id)
{      	
	let _dev_info = {};
	let _ids = _id.split('.');
	
	if(_ids.length > 0)
	{
		_dev_info.uuid = _ids[0];
		_dev_info.target='command';
		_dev_info.cmd = 'reboot';
		_dev_info.type = 'device_control';
				
	    let _data=this.http_post_sync("/api/put?command", JSON.stringify(_dev_info));    
	    return (_data != null);
	}
	
    return false;
}
//-----
_User.GetTask = function(_ukey)
{      
	return this.http_post_sync("/api/get?task", JSON.stringify({ukey: _ukey}));  
}
//-----
_User.UpdateTask = function(_ukey, _task)
{      
	_task.ukey=_ukey;
	    
    return (this.http_post_sync("/api/put?task", JSON.stringify(_task)) != null);
}
//-----
_User.RemoveTask = function(_ukey)
{      
    return (this.http_post_sync("/api/del?task", JSON.stringify({ukey: _ukey})) != null);
}
//-----
_User.GetSch = function(_ukey)
{      
	return this.http_post_sync("/api/get?sch", JSON.stringify({ukey: _ukey}));  
}
//-----
_User.UpdateSch = function(_ukey, _sch)
{      	
	_sch.ukey=_ukey;
	
    return (this.http_post_sync("/api/put?sch", JSON.stringify(_sch)) != null);
}
//-----
_User.RemoveSch = function(_ukey)
{      
	return (this.http_post_sync("/api/del?sch", JSON.stringify({ukey: _ukey})) != null);
}
//-----
_User.PutEvent = _User.putEvent = function(_src, _event)
{      		
	_src = _src.replace('.', '_');	
	_src = _src.split('_');
		
	if(_src.length == 1) _event = { uuid: _src[0], from: this.ClientType,  values: [ _event ] }; 
	if(_src.length == 2) _event = { uuid: _src[0], suid: _src[1], from: this.ClientType,  values: [ _event ] }; 
	
	let _simUser = this.GetSimulatedUser();	
	if(_simUser != null) _event.user = _simUser;
	
	//console.log(_src);
	//console.log(_event);
			
    let _data = this.http_post_sync("/api/put?event", JSON.stringify(_event));    
    return (_data != null);
}
//-----
_User.PutEvents = _User.putEvents = function(_src, _events)
{      		
	_src = _src.replace('.', '_');	
	_src = _src.split('_');
		
	if(_src.length == 1) _event = { uuid: _src[0], from: this.ClientType,  values: _events }; 
	if(_src.length == 2) _event = { uuid: _src[0], suid: _src[1], from: this.ClientType,  values: _events }; 
	
	let _simUser = this.GetSimulatedUser();
	if(_simUser != null) _event.user = _simUser;
				
    let _data = this.http_post_sync("/api/put?event", JSON.stringify(_event));    
    return (_data != null);
}
//-----
//Data control functions
//-----
_User.GetDataByUser = _User.getDataByUser = function(_user, _src, _stime, _etime, _callback, _item, _isCompressed)
{                   	
	if(_callback != null) return this.GetDataAsyncByUser(_user, _src, _stime, _etime, _callback, _item, _isCompressed);
	if(_src == null || _src.length <= 0) return null;
	
	if(_stime == null) { let _date = new Date(); _stime = Math.floor(_date.getTime()/86400000)*86400000+_date.getTimezoneOffset()*60000; };
	if(_etime == null) _etime = _stime + 86400000;
		
	let _cmd= { uuid: _src.replace('.', '_'), timestamp: _stime+'~'+_etime };	
	
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
	return this.http_post_sync("/api/get?data", JSON.stringify(_cmd));
}
//-----
_User.GetData = _User.getData = function(_src, _stime, _etime, _callback, _item, _isCompressed)
{                   
	return this.getDataByUser(this.GetSimulatedUser(), _src, _stime, _etime, _callback, _item, _isCompressed);
}
//-----
_User.GetDataAsyncByUser = _User.getDataAsyncByUser = function(_user, _src, _stime, _etime, _callback, _item, _isCompressed)
{                   
	if(_src == null || _src.length <= 0) return null;
	
	if(_stime == null) { let _date = new Date(); _stime = Math.floor(_date.getTime()/86400000)*86400000+_date.getTimezoneOffset()*60000; };
	if(_etime == null) _etime = _stime + 86400000;
		
	let _cmd= { uuid: _src.replace('.', '_'), timestamp: _stime+'~'+_etime };	
		
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
		
	return this.http_post_async("/api/get?data", JSON.stringify(_cmd), _item ? function(_data){ _callback(_data, _item); } : _callback);
}
//-----
_User.GetDataAsync = _User.getDataAsync = function(_src, _stime, _etime, _callback, _item)
{                   	
	return this.getDataAsyncByUser(this.GetSimulatedUser(), _src, _stime, _etime, _callback, _item);
}
//-----
_User.GetDataTodayByUser = _User.getDataTodayByUser = function(_user, _src, _callback, _isCompressed)
{                   
	if(_src == null || _src.length <= 0) return null;
	
	let _oneday_ms = 86400000;
	let _date = new Date();
	let _curTime = _date.getTime() + _date.getTimezoneOffset();
	let _stime = Math.floor(_curTime/_oneday_ms)*_oneday_ms;
	let _etime = _stime + _oneday_ms;
	let _cmd= { uuid: _src.replace('.', '_'), timestamp: _stime+'~'+_etime };	
	
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
	return this.http_post_async("/api/get?data", JSON.stringify(_cmd), _callback);
}
//-----
_User.GetDataToday = _User.getDataToday = function(_src, _callback, _isCompressed)
{                   
	return this.GetDataTodayByUser(this.GetSimulatedUser(), _src, _callback, _isCompressed);
}
//----- 
_User.GetLastDataByUser = _User.getLastDataByUser = function(_user, _src, _count, _callback, _isCompressed)
{                   
	if(_callback != null) return this.getLastDataByUserAsync(_user, _src, _count, _callback, _isCompressed);
	if(_src == null || _src.length <= 0) return null;
	
	let _cmd= { src: _src.replace('.', '_'), count: _count };
	
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;	
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
    return this.http_post_sync("/api/lastdata?method=read", JSON.stringify(_cmd));	
}
//----- 
_User.GetLastData = _User.getLastData = function(_src, _count, _callback, _isCompressed)
{                   	
    return this.getLastDataByUser(this.GetSimulatedUser(), _src, _count, _callback, _isCompressed);	
}
//----- 
_User.GetLastDataByUserAsync = _User.getLastDataByUserAsync = function(_user, _src, _count, _callback, _isCompressed)
{                   
	if(_src == null || _src.length <= 0) return null;
	if(_callback == null) return null;
	
	let _cmd= { src: _src.replace('.', '_'), count: _count };	
	
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;	
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
    return this.http_post_async("/api/lastdata?method=read", JSON.stringify(_cmd), function(_data){ _callback(_src, _data); });	
}
//----- 
_User.GetLastDataAsync = _User.getLastDataAsync = function(_src, _count, _callback, _isCompressed)
{                   
	return this.getLastDataByUserAsync(this.GetSimulatedUser(), _src, _count, _callback, _isCompressed);	
}
//----- 
_User.GetLastDataAndOrderByUser = _User.getLastDataAndOrderByUser = function(_user, _src, _order_by, _order_type, _count, _callback, _isCompressed)
{                   
	if(_callback != null) return this.getLastDataAndOrderByUserAsync(_user, _src, _order_by, _order_type, _count, _callback, _isCompressed);
	if(_src == null || _src.length <= 0) return null;
	
	let _cmd= { src: _src.replace('.', '_'), count: _count };
	
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;
	if(_order_by != null) _cmd.order_by = _order_by;	
	if(_order_type != null) _cmd.order_type = _order_type;	
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
    return this.http_post_sync("/api/lastdata?method=read", JSON.stringify(_cmd));	
}
//----- 
_User.GetLastDataAndOrder = _User.getLastDataAndOrder = function(_src, _count, _callback, _isCompressed)
{                   	
    return this.getLastDataAndOrderByUser(this.GetSimulatedUser(), _src, _order_by, _order_type, _count, _callback, _isCompressed);	
}
//----- 
_User.GetLastDataAndOrderByUserAsync = _User.getLastDataAndOrderByUserAsync = function(_user, _src, _order_by, _order_type, _count, _callback, _isCompressed)
{                   
	if(_src == null || _src.length <= 0) return null;
	if(_callback == null) return null;
	
	let _cmd= { src: _src.replace('.', '_'), count: _count };	
	
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;
	if(_order_by != null) _cmd.order_by = _order_by;	
	if(_order_type != null) _cmd.order_type = _order_type;		
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
    return this.http_post_async("/api/lastdata?method=read", JSON.stringify(_cmd), function(_data){ _callback(_src, _data); });	
}
//----- 
_User.GetLastDataAndOrderAsync = _User.getLastDataAndOrderAsync = function(_src, _order_by, _order_type, _count, _callback, _isCompressed)
{                   
	return this.getLastDataAndOrderByUserAsync(this.GetSimulatedUser(), _src, _order_by, _order_type, _count, _callback, _isCompressed);	
}
//-----
_User.GetDataByKeyByUser = _User.getDataByKeyByUser = function(_user, _src, _key, _callback, _isCompressed)
{                   
	if(_src == null || _src.length <= 0) return null;
	if(_callback != null) return this.getDataByKeyAsync(_src, _key, _callback);
			
	let _cmd = { uuid: _src.replace('.', '_'), ukey: _key};		
	
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;	
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
	return this.http_post_sync("/api/get?data", JSON.stringify(_cmd));
}
//-----
_User.GetDataByKey = _User.getDataByKey = function(_src, _key, _callback, _isCompressed)
{                   	
	return this.GetDataByKeyByUser(this.GetSimulatedUser(), _src, _key, _callback, _isCompressed);
}
//-----
_User.GetDataByKeyAsyncByUser = _User.getDataByKeyAsyncByUser = function(_user, _src, _key, _callback, _isCompressed)
{                   
	if(_src == null || _src.length <= 0) return null;
			
	let _cmd = { uuid: _src.replace('.', '_'), ukey: _key};		
	
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;	
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
	return this.http_post_async("/api/get?data", JSON.stringify(_cmd), _callback);
}
//-----
_User.GetDataByKeyAsync = _User.getDataByKeyAsync = function(_src, _key, _callback, _isCompressed)
{                   
	return this.GetDataByKeyAsyncByUser(this.GetSimulatedUser(), _src, _key, _callback, _isCompressed);
}
//-----
_User.GetImageByAuthAsync = _User.getImageByAuthAsync = function(url, account, password, _callback)
{               
	let _cmd = { img_url: url, encode: 'base64', "account": account, "password": password};
	return this.http_post_async("/api/getimage?", JSON.stringify(_cmd), _callback);
}
//-----
_User.GetImageByAuth = _User.getImageByAuth = function(url, account, password)
{               
	let _cmd = { img_url: url, encode: 'base64', "account": account, "password": password};
	return this.http_post_sync("/api/getimage?", JSON.stringify(_cmd));
}
//-----
_User.GetImageAsync = _User.getImageAsync =  function(url, _callback, _obj)
{               	
	let _cmd = { img_url: url, encode: 'base64'};
	return this.http_post_async("/api/getimage?", JSON.stringify(_cmd), _callback, _obj);
}
//-----
_User.GetImage = _User.getImage =  function(url)
{               	
	let _cmd = { img_url: url, encode: 'base64'};
	return this.http_post_sync("/api/getimage?", JSON.stringify(_cmd));
}
//-----
//Share Data
//-----
_User.readShareDataByKey = function(_user, _type, _key)
{               
	if(_key != null)
	{		
		let _cmd = { ukey: _key };
		
		if(_user == null) _user = _GlobalDataOwner;
		if(_user != null) _cmd.user = _user;
		if(_type != null) _cmd.type = _type;
		
		return this.http_post_sync("/auth/share_data?method=read", JSON.stringify(_cmd));
	}
	
	return null;
}
//-----
_User.readShareDataByKeyAsync = function(_user, _type, _key, _id, _callback)
{                                                       
	if(_key != null)
	{		
		let _cmd = { ukey: _key };
		
		if(_user == null) _user = _GlobalDataOwner;
		if(_user != null) _cmd.user = _user;
		if(_type != null) _cmd.type = _type;
		
		this.http_post_async("/auth/share_data?method=read", JSON.stringify(_cmd), function(_data){ _callback(_id, _data); });
	}
}
//-----
_User.readShareDataByTimeAsync = function(_user, _type, _start_time, _end_time, _callback)
{               
	if(_start_time != null && _end_time != null)
	{		
		let _cmd = { start_time: _start_time, end_time: _end_time };
		
		if(_user == null) _user = _GlobalDataOwner;
		if(_user != null) _cmd.user = _user;
		if(_type != null) _cmd.type = _type;
		
		return this.http_post_async("/auth/share_data?method=read", JSON.stringify(_cmd), _callback);
	}
	
	return null;
}
//-----
_User.readShareDataByTime = function(_user, _type, _start_time, _end_time)
{               
	if(_start_time != null && _end_time != null)
	{		
		let _cmd = { start_time: _start_time, end_time: _end_time };
		
		if(_user == null) _user = _GlobalDataOwner;
		if(_user != null) _cmd.user = _user;
		if(_type != null) _cmd.type = _type;
		
		return this.http_post_sync("/auth/share_data?method=read", JSON.stringify(_cmd));
	}
	
	return null;
}
//-----
_User.readLastShareData = function(_user, _type, _count)
{               
	if(_count != null)
	{		
		let _cmd = { count: _count };
		
		if(_user == null) _user = _GlobalDataOwner;
		if(_user != null) _cmd.user = _user;
		if(_type != null) _cmd.type = _type;
		
		return this.http_post_sync("/auth/share_data?method=read&lastrec="+_count, JSON.stringify(_cmd));
	}
	
	return null;
}
//-----
_User.readLastShareDataAsync = function(_user, _type, _count, _callback)
{               
	if(_count != null)
	{		
		let _cmd = { count: _count };
		
		if(_user == null) _user = _GlobalDataOwner;
		if(_user != null) _cmd.user = _user;
		if(_type != null) _cmd.type = _type;
		
		return this.http_post_async("/auth/share_data?method=read&lastrec="+_count, JSON.stringify(_cmd), function(_data){ _callback(_user, _data); });
	}
	
	return null;
}
//-----
_User.writeShareDataByKey = function(_user, _type, _key, _data)
{               
	if(_key != null && _data != null)
	{		
		let _cmd = { ukey: _key, values: [_data] };
		
		if(_user == null) _user = _GlobalDataOwner;
		if(_user != null) _cmd.user = _user;
		if(_type != null) _cmd.type = _type;
			
		if(this.http_post_sync("/auth/share_data?method=write", JSON.stringify(_cmd)) != null)
			return true;
	}
	
	return false;
}
//-----
_User.removeShareDataByKey = function(_user, _type, _key)
{               
	if(_key != null)
	{		
		let _cmd = { ukey: _key };
		
		if(_user == null) _user = _GlobalDataOwner;
		if(_user != null) _cmd.user = _user;
		if(_type != null) _cmd.type = _type;
			
		if(this.http_post_sync("/auth/share_data?method=remove", JSON.stringify(_cmd)) != null)
			return true;
	}
	
	return false;
}
//----- 
//Update data
//-----
_User.updateDataByKeyByUser = _User.UpdateDataByKeyByUser = function(_user, _src, _key, _data)
{               
	if(_src != null && _key != null)
	{				
		let _cmd = { uuid: _src.replace('.', '_'), ukey: _key, values: [_data]};
		
		if(_user == null) _user = _GlobalDataOwner;
		if(_user != null) _cmd.user = _user;	
		
		if(this.http_post_sync("/api/put?data", JSON.stringify(_cmd)) != null)
			return true;
	}
	
	return false;
}
//-----
_User.updateDataByKey = _User.UpdateDataByKey = function(_src, _key, _data)
{               	
	return this.updateDataByKeyByUser(this.GetSimulatedUser(), _src, _key, _data);
}
//----- 
//Delete data
//----- 
_User.delData = _User.DelData = function(_src, _stime, _etime, _callback)
{                  
	if(_callback != null) return this.delDataAsync(_src, _stime, _etime, _callback);
	
	let _cmd = { uuid: _src.replace('.', '_'), timestamp: _stime+'~'+_etime };        
	let _simUser = this.GetSimulatedUser();
	
	if(_simUser != null) _cmd.user = _simUser;	
	return this.http_post_async("/api/del?data", JSON.stringify(_cmd));
}
//----- 
_User.delDataAsync = _User.DelDataAsync = function(_src, _stime, _etime, _callback)
{                  
	let _cmd = { uuid: _src.replace('.', '_'), timestamp: _stime+'~'+_etime };        
	let _simUser = this.GetSimulatedUser();
	
	if(_simUser != null) _cmd.user = _simUser;	
	return this.http_post_async("/api/del?data", JSON.stringify(_cmd), _callback);
}
//----- 
_User.delDataByKeyByUser = _User.DelDataByKeyByUser = function(_user, _src, _ukey, _callback)
{                  	
	if(_callback != null) return this.delDataByKeyAsync(_src, _ukey, _callback);
	
	let _cmd = { uuid: _src.replace('.', '_'), ukey: _ukey };   	
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;	
	
    return this.http_post_sync("/api/del?data", JSON.stringify(_cmd));
}
//----- 
_User.delDataByKey = _User.DelDataByKey = function(_src, _ukey, _callback)
{                  	
    return this.delDataByKeyByUser(this.GetSimulatedUser(), _src, _ukey, _callback);
}
//----- 
_User.delDataByKeyAsync = _User.DelDataByKeyAsync = function(_src, _ukey, _callback)
{                  	
	let _cmd = { uuid: _src.replace('.', '_'), ukey: _ukey };        
	let _simUser = this.GetSimulatedUser();
	
	if(_simUser != null) _cmd.user = _simUser;	
	
    return this.http_post_async("/api/del?data", JSON.stringify(_cmd), _callback);
}
//----- 
//Show multi-day report
//-----       
_User.getDaysReport = _User.GetDaysReport = function(_src, _start_day, _end_day, _callback)
{  
	return this.GetDaysReportByUser(this.GetSimulatedUser(), _src, _start_day, _end_day, _callback);
}
//-----       
_User.getDaysReportAsync = _User.GetDaysReportAsync = function(_src, _start_day, _end_day, _callback)
{ 
  return this.getDaysReportAsyncByUser(this.GetSimulatedUser(), _start_day, _end_day, _callback);
}
//----- 
//Show one day report (by 24 hours)
//----- 
_User.getOneDayReport = _User.GetOneDayReport = function(_src, _day, _callback)
{                     
    return this.getDaysReportByUser(this.GetSimulatedUser(), _src, _day, null, _callback);
}
//----- 
//Show multi-daya report By User
//-----       
_User.getDaysReportByUser = _User.GetDaysReportByUser = function(_user, _src, _start_day, _end_day, _callback, _isCompressed)
{
	if(_src == null) return null;
	if(_callback != null) return this.getDaysReportAsyncByUser(_user, _src, _start_day, _end_day, _callback);
		
	let _cmd = { uuid: _src = _src.replace('.', '_') };  
	
	if(_start_day) _start_day = new Date(_start_day).GetUTC();
	if(_end_day) _end_day = new Date(_end_day).GetUTC();
	if(_end_day == null) _cmd.day = _start_day; else _cmd.day = _start_day+"~"+_end_day;
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
	return this.http_get_sync("/api/getbyday?"+"src="+_src+"&day="+_cmd.day, JSON.stringify(_cmd));
}
//-----       
_User.getDaysReportAsyncByUser = _User.GetDaysReportAsyncByUser = function(_user, _src, _start_day, _end_day, _callback, _isCompressed)
{   
	let _cmd = { uuid: _src.replace('.', '_') };  
	
	if(_start_day) _start_day = new Date(_start_day).GetUTC();
	if(_end_day) _end_day = new Date(_end_day).GetUTC();	
	if(_end_day == null) _cmd.day = _start_day; else _cmd.day = _start_day+"~"+_end_day;
	if(_user == null) _user = _GlobalDataOwner;
	if(_user != null) _cmd.user = _user;
	if(_isCompressed != null) _cmd.compressed = _isCompressed;
	
	return this.http_get_async("/api/getbyday?"+"src="+_src+"&day="+_cmd.day, JSON.stringify(_cmd), _callback);
}
//----- 
//Show one day report By User(by 24 hours)
//----- 
_User.getOneDayReportByUser = _User.GetOneDayReportByUser = function(_user, _src, _day, _callback, _isCompressed)
{                     
    return this.getDaysReportByUser(_user, _src, _day, _day, _callback, _isCompressed);
}
//-----
// Admin/User Tools
//---
_User.ConsoleCMD = function(_cmd, _callback)
{
	return  this.http_post_async("/sys/command?default", _cmd, _callback);	
}
//-----
_User.GetUserList = function(_userfilter)
{        
	return this.http_get_sync("/sys/userlist?user="+_userfilter);
}
//-----
_User.GetRecentUserList = function()
{        
	return this.http_get_sync("/auth/get_recent_login?");
}
//---
_User.GetUserInfo = function(_user)
{ 	    
	let _info = null;
  
	if(_user != null) _info = this.http_get_sync("/api/usrinfo?method=read&user="+_user);	
	if(_info == null) _info = this.http_get_sync("/api/usrinfo?method=read");
	
	if(_info != null && _info.length > 0)
    {		
		let _payload = JSON.parse(_info);
		
		//if(_payload.active != true) return null;
	    if(_payload.note == "DISABLE") return null;
    }
	
    return _info;
}
//---
_User.GetUserInfoAsync = function(_user, _key, _callback)
{ 	    
	if(_user != null) return this.http_post_async("/api/usrinfo?method=read&user="+_user, '', function(_payload) { let _info = JSON.parse(_payload); if(_info.note == "DISABLE") _payload = null; _callback(_key, _payload); });	
	return this.http_get_sync("/api/usrinfo?method=read", '', function(_payload) { let _info = JSON.parse(_payload); if(_info.note == "DISABLE") _payload = null; _callback(_key, _payload); });	
}
//---
_User.UpdateUser = function(_user, _info)
{ 	    	
	let _data=null;
	
	if(_user != null) _data=this.http_post_sync("/api/usrinfo?method=write&user="+_user, JSON.stringify(_info));   		
	if(_data == null) _data=this.http_post_sync("/api/usrinfo?method=write", JSON.stringify(_info));   
	
    if(_data != null) return true;    
    return false;
}
//---
_User.UpdateUserAdv = function(_user, _info, _fully)
{ 	    	
	let _data=null;
	if(_user != null) _data=this.http_post_sync("/api/usrinfo?method=write&user="+_user+"&full_apply="+_fully, JSON.stringify(_info));  
	if(_data == null) _data=this.http_post_sync("/api/usrinfo?method=write", JSON.stringify(_info));   
	
    if(_data != null) return true;    
    return false;
}
//---
_User.DisableUser = function(_id)
{ 	    
    if(_id == null) return false;        
    
    return this.UpdateUser(_id, { id: _id, note: "DISABLE" });
}
//---
_User.DelUser = function(_user)
{ 	    
	let _data=null;
	
	if(_user != null) _data=this.http_get_sync("/sys/userinfo?method=del&user="+_user);
	if(_data == null) _data=this.http_get_sync("/api/usrinfo?method=del");
	
	if(_data != null) return true;    
    return false;
}
//---
_User.GetMapInfo = function()
{ 	    	
	return this.http_get_sync("/api/usrinfo?method=mapinfo");
}
//---
_User.UpdateMapInfo = function(_info)
{ 	    
	return this.http_post_sync("/api/usrinfo?method=updatemap", JSON.stringify(_info));
}
//---
_User.DelMapInfo = function(_info)
{ 	    
	return this.http_post_sync("/api/usrinfo?method=delmap", JSON.stringify(_info));
}
//---
//DB
//---
_User.GetDBStatus = function(_callback)
{        
	if(_callback == null) return this.http_get_sync("/db/get_status?");
	
	this.http_get_async("/db/get_status?", _callback);
}
//---
_User.GetDBLastStatus = function(_dbType, _count, _callback)
{        
	if(_callback == null) return this.http_get_sync("/db/get_last_status?count="+_count+"&type="+_dbType);
	
	this.http_get_async("/db/get_last_status?count="+_count+"&type="+_dbType, function(_data){ _callback(_dbType, _data); });
}
//---
// Servers
//---
_User.GetServerList = function(_isForceUpdate)
{        
	return this.http_get_sync(_isForceUpdate ? "/sys/serverlist?status=all&update=1" : "/sys/serverlist?status=all");
}
//---
_User.GetServerDetail = function()
{ 	    
	return this.http_get_sync("/sys/serverinfo?method=all");
}
//----- 
_User.getOneDaySystemLog = function(_day, _callback)
{                     
	let _respone = this.http_get_sync("/sys/getsyslog?day="+_day);
	if(_callback != null) _callback(_respone);
    return _respone;
}  
//----- 
_User.addSystemLog = function(_payload)
{                     
	_payload.user = this.GetCurrentUserID();
	_payload.level = this.GetCurrentUserGroup();
	
	return this.http_post_sync("/sys/addsyslog?", JSON.stringify(_payload));
}
_User.getSystemNotifyMail = function(_callback)
{                     
	this.http_get_async("/sys/notify_mail?method=get", _callback);
}
//----- 
_User.setSystemNotifyMail = function(_payload, _callback)
{                     
	this.http_post_async("/sys/notify_mail?method=set", JSON.stringify(_payload), _callback);
}
//----- 
_User.CreateQRCode = function(_data, _callback)
{ 
	//this.http_post_async("/test/qrcode?write="+_data, JSON.stringify({ data : _data }), _callback);
	this.http_post_async("/test/qrcode?write="+_data, "", _callback);
}
//----- 
_User.PushTCPData = _User.pushTCPData = function(_url, _port, _data)
{                     
	let _cmd = { url: _url, port: _port, values: [_data] };        
	return this.http_post_sync("/sys/push_tcp_data?none", JSON.stringify(_cmd));
}
//---
// Events & Data
//---
_User.SubscribeByMQTT_Adv = function(_topic, _callback_OnMessage, _callback_OnConnChange)
{ 
	if(_topic == null || _callback_OnMessage == null)
		return null;
	
	let _token = this.WasLogin();
	let _id = _token+'@'+new Date().getTime()%1000;//_account+'_'+(new Date().getTime()%1000);	
	let _url = window.location.protocol == "https:" ? 'wss://'+window.location.hostname+':8884/' : 'ws://'+window.location.hostname+':1884/';
	let _client = new Paho.MQTT.Client(_url, _id);
		
	_topic = _topic.replace('*/', '+/');
	_topic = _topic.replace('/*', '/#');
			
	_client.parent = this;
	_client.topic = _topic;
	_client.OnConnChangeCallback = _callback_OnConnChange;
	_client.OnMessageCallback = _callback_OnMessage;
				
	this._SubscribeClients.push(_client);		
	
	_client.onConnected = function(_isReconnect)
	{
		try
		{	
			
		}
		catch(_e){}
	}
			
	_client.onMessageArrived = function(_msg)
	{
		try
		{						
			if(_client.OnMessageCallback != null && _msg != null)
				_client.OnMessageCallback({ data: _msg.payloadString });
		}
		catch(_e){}
	}
	
	_client.onConnectionLost = function(_item)
	{			
		try
		{
			if(_client.OnConnChangeCallback != null) _client.OnConnChangeCallback(false);			
			if(_client.clientId != null) _client.parent.CloseSubscribClient(_client.clientId);
			if(_client.errorCode == 8) return; // duplicate connect		
			
			console.log("MQTT Connection Lost");
		}
		catch(_e){}			
	}		

	let _mqtt_account = this._account.replace('@','-').replace('.','-').replace('_','-');

	_client.connect({userName: _mqtt_account, password: _token, timeout: 8, reconnect: true,
					onSuccess: function()
					{						
						if(_client != null) _client.subscribe(_client.topic);		
						if(_client.OnConnChangeCallback != null) _client.OnConnChangeCallback(true);		
					},
					onFailure:function()
					{
						if(_client != null) _client.onConnectionLost(null);
					}});
	
	return _client.clientId;
}
//---
_User.SubscribeByMQTT = function(_type, _src_id, _callback_OnMessage, _callback_OnConnChange)
{ 
	if(_type == null || _src_id == null || _callback_OnMessage == null) return null;
	
	let _simUser = null;	
	
	if(_simUser == null) _simUser = (this.GetCurrentUserProj() != "none") ? this.GetCurrentUserProj() : null;
	if(_simUser == null) _simUser = _GlobalDataOwner;	

	let _account = (_simUser != null && _simUser.length > 0) ? _simUser : this._account;
	
	_account = _account.replace('@','-').replace('.','-').replace('_','-');
	
	let _topic = _account+'/'+_type+'/'+_src_id.replace('.', '_');
	
	return this.SubscribeByMQTT_Adv(_topic, _callback_OnMessage, _callback_OnConnChange);
}
//---
_User.CloseSubscribClient = function(_id)
{
	//
}
//---
_User.SubscribeData = function(_id, _callback_OnMessage, _callback_OnConnChange)
{	
	try
	{
		return this.SubscribeByMQTT('data', _id, _callback_OnMessage, _callback_OnConnChange);
	}
	catch(_e){}
	return null;
}
//---
_User.SubscribeEvents = function(_id, _callback_OnMessage, _callback_OnConnChange)
{	
	try
	{
		return this.SubscribeByMQTT('event', _id, _callback_OnMessage, _callback_OnConnChange);
	}
	catch(_e){}
	return null;
}
//---
_User.RequestNewPassword = function(_user)
{	
	try
	{
		if(_user != null) 
		{
			let xmlHttp = new XMLHttpRequest();   

	   		xmlHttp.open("GET", "/auth/req_new_pwd?usr="+_user, false);
	    	xmlHttp.send();         
	    	        
	    	if(xmlHttp.readyState == 4 && xmlHttp.status == 200)
	    	{
	    		let data = xmlHttp.responseText;
	    		if(_data != null)  return true;
	    	}
		}

	    return false;
	}
	catch(_e){}
	return null;
}
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
_User.GetProjectList = _User.apiGetProjectList = function()
{	
	try
	{
		let _xmlHttp = new XMLHttpRequest();   

		_xmlHttp.open("GET", "/api/project_list", false);
		_xmlHttp.send();

		if(_xmlHttp.readyState == 4 && _xmlHttp.status == 200)
		{
			let _data = _xmlHttp.responseText;
			if(_data != null && _data.length > 0) {
				return _data
			}
		}
		
	    return "";
	}
	catch(_e){}
	return "";
}
//---
_User.GetProjectID = _User.apiGetProjectID = function()
{	
	return this.http_get_sync("/api/project?method=id");
}
//---
_User.GetUserAvailableProjec = _User.apiGetUserAvailableProject = function()
{	
	return this.http_get_sync("/api/userinfo?method=project_list");
}
//---
_User.GetUserAvailableFunction = _User.apiGetUserAvailableFunction = function(_type)
{	
	// type: sys / adv / norm
	return this.http_get_sync("/api/userinfo?method=function_list&target="+_type);
}
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
_User.projGetProjectList = function()
{
	return this.http_get_sync("/proj/proj?method=list");
}
//---
_User.projCreateProject = function(_id, _name, _owner, _url, _note)
{
	let _cmd = {};
	_cmd.id = _id;
	_cmd.name = _name;
	_cmd.owner = _owner;
	_cmd.url = _url;
	_cmd.note = _note;

	let _resp = this.http_post_sync("/proj/proj?method=create", JSON.stringify(_cmd));
	return (_resp != null && _resp == "Ok");
}
//---
_User.projDeleteProject = function(_proj_id)
{
	let _cmd = { id: _proj_id };
	let _resp = this.http_post_sync("/proj/proj?method=delete", JSON.stringify(_cmd));
	return (_resp != null && _resp == "Ok");
}
//---
_User.projReadProjectInfo = function(_proj_id)
{
	let _cmd = { id: _proj_id };
	return this.http_post_sync("/proj/proj?method=read", JSON.stringify(_cmd));
}
//---
_User.projUpdateProjectInfo = function(_proj_id, _info, full_apply)
{
	_info.id = _proj_id;
	let _resp = this.http_post_sync("/proj/proj?method=write&full_apply="+full_apply, JSON.stringify(_info));
	return (_resp != null && _resp == "Ok");
}
//---
_User.projGetProjectUesrList = function(_proj_id)
{
	let _cmd = { id: _proj_id };
	return this.http_post_sync("/proj/user?method=list", JSON.stringify(_cmd));
}
//---
_User.projGetProjectUesrInfo = function(_proj_id, _user_id)
{
	let _cmd = { id: _proj_id, user: _user_id };
	return this.http_post_sync("/proj/user?method=read", JSON.stringify(_cmd));
}
//---
_User.projAddProjectUesr = function(_proj_id, _user_id)
{
	let _cmd = { id: _proj_id, user: _user_id };
	return this.http_post_sync("/proj/user?method=add", JSON.stringify(_cmd));
}
//---
_User.projDelProjectUesr = function(_proj_id, _user_id)
{
	let _cmd = { id: _proj_id, user: _user_id };
	return this.http_post_sync("/proj/user?method=del", JSON.stringify(_cmd));
}
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
_User.projGetFunctionList = function(_type)
{
	return this.http_get_sync("/proj/func?method=list&type="+_type);
}
//---
_User.projCreateFunction = function(_type, _id, _name, _url, _permission, _note)
{
	// type: sys / adv / norm / app

	let _cmd = {};
	if(_id   != null)	_cmd.id   = _id;
	if(_name != null)	_cmd.name = _name;
	if(_url  != null)	_cmd.url  = _url;
	if(_note != null)	_cmd.note = _note;
	if(_permission != null) _cmd.permission = _permission;

	let _resp = this.http_post_sync("/proj/func?method=create&type="+_type, JSON.stringify(_cmd));
	return (_resp == "Ok");
}
//---
_User.projDeleteFunction = function(_type, _id)
{
	let _cmd = {};
	if(_id != null)	_cmd.id = _id;

	let _resp = this.http_post_sync("/proj/func?method=delete&type="+_type, JSON.stringify(_cmd));
	return (_resp == "Ok");
}
//---
_User.projUpdateFunction = function(_type, _id, _name, _url, _permission, _note)
{
	let _cmd = {};
	if(_id   != null)	_cmd.id   = _id;
	if(_name != null)	_cmd.name = _name;
	if(_url  != null)	_cmd.url  = _url;
	if(_note != null)	_cmd.note = _note;
	if(_permission != null) _cmd.permission = _permission;

	let _resp = this.http_post_sync("/proj/func?method=update&type="+_type, JSON.stringify(_cmd));
	return (_resp == "Ok");
}
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
_User.projGetFunctionGroupList = function()
{
	return this.http_get_sync("/proj/funcgroup?method=list");
}
//---
_User.projCreateFunctionGroup = function(_id, _name, _note, _list_normal, _list_proj, _list_system, _list_app)
{
	let _cmd = {};
	if(_id   != null)	_cmd.id   = _id;
	if(_name != null)	_cmd.name = _name;
	if(_note != null)	_cmd.note = _note;
	if(_list_normal != null)	_cmd.list_normal = _list_normal;
	if(_list_proj   != null)	_cmd.list_proj = _list_proj;
	if(_list_system != null)	_cmd.list_system = _list_system;
	if(_list_app 	!= null)	_cmd.list_app = _list_app;

	let _resp = this.http_post_sync("/proj/funcgroup?method=create", JSON.stringify(_cmd));
	return (_resp == "Ok");
}
//---
_User.projDeleteFunctionGroup = function(_id)
{
	let _cmd = {};
	if(_id != null)	_cmd.id = _id;

	let _resp = this.http_post_sync("/proj/funcgroup?method=delete", JSON.stringify(_cmd));
	return (_resp == "Ok");
}
//---
_User.projUpdateFunctionGroup = function(_id, _name, _note, _list_normal, _list_proj, _list_system, _list_app)
{
	let _cmd = {};
	if(_id   != null)	_cmd.id   = _id;
	if(_name != null)	_cmd.name = _name;
	if(_note != null)	_cmd.note = _note;
	if(_list_normal != null)	_cmd.list_normal = _list_normal;
	if(_list_proj   != null)	_cmd.list_proj = _list_proj;
	if(_list_system != null)	_cmd.list_system = _list_system;
	if(_list_app 	!= null)	_cmd.list_app = _list_app;

	let _resp = this.http_post_sync("/proj/funcgroup?method=update", JSON.stringify(_cmd));
	return (_resp == "Ok");
}
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
_User.Base64ToBlob = function(_base64)
{
    let _binaryString =  window.atob(_base64);
    let _binaryLen = _binaryString.length;

    let _ab = new ArrayBuffer(_binaryLen);
    let _ia = new Uint8Array(_ab);
    
    for (let i = 0; i < _binaryLen; i++)
       _ia[i] = _binaryString.charCodeAt(i);

    let _bb = new Blob([_ab]);
    
    _bb.lastModifiedDate = new Date();
    _bb.name = "onfly.zip";
    _bb.type = "zip";

    return _bb;
}
//---
_User.unZipData = _User.UnzipData = function(_data, _callback, _extItem)
{
	if(_callback == null) return;
	if(_data == null) { _callback(null, null); return; }
	
	let _zip = new zip.ZipReader(new zip.BlobReader(this.Base64ToBlob(_data)));		
	if(_zip)
	{
		try
		{
			_zip.getEntries().then(function(_entries)
			{				
				_entries[0].getData(new zip.TextWriter()).then(function(_unzipData)
				{
					_callback(_unzipData, _entries[0], _extItem);
				});				
			});		
		}
		catch(_e){}
		
		_zip.close();
	}	
	else
		_callback(null, null, _extItem);
}
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
_User.dbGetTableList = function(_callback)
{
	return this.http_get_async("/db/table?method=list", _callback);
}
//---
_User.dbRemoveTable = function(_table, _family)
{
	let _params = {};
	
	_params.table = _table;
	_params.family = _family;

	return this.http_post_async("/db/table?method=del", JSON.stringify(_params));
}
//---
_User.dbRemoveDataByKey = function(_table, _family, _column, _ukey)
{
	let _params = {};
	_params.table = _table;
	_params.family = _family;
	_params.column = _column;
	_params.ukey = _ukey;

	return this.http_post_async("/db/data?method=del&type=bykey", JSON.stringify(_params));
}
//---
_User.dbGetDataByKey = function(_table, _family, _column, _ukey, _callback)
{
	let _params = {};
	_params.table = _table;
	_params.family = _family;
	_params.column = _column;
	_params.ukey = _ukey;

	return this.http_post_async("/db/data?method=get&type=bykey", JSON.stringify(_params), _callback);
}
//---
_User.dbGetDataByTime = function(_table, _family, _column, _start_time, _end_time, _callback)
{
	let _params = {};
	_params.table = _table;
	_params.family = _family;
	_params.column = _column;
	_params.start = _start_time;
	_params.end = _end_time;

	return this.http_post_async("/db/data?method=get&type=bytime", JSON.stringify(_params), _callback);
}
//---
_User.dbGetDataByCount = function(_table, _family, _column, _count, _callback)
{
	let _params = {};
	
	_params.table = _table;
	_params.family = _family;
	_params.column = _column;
	_params.count = _count;

	return this.http_post_async("/db/data?method=get&type=bycount", JSON.stringify(_params), _callback);
}
//---
_User.dbPutData = function(_table, _family, _column, _ukey, _value)
{
	let _params = {};
	_params.table = _table;
	_params.family = _family;
	_params.column = _column;
	_params.ukey = _ukey;
	_params.value = _value;

	return this.http_post_async("/db/data?method=put", JSON.stringify(_params), _callback);
}
//-----------------------------------------------------------------------
//
//-----------------------------------------------------------------------
