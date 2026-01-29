package HttpService

// -------------------------------------------------------------------------------------
import (
	"net/http"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
)

// -------------------------------------------------------------------------------------
// HttpAPI_System 處理系統管理 API
type HttpAPI_System struct {
	_Service IServiceControl
}

// -------------------------------------------------------------------------------------
type IServiceControl interface {
	RestartService()
	ShutdownService()
	GetProperty() *MarsJSON.JSONObject
	MergePropertyFrom(*MarsJSON.JSONObject)
	OnUpdateProperty()
}

// -------------------------------------------------------------------------------------
// NewHttpAPI_System 建立實例
func CreateHttpAPI_System(_service IServiceControl) *HttpAPI_System {

	return &HttpAPI_System{_Service: _service}
}

// -------------------------------------------------------------------------------------
// GetSetting 獲取當前服務設定
func (_h *HttpAPI_System) GetSetting() string {
	if _h._Service != nil && _h._Service.GetProperty() != nil {
		// 將屬性轉換為 JSON 字串
		return _h._Service.GetProperty().ToString()
	}
	return "{}"
}

// -------------------------------------------------------------------------------------
// UpdateSetting 更新服務設定
func (_h *HttpAPI_System) UpdateSetting(_body string) string {

	_payload := string(_body)
	_info := MarsJSON.NewJSONObject(_payload)

	if _h._Service != nil {
		// 合併並更新屬性
		_h._Service.MergePropertyFrom(_info)
		_h._Service.OnUpdateProperty()
		return "ok"
	}
	return "fail"
}

// -------------------------------------------------------------------------------------
func (_h *HttpAPI_System) Process(_w http.ResponseWriter, _r *http.Request, _jwt *MarsJSON.JSONObject, _path []string, _params *MarsJSON.JSONObject, _body string) []byte {

	_resp := ""
	_cmd := _path[len(_path)-1]

	switch _cmd {

	case "restart":
		if _h._Service != nil {
			_h._Service.RestartService() // 執行重啟
			return []byte("ok")
		}

	case "shutdown":
		if _h._Service != nil {
			_h._Service.ShutdownService() // 執行關閉
			return []byte("ok")
		}

	case "get_setting":
		_resp = _h.GetSetting()

	case "update_setting":
		_resp = _h.UpdateSetting(_body)

	}

	if len(_resp) > 0 {
		return []byte(_resp)
	}

	return nil
}

// -------------------------------------------------------------------------------------
