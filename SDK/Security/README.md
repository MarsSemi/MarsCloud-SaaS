# Security

`Security` 模組提供 JWT、AES、RSA 與使用者 token 驗證相關功能，是整個 SDK 的安全基礎。

## 主要能力

- 產生與解密 JWE token
- 產生與驗證 MARS Extended JWT
- 載入 AES / RSA 金鑰
- 驗證 HTTP token
- 舊版 token 與舊金鑰 fallback

## 主要型別

- `JWTProcessor`
- `UserGroup`

## 全域物件

- `Security.JWT`

## 常用函式

JWT：

- `(_this *JWTProcessor) LoadRSAKey(_pubKey []byte, _priKey []byte) bool`
- `(_this *JWTProcessor) LoadRSAKeyFromFile(_pubPath string, _priPath string) bool`
- `(_this *JWTProcessor) LoadAESKey(_key []byte) bool`
- `(_this *JWTProcessor) LoadAESKeyFromFile(_path string) bool`
- `(_this *JWTProcessor) CreateToken(_method string, _root map[string]interface{}) string`
- `(_this *JWTProcessor) DecryptToken(_tokenStr string, _ignoreExp bool) *MarsJSON.JSONObject`
- `(_this *JWTProcessor) CreateExtendedToken(_claims map[string]interface{}, _extensions []map[string]interface{}) (string, error)`
- `(_this *JWTProcessor) VerifyExtendedToken(_token string, _ignoreExp bool) (*ExtendedJWT, error)`
- `(_this *JWTProcessor) IsExtendedToken(_token string) bool`

User verify：

- `VerifyToken(_auth_string string, _group UserGroup, _ipadd_from string) *MarsJSON.JSONObject`
- `DecryptToken(_auth_string string, _ignore_timetolive bool) *MarsJSON.JSONObject`

## 相容性重點

- Security 驗證器處理 compact JWE，以及 header 帶有 `mars_ext: 1` 的 MARS Extended JWT；opaque session token 會直接回傳 `nil`，由應用層驗證器接手
- RSA token 建立採用 `RSA-OAEP + A128GCM`
- 解密時會同時嘗試：
  - `RSA-OAEP`
  - `RSA-OAEP-256`
- `UserVerify` 會自動 fallback 相容金鑰：
  - `default_*`
  - `legacy_*`
  - `compat_*`

## 基本範例

```go
Security.JWT.LoadRSAKey(nil, nil)
Security.JWT.LoadAESKey(nil)

token := Security.JWT.CreateToken(Security.TM_AES.Value(), map[string]interface{}{
    "iss": "tester",
    "exp": time.Now().Add(time.Hour).Unix(),
})

claims := Security.JWT.DecryptToken(token, false)
```

## MARS Extended JWT

格式如下：

```text
header.payload.signature[.encrypted-extension4][.encrypted-extension5]
```

- 前三段是 `RS256` JWT，第二段 payload 完整由呼叫端提供。
- 第四、五段為選填 JSON object，使用衍生自 SDK AES key 的 AES-256-GCM 加密。
- header 內的 `ext_count` 由 RSA 簽章保護；AES-GCM AAD 綁定前三段、extension 索引與總數，因此 extension 不可被抽換、移位、附加或截短。
- 最多接受兩個 extension，每段明文上限 16 KiB，JWT payload 上限 64 KiB。
- `VerifyExtendedToken` 任一段驗證失敗時不回傳部分 claims。

```go
token, err := Security.JWT.CreateExtendedToken(
    map[string]interface{}{
        "iss": "account-manager",
        "sub": "account-1",
        "exp": time.Now().Add(time.Hour).Unix(),
    },
    []map[string]interface{}{
        {"tenant": "factory"},
        {"features": []string{"audit"}},
    },
)

verified, err := Security.JWT.VerifyExtendedToken(token, false)
```

## 注意事項

- `exp` 以秒為單位
- 未載入可用金鑰、解密失敗、payload 不是有效 JSON object 或 claims 為空時，一律回傳 `nil`
- `VerifyToken` 只會對結構有效但驗證失敗的 compact JWE 或 MARS Extended JWT 引入短暫 delay；其他 token 格式不做無效的密碼學運算
