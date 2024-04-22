package com.mars.simpleclient;

import android.os.Bundle;

import com.google.android.material.snackbar.Snackbar;

import androidx.appcompat.app.AppCompatActivity;

import android.view.View;

import androidx.navigation.NavController;
import androidx.navigation.Navigation;
import androidx.navigation.ui.AppBarConfiguration;
import androidx.navigation.ui.NavigationUI;

import com.mars.cloud.AbstractObject;
import com.mars.cloud.MarsClient;
import com.mars.cloud.Tools;
import com.mars.simpleclient.databinding.ActivityMainBinding;

import android.view.Menu;
import android.view.MenuItem;

import org.eclipse.paho.client.mqttv3.IMqttDeliveryToken;
import org.eclipse.paho.client.mqttv3.MqttCallback;
import org.eclipse.paho.client.mqttv3.MqttMessage;
import org.json.JSONObject;

import java.util.Timer;
import java.util.TimerTask;

public class MainActivity extends AppCompatActivity
{
    //------------------------------------------------------------
    //
    //------------------------------------------------------------
    private AppBarConfiguration appBarConfiguration;
    private ActivityMainBinding binding;
    private MarsClient _Client = null;
    public static Timer _InitTimer = new Timer();
    //------------------------------------------------------------
    @Override
    protected void onCreate(Bundle savedInstanceState)
    {
        super.onCreate(savedInstanceState);

        try
        {
            binding = ActivityMainBinding.inflate(getLayoutInflater());
            setContentView(binding.getRoot());

            setSupportActionBar(binding.toolbar);

            NavController navController = Navigation.findNavController(this, R.id.nav_host_fragment_content_main);
            appBarConfiguration = new AppBarConfiguration.Builder(navController.getGraph()).build();
            NavigationUI.setupActionBarWithNavController(this, navController, appBarConfiguration);

            binding.fab.setOnClickListener(new View.OnClickListener()
            {
                @Override
                public void onClick(View view)
                {
                    Snackbar.make(view, "Replace with your own action", Snackbar.LENGTH_LONG)
                            .setAction("Action", null).show();
                }
            });

            _InitTimer.schedule(new TimerTask(){ public void run() { InitMarsClient(); }}, 100);
            _InitTimer.schedule(new TimerTask(){ public void run() { MarsClientTest(); }}, 1000);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //------------------------------------------------------------
    @Override
    public boolean onCreateOptionsMenu(Menu menu)
    {
        try
        {
            getMenuInflater().inflate(R.menu.menu_main, menu);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return true;
    }
    //------------------------------------------------------------
    @Override
    public boolean onOptionsItemSelected(MenuItem item)
    {
        try
        {
            int id = item.getItemId();

            //noinspection SimplifiableIfStatement
            if (id == R.id.action_settings)
            {
                return true;
            }

            return super.onOptionsItemSelected(item);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //------------------------------------------------------------
    @Override
    public boolean onSupportNavigateUp()
    {
        try
        {
            NavController navController = Navigation.findNavController(this, R.id.nav_host_fragment_content_main);
            return NavigationUI.navigateUp(navController, appBarConfiguration)
                    || super.onSupportNavigateUp();
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        return false;
    }
    //------------------------------------------------------------
    public AbstractObject.IStringDataCallback _LoginCallBack = new AbstractObject.IStringDataCallback()
    {
        public void OnData(String _token)
        {
            try
            {
                if(_Client.IsConnected())
                {
                    _Client.MQTTSubscribe("+/#");
                    _Client.AddMQTTCallback(new MqttCallback()
                    {
                        public void connectionLost(Throwable cause){}
                        public void deliveryComplete(IMqttDeliveryToken token){}
                        public void messageArrived(String _topic, MqttMessage _msg) throws Exception
                        {
                            try
                            {
                                Tools.Log.Print(Tools.LogLevel.ll_Info, "Get MQTT Msg : "+_topic);
                                Tools.Log.Print(Tools.LogLevel.ll_Info, new String(_msg.getPayload()));
                            }
                            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
                        }
                    });
                }
            }
            catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
        }
    };
    //------------------------------------------------------------
    public void InitMarsClient()
    {
        try
        {
            _Client = new MarsClient("test", "test", "justtest");
            _Client.SetCloudServerURL("https://test.mars-cloud.com");
            _Client.Start(_LoginCallBack);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //------------------------------------------------------------
    public void MarsClientTest()
    {
        try
        {
            if(_Client.IsConnected() == false) return;

            JSONObject _data = new JSONObject();

            _data.put("temp", 23.7);
            _data.put("humi", 80);

            _Client.PutData("dev", "test", _data);
            _Client.GetLastData("dev", "test", 1, new
                    AbstractObject.IStringDataCallback()
            {
                public void OnData(String _payload)
                {
                    Tools.Log.Print(Tools.LogLevel.ll_Info, "Put Data SUCCESS");
                }
            });

            _Client.PutEvent("dev", "event", _data);
            _Client.PutMessage("msg/my/test", _data);
        }
        catch(Exception _e){ Tools.ExceptionMsgPrintOut(_e, new Object(){}.getClass()); };
    }
    //------------------------------------------------------------
}