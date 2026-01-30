package MQTTClient

import (
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// -------------------------------------------------------------------------------------

// -------------------------------------------------------------------------------------
//
// -------------------------------------------------------------------------------------
type MQTTMessage struct {
	_Payload  []byte
	_Qos      byte
	_Retained bool
}

// -------------------------------------------------------------------------------------
func (_this *MQTTMessage) GetPayload() []byte   { return _this._Payload }
func (_this *MQTTMessage) SetPayload(_p []byte) { _this._Payload = _p }

//-------------------------------------------------------------------------------------
// MQTTCallback 模擬 org.eclipse.paho.client.MQTTv3.MQTTCallback 介面
//-------------------------------------------------------------------------------------

type MQTTCallback interface {
	OnConnected()
	OnConnectionLost(error)
	OnMessageArrived(string, *MQTTMessage)
	OnDeliveryComplete(string) // 在 Go 中簡化處理
}

//-------------------------------------------------------------------------------------
// MQTTConnectOptions 模擬 org.eclipse.paho.client.mqttv3.MqttConnectOptions
//-------------------------------------------------------------------------------------

type MQTTConnectOptions struct {
	Server             string
	ClientID           string
	UserName           string
	Password           string
	CleanSession       bool
	KeepAlive          int
	ConnectionTimeout  int
	AutomaticReconnect bool
}

// -------------------------------------------------------------------------------------
func NewMQTTConnectOptions() *MQTTConnectOptions {
	return &MQTTConnectOptions{
		CleanSession:       true,
		KeepAlive:          60,
		ConnectionTimeout:  30,
		AutomaticReconnect: true,
	}
}

// -------------------------------------------------------------------------------------
func (_o *MQTTConnectOptions) SetServer(_s string)           { _o.Server = _s }
func (_o *MQTTConnectOptions) SetClientID(_c string)         { _o.ClientID = _c }
func (_o *MQTTConnectOptions) SetUserName(_u string)         { _o.UserName = _u }
func (_o *MQTTConnectOptions) SetPassword(_p []byte)         { _o.Password = string(_p) }
func (_o *MQTTConnectOptions) SetCleanSession(_c bool)       { _o.CleanSession = _c }
func (_o *MQTTConnectOptions) SetKeepAliveInterval(_k int)   { _o.KeepAlive = _k }
func (_o *MQTTConnectOptions) SetConnectionTimeout(_c int)   { _o.ConnectionTimeout = _c }
func (_o *MQTTConnectOptions) SetAutomaticReconnect(_a bool) { _o.AutomaticReconnect = _a }

// -------------------------------------------------------------------------------------
// MQTTClient
// -------------------------------------------------------------------------------------
type MQTTClient struct {
	_Client   mqtt.Client
	_Callback MQTTCallback
}

// -------------------------------------------------------------------------------------
func Create() (*MQTTClient, error) {
	_thisc := &MQTTClient{}

	return _thisc, nil
}

// -------------------------------------------------------------------------------------
func (_thisc *MQTTClient) SetCallback(_cb MQTTCallback) {
	_thisc._Callback = _cb
}

// -------------------------------------------------------------------------------------
func (_thisc *MQTTClient) Connect(_options *MQTTConnectOptions) error {

	_opts := mqtt.NewClientOptions()

	_opts.AddBroker(_options.Server)
	_opts.SetClientID(_options.ClientID)
	_opts.SetUsername(_options.UserName)
	_opts.SetPassword(_options.Password)
	_opts.SetCleanSession(_options.CleanSession)
	_opts.SetKeepAlive(time.Duration(_options.KeepAlive) * time.Second)
	_opts.SetConnectTimeout(time.Duration(_options.ConnectionTimeout) * time.Second)
	_opts.SetAutoReconnect(_options.AutomaticReconnect)
	//_opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})

	// 設定連線遺失回調
	_opts.OnConnectionLost = func(_c mqtt.Client, _err error) {
		if _thisc._Callback != nil {
			_thisc._Callback.OnConnectionLost(_err)
		}
	}

	// 設定預設訊息處理 (用於 Subscribe 時沒指定處理器的情況)
	_opts.DefaultPublishHandler = func(_c mqtt.Client, _this mqtt.Message) {
		if _thisc._Callback != nil {
			_thissg := &MQTTMessage{
				_Payload:  _this.Payload(),
				_Qos:      _this.Qos(),
				_Retained: _this.Retained(),
			}
			_thisc._Callback.OnMessageArrived(_this.Topic(), _thissg)
		}
	}

	_thisc._Client = mqtt.NewClient(_opts)
	if _token := _thisc._Client.Connect(); _token.Wait() && _token.Error() != nil {
		return _token.Error()
	}

	return nil
}

// -------------------------------------------------------------------------------------
func (_thisc *MQTTClient) Subscribe(_topic string, _qos int) error {
	if _token := _thisc._Client.Subscribe(_topic, byte(_qos), nil); _token.Wait() && _token.Error() != nil {
		return _token.Error()
	}
	return nil
}

// -------------------------------------------------------------------------------------
func (_thisc *MQTTClient) Publish(_topic string, _thisessage *MQTTMessage) error {
	_token := _thisc._Client.Publish(_topic, _thisessage._Qos, _thisessage._Retained, _thisessage._Payload)
	_token.Wait()
	return _token.Error()
}

// -------------------------------------------------------------------------------------
func (_thisc *MQTTClient) Disconnect(_quiesce int) {
	_thisc._Client.Disconnect(uint(_quiesce))
}

// -------------------------------------------------------------------------------------
func (_thisc *MQTTClient) IsConnected() bool {
	return _thisc._Client.IsConnected()
}

// -------------------------------------------------------------------------------------
