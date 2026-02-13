package HttpService

// -------------------------------------------------------------------------------------
import (
	"net/http"
	"net/url"
	"strings"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Security"
)

// -------------------------------------------------------------------------------------
// HttpAPI Callback 基礎類別
// -------------------------------------------------------------------------------------
type HttpAPI_Callback interface {
	Process(http.ResponseWriter, *http.Request, *MarsJSON.JSONObject, []string, *MarsJSON.JSONObject, string) []byte
}

// -------------------------------------------------------------------------------------
type HttpAPI struct {
	resfulAPI string
	callBack  HttpAPI_Callback
}

// -------------------------------------------------------------------------------------
func (_h *HttpAPI) servHTTP(_w http.ResponseWriter, _r *http.Request) {

	_uriOrg, _ := url.PathUnescape(_r.RequestURI)

	if checkURI(_w, _uriOrg) == false {
		return
	}

	if _h.callBack == nil {
		http.Error(_w, "Not Found", http.StatusNotFound)
		return
	}

	var _jwt *MarsJSON.JSONObject

	_jwt = nil
	_uriOrg = strings.Replace(_uriOrg, _h.resfulAPI+"/", "", 1)
	_auth := _r.Header.Get("Authentication")
	_items := strings.Split(_uriOrg, "/")

	if len(_auth) > 0 && strings.Contains(_auth, " ") {
		_auth = strings.Split(_auth, " ")[1]
		_jwt = Security.VerifyToken(_auth, "", _r.RemoteAddr)
	}

	_resp := _h.callBack.Process(_w, _r, _jwt, _items, nil, "")

	if _resp != nil {
		SendRespone(_w, http.StatusOK, "application/json; charset=UTF-8", _resp)
		return
	}

	http.Error(_w, "Not Found", http.StatusNotFound)
}

// -------------------------------------------------------------------------------------
func CreateHttpAPI(_impl HttpAPI_Callback) *HttpAPI {
	_httpAPI := &HttpAPI{
		callBack: _impl,
	}

	return _httpAPI
}

// ------------------------------------------------------------------------------------
