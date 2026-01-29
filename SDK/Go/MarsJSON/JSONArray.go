package MarsJSON

// -------------------------------------------------------------------------------------
import (
	"encoding/json"
	"fmt"
)

//-------------------------------------------------------------------------------------
// JSONArray 實作
//-------------------------------------------------------------------------------------

type JSONArray struct {
	_Src []interface{}
}

// -------------------------------------------------------------------------------------
func NewJSONArray(_input interface{}) *JSONArray {

	if _input != nil {

		switch _v := _input.(type) {
		case string:
			{
				_v = fixJsonString(_v)
				_json := make([]interface{}, 0)
				_err := json.Unmarshal([]byte(_v), &_json)

				if _err != nil {
					//fmt.Println(_err)
					return &JSONArray{_Src: make([]interface{}, 0)}
				}

				return &JSONArray{_Src: _json}
			}
		case []byte:
			{
				_json := make([]interface{}, 0)
				_err := json.Unmarshal(_v, &_json)

				if _err != nil {
					//fmt.Println(_err)
					return &JSONArray{_Src: make([]interface{}, 0)}
				}

				return &JSONArray{_Src: _json}
			}
		case []interface{}:
			return &JSONArray{_Src: _v}
		default:
			break
		}
	}

	return &JSONArray{_Src: make([]interface{}, 0)}
}

// -------------------------------------------------------------------------------------
func (_ja *JSONArray) Put(_value interface{}) *JSONArray {
	switch _v := _value.(type) {
	case *JSONObject:
		_ja._Src = append(_ja._Src, _v._Src)
	case *JSONArray:
		_ja._Src = append(_ja._Src, _v._Src)
	default:
		_ja._Src = append(_ja._Src, _value)
	}
	return _ja
}

// -------------------------------------------------------------------------------------
func (_ja *JSONArray) OptJSONObject(_index int) *JSONObject {
	if _index >= 0 && _index < len(_ja._Src) {
		if _m, _ok := _ja._Src[_index].(map[string]interface{}); _ok {
			return NewJSONObject(_m)
		}
	}
	return NewJSONObject(nil)
}

// -------------------------------------------------------------------------------------
func (_ja *JSONArray) OptString(_index int, _defaultValue string) string {
	if _index >= 0 && _index < len(_ja._Src) {
		return fmt.Sprintf("%v", _ja._Src[_index])
	}
	return _defaultValue
}

// -------------------------------------------------------------------------------------
func (_ja *JSONArray) Opt(_index int) any {
	if _index >= 0 && _index < len(_ja._Src) {
		return _ja._Src[_index]
	}

	return nil
}

// -------------------------------------------------------------------------------------
func (_ja *JSONArray) Remove(_index int) interface{} {
	if _ja._Src == nil || _index < 0 || _index >= len(_ja._Src) {
		return nil
	}

	// 取得要被移除的元素
	_val := _ja._Src[_index]

	// 執行切片移除邏輯：將 index 之前與之後的元素重新合併
	// 寫法等同於：_ja._Src = append(_ja._Src[:_index], _ja._Src[_index+1:]...)
	_ja._Src = append(_ja._Src[:_index], _ja._Src[_index+1:]...)

	return _val
}

// -------------------------------------------------------------------------------------
func (_ja *JSONArray) Length() int {
	return len(_ja._Src)
}

// -------------------------------------------------------------------------------------
func (_ja *JSONArray) ToString() string {
	_b, _ := json.Marshal(_ja._Src)
	return string(_b)
}

// -------------------------------------------------------------------------------------
