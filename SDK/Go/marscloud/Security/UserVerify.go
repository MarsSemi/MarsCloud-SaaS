package Security

// -------------------------------------------------------------------------------------
import (
	"strings"
	"time"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Tools"
)

// -------------------------------------------------------------------------------------
//
// -------------------------------------------------------------------------------------
type UserGroup string

// -------------------------------------------------------------------------------------
const (
	UT_Administrator UserGroup = "administrator"
	UT_User          UserGroup = "user"
	UT_Guest         UserGroup = "guest"
)

// ------------------------------------------------------------------------------------
// VerifyToken 驗證 Token 合法性與權限群組
func VerifyToken(_auth_string string, _group UserGroup, _ipadd_from string) *MarsJSON.JSONObject {
	// 使用 defer 捕捉異常並列印訊息
	defer Tools.GlobalRecovery()

	// 1. 執行解密
	_jobj := DecryptToken(_auth_string, false)

	if _jobj != nil {

		// 2. 驗證群組是否匹配 (如果 _group 為空則跳過驗證)
		_tokenGroup := strings.ToLower(_jobj.OptString("group", ""))

		if _group == "" || string(_group) == _tokenGroup {

			// 3. 記錄來源 IP 並回傳
			_connectFrom := ""
			if _ipadd_from != "" {
				_connectFrom = _ipadd_from
			}

			_jobj.Put("connect_from", _connectFrom)

			return _jobj
		}

		Tools.Log.Print(Tools.LL_Debug, "User Group Verify Error : %s != %s", _tokenGroup, _group)

	} else {

		// 4. 失敗時記錄 Debug Log
		if _ipadd_from != "" {
			Tools.Log.Print(Tools.LL_Debug, "Verify token fail from "+_ipadd_from)
		} else {
			Tools.Log.Print(Tools.LL_Debug, "Verify token fail : "+_auth_string)
		}
	}

	return nil
}

// -------------------------------------------------------------------------------------
// DecryptToken 執行 JWT 解密並處理逾時邏輯
func DecryptToken(_auth_string string, _ignore_timetolive bool) *MarsJSON.JSONObject {
	defer Tools.GlobalRecovery()

	if len(_auth_string) > 16 {
		// 1. 處理 "Bearer <token>" 格式
		_authParts := strings.Split(_auth_string, " ")
		_token := _authParts[len(_authParts)-1]

		// 2. 呼叫 Security.JWT 執行解密 (此處假設您已實作 Security 套件)
		_jobj := JWT.DecryptToken(_token, _ignore_timetolive)

		if _jobj != nil {
			return _jobj
		}
	}

	// 3. 失敗時強制延遲，防止網路暴力攻擊
	time.Sleep(500 * time.Millisecond)

	return nil
}

// -------------------------------------------------------------------------------------
