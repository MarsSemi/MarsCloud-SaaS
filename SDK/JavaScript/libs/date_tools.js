//-----
// for Date class
//-----
Date.prototype.YMD = function(_spliter)
{
	_spliter = _spliter == null ? '-' : _spliter;	
	
  let _y = this.getFullYear().toString();
  let _m = (this.getMonth()+1).toString();
  let _d = this.getDate().toString();
  return _y + _spliter + (_m[1]?_m:"0"+_m[0]) + _spliter + (_d[1]?_d:"0"+_d[0]);
}
//-----
Date.prototype.YMDT = function(_spliter)
{
  _spliter = _spliter == null ? '-' : _spliter;
	
  let _y = this.getFullYear().toString();
  let _m = (this.getMonth()+1).toString();
  let _d = this.getDate().toString();
  let _hour = this.getHours().toString();
  let _min = this.getMinutes().toString();
  let _sec = this.getSeconds().toString();
  
  return _y+_spliter+(_m[1]?_m:"0"+_m[0])+_spliter+(_d[1]?_d:"0"+_d[0])+" "+(_hour[1]?_hour:"0"+_hour[0])+":" +(_min[1]?_min:"0"+_min[0])+":" +(_sec[1]?_sec:"0"+_sec[0]);
}
//-----
Date.prototype.YMDT2 = function(_spliter)
{
	_spliter = _spliter == null ? '-' : _spliter;
	
  let _y = this.getFullYear().toString();
  let _m = (this.getMonth()+1).toString();
  let _d = this.getDate().toString();
  let _hour = this.getHours().toString();
  let _min = this.getMinutes().toString();
  let _sec = this.getSeconds().toString();
  
  return _y+_spliter+(_m[1]?_m:"0"+_m[0])+_spliter+(_d[1]?_d:"0"+_d[0])+" "+(_hour[1]?_hour:"0"+_hour[0])+":" +(_min[1]?_min:"0"+_min[0]);
}
//-----
Date.prototype.YMDT_DateTimePicker = function()
{
  let _y = this.getFullYear().toString();
  let _m = (this.getMonth()+1).toString();
  let _d = this.getDate().toString();
  let _hour = this.getHours().toString();
  let _min = this.getMinutes().toString();
  let _sec = this.getSeconds().toString();
  
  return _y+"-"+(_m[1]?_m:"0"+_m[0])+"-"+(_d[1]?_d:"0"+_d[0])+"T"+(_hour[1]?_hour:"0"+_hour[0])+":" +(_min[1]?_min:"0"+_min[0])+":" +(_sec[1]?_sec:"0"+_sec[0]);
}
//-----
Date.prototype.MD = function()
{
  let _m = (this.getMonth()+1).toString();
  let _d = this.getDate().toString();
  return (_m[1]?_m:"0"+_m[0]) + "/" + (_d[1]?_d:"0"+_d[0]);
}
//-----
Date.prototype.MDT = function()
{
	let _m = (this.getMonth()+1).toString();
	let _d = this.getDate().toString();
	let _hour = this.getHours().toString();
	let _min = this.getMinutes().toString();
	let _sec = this.getSeconds().toString();
	  
	return (_m[1]?_m:"0"+_m[0]) + "/" + (_d[1]?_d:"0"+_d[0])+" "+(_hour[1]?_hour:"0"+_hour[0])+":" +(_min[1]?_min:"0"+_min[0])+":" +(_sec[1]?_sec:"0"+_sec[0]);
}
//-----
Date.prototype.MDT2 = function()
{
	let _m = (this.getMonth()+1).toString();
	let _d = this.getDate().toString();
	let _hour = this.getHours().toString();
	let _min = this.getMinutes().toString();
	let _sec = this.getSeconds().toString();
	  
	return (_m[1]?_m:"0"+_m[0]) + "/" + (_d[1]?_d:"0"+_d[0])+" "+(_hour[1]?_hour:"0"+_hour[0])+":" +(_min[1]?_min:"0"+_min[0]);
}
//-----
Date.prototype.YM = function()
{
	let _y = this.getFullYear().toString();
	let _m = (this.getMonth()+1).toString();
	  
	return _y+"-"+(_m[1]?_m:"0"+_m[0]);
}
//-----
Date.prototype.Time = function()
{
  let _y = this.getFullYear().toString();
  let _m = (this.getMonth()+1).toString();
  let _d = this.getDate().toString();
  let _hour = this.getHours().toString();
  let _min = this.getMinutes().toString();
  let _sec = this.getSeconds().toString();
  
  return (_hour[1]?_hour:"0"+_hour[0])+":" +(_min[1]?_min:"0"+_min[0])+":" +(_sec[1]?_sec:"0"+_sec[0]);
}
//-----
Date.prototype.GetUTC = function()
{
  return this.getTime();
}
//-----
Date.prototype.GetTimezoneOffset = function()
{
  return this.getTimezoneOffset()*60000;
}
//-----
Date.prototype.GetLocalTime = function()
{
  return this.getTime()-this.GetTimezoneOffset();
}
//-----
Date.prototype.addSeconds = function(_s)
{
	return new Date(this.getTime() + (_s*1000)); 
}
//-----
Date.prototype.addMins = function(_m)
{
	return new Date(this.getTime() + (_m*60*1000)); 
}
//-----
Date.prototype.addHours = function(_h)
{
	return new Date(this.getTime() + (_h*60*60*1000)); 
}
//-----
Date.prototype.addDays = function(_d)
{
	return new Date(this.getTime() + (_d*24*60*60*1000));   
}
//-----
