package com.mars.cloud;

import android.graphics.Bitmap;

import org.json.JSONArray;
import org.json.JSONObject;

//----------------------------------------------------------------------------------------
public class AbstractObject
{
    //----------------------------------------------------------------------------------------
    public static abstract class ICallback
    {
        protected long _ID = System.currentTimeMillis();
    }
    //----------------------------------------------------------------------------------------
    public static abstract class IStringDataCallback extends ICallback
    {
        public abstract void OnData(String _payload);
    }
    //----------------------------------------------------------------------------------------
    public static abstract class IByteArrayDataCallback extends ICallback
    {
        public abstract void OnData(byte[] _payload);
    }
    //----------------------------------------------------------------------------------------
    public static abstract class IJSONObjectCallback extends ICallback
    {
        public abstract boolean OnResult(JSONObject _data);
    }
    //----------------------------------------------------------------------------------------
    public static abstract class IFwUpdateCallback extends ICallback
    {
        protected abstract void OnResults(JSONObject _data);
    }
    //----------------------------------------------------------------------------------------
    public static abstract class IJSONArrayCallback extends ICallback
    {
        public abstract boolean OnResults(JSONArray _data);
    }
    //----------------------------------------------------------------------------------------
    public static abstract class IDataItem
    {
        protected String _DataItemType = "null";
        public abstract boolean LoadFromJSONObject(JSONObject _item);
        public abstract JSONObject ToJSONObject();
        public abstract boolean Equals(IDataItem _src);
        public abstract boolean Assign(IDataItem _src);
        public abstract boolean IsEmpty();
    }
    //----------------------------------------------------------------------------------------
}
//----------------------------------------------------------------------------------------