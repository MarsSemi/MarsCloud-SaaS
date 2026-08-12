package Security

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/MarsJSON"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Tools"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
)

// -------------------------------------------------------------------------------------
//
// -------------------------------------------------------------------------------------
type TokenMethod string

// -------------------------------------------------------------------------------------
const (
	TM_NONE  TokenMethod = "NONE"
	TM_AES   TokenMethod = "AES"
	TM_AES32 TokenMethod = "AES32"
	TM_RSA   TokenMethod = "RSA"
)

const (
	extendedJWTVersion       = 1
	extendedJWTMaxExtensions = 2
	extendedJWTMaxPayload    = 64 * 1024
	extendedJWTMaxExtension  = 16 * 1024
	extendedJWTMaxToken      = 160 * 1024
	extendedJWTNonceSize     = 12
	extendedJWTMarker        = "mars_ext"
	extendedJWTCount         = "ext_count"
	extendedJWTClaimsKey     = "mars_extensions"
)

// ExtendedJWT 是驗證完成的 MARS Extended JWT。
// Claims 對應正式 JWT 第二段；Extensions 是已完成 AES-GCM 解密的第四、五段。
type ExtendedJWT struct {
	Header     map[string]interface{}   `json:"header"`
	Claims     map[string]interface{}   `json:"claims"`
	Extensions []map[string]interface{} `json:"extensions,omitempty"`
}

// -------------------------------------------------------------------------------------
func (_tm TokenMethod) Value() string {
	return string(_tm)
}

// -------------------------------------------------------------------------------------
// JWTProcessor 結構體
// -------------------------------------------------------------------------------------
type JWTProcessor struct {
	_SecretKey  []byte
	_PublicKey  *rsa.PublicKey
	_PrivateKey *rsa.PrivateKey
	_AESBlock   cipher.Block
	_AES_IV     []byte
	_mu         sync.RWMutex // 保護以上欄位的並行讀寫；外部 method 拿鎖，內部 *Locked helper 不再拿
}

// ExtendedTokenReady 回報 Extended JWT 所需的 RSA 簽章與 AES extension 金鑰是否已載入。
func (_this *JWTProcessor) ExtendedTokenReady() bool {
	_this._mu.RLock()
	defer _this._mu.RUnlock()
	return _this._PublicKey != nil && _this._PrivateKey != nil && len(_this._SecretKey) > 0
}

// -------------------------------------------------------------------------------------
// 全域實例 (模擬 Java 的靜態成員)
var JWT = &JWTProcessor{
	_AES_IV: generateNumericString(16), // 與 Java 相同的 IV
}

// -------------------------------------------------------------------------------------
func generateNumericString(length int) []byte {
	const charset = "0123456789"

	_result := make([]byte, length)
	for i := range _result {
		// 安全地生成 0-9 的隨機索引
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil
		}

		_result[i] = charset[num.Int64()]
	}

	return _result
}

// -------------------------------------------------------------------------------------
// RSA 相關功能
// -------------------------------------------------------------------------------------
// NewRSAKey 重新產生並儲存 RSA 金鑰
func (_this *JWTProcessor) NewRSAKey(_method string, _pubPath string, _priPath string) bool {
	os.Remove(_pubPath)
	os.Remove(_priPath)
	return _this.LoadRSAKeyFromFile(_pubPath, _priPath)
}

// -------------------------------------------------------------------------------------
// LoadRSAKey 從位元組載入 RSA 金鑰
func (_this *JWTProcessor) LoadRSAKey(_pubKey []byte, _priKey []byte) bool {
	_this._mu.Lock()
	defer _this._mu.Unlock()
	return _this.loadRSAKeyLocked(_pubKey, _priKey)
}

// -------------------------------------------------------------------------------------
func (_this *JWTProcessor) loadRSAKeyLocked(_pubKey []byte, _priKey []byte) bool {

	if len(_pubKey) > 0 && len(_priKey) > 0 {

		// 解析公鑰 (X509)
		_blockPub, _ := pem.Decode(_pubKey)
		if _blockPub != nil {
			_pub, _err := x509.ParsePKIXPublicKey(_blockPub.Bytes)
			if _err == nil {
				_this._PublicKey = _pub.(*rsa.PublicKey)
			}
		}

		// 解析私鑰 (PKCS8)
		_blockPri, _ := pem.Decode(_priKey)
		if _blockPri != nil {
			_pri, _err := x509.ParsePKCS8PrivateKey(_blockPri.Bytes)
			if _err == nil {
				_this._PrivateKey = _pri.(*rsa.PrivateKey)
			}
		}
	}

	if _this._PublicKey == nil || _this._PrivateKey == nil {

		Tools.Log.Print(Tools.LL_Warning, "JWS RSA Key is empty, dynamic generating ...")
		_pri, _err := rsa.GenerateKey(rand.Reader, 2048) // Go 建議至少 2048
		if _err != nil {
			return false
		}

		_this._PrivateKey = _pri
		_this._PublicKey = &_pri.PublicKey
	}

	Tools.Log.Print(Tools.LL_Info, "JWS RSA Key is ready")
	return true
}

// -------------------------------------------------------------------------------------
func (_this *JWTProcessor) LoadRSAKeyFromFile(_pubPath string, _priPath string) bool {
	_pubBinary, _ := os.ReadFile(_pubPath)
	_priBinary, _ := os.ReadFile(_priPath)

	_this._mu.Lock()
	_resp := _this.loadRSAKeyLocked(_pubBinary, _priBinary)
	_this._mu.Unlock()

	if _pubBinary == nil || _priBinary == nil {
		_this.SaveRSAKey(_pubPath, _priPath)
	}
	return _resp
}

// -------------------------------------------------------------------------------------
func (_this *JWTProcessor) SaveRSAKey(_pubPath string, _priPath string) bool {
	_this._mu.RLock()
	defer _this._mu.RUnlock()

	if _this._PublicKey != nil && _this._PrivateKey != nil {
		_pubASN1, _ := x509.MarshalPKIXPublicKey(_this._PublicKey)
		_priASN1, _ := x509.MarshalPKCS8PrivateKey(_this._PrivateKey)

		_pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: _pubASN1})
		_priPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: _priASN1})

		os.WriteFile(_pubPath, _pubPEM, 0644)
		os.WriteFile(_priPath, _priPEM, 0600)
		return true
	}
	return false
}

//-------------------------------------------------------------------------------------
// AES 相關功能
//-------------------------------------------------------------------------------------

func (_this *JWTProcessor) LoadAESKey(_key []byte) bool {
	_this._mu.Lock()
	defer _this._mu.Unlock()
	return _this.loadAESKeyLocked(_key, 16)
}

// -------------------------------------------------------------------------------------
func (_this *JWTProcessor) LoadAESKey32(_key []byte) bool {
	_this._mu.Lock()
	defer _this._mu.Unlock()
	return _this.loadAESKeyLocked(_key, 32)
}

// -------------------------------------------------------------------------------------
func (_this *JWTProcessor) loadAESKeyLocked(_key []byte, _size int) bool {
	if len(_key) > 0 {
		_this._SecretKey = _key
	} else {
		Tools.Log.Print(Tools.LL_Warning, "JWS AES Key is empty, generating ...")
		_this._SecretKey = make([]byte, _size)
		io.ReadFull(rand.Reader, _this._SecretKey)
	}

	_block, _err := aes.NewCipher(_this._SecretKey)
	if _err != nil {
		return false
	}
	_this._AESBlock = _block
	Tools.Log.Print(Tools.LL_Info, "JWS AES Key is ready")
	return true
}

// -------------------------------------------------------------------------------------
func (_this *JWTProcessor) LoadAESKeyFromFile(_path string) bool {
	_key, _ := os.ReadFile(_path)

	_this._mu.Lock()
	_resp := _this.loadAESKeyLocked(_key, 16)
	_secret := _this._SecretKey
	_this._mu.Unlock()

	if _key == nil {
		os.WriteFile(_path, _secret, 0600)
	}
	return _resp
}

// -------------------------------------------------------------------------------------
// Token (JWE) 加解密
// -------------------------------------------------------------------------------------
// CreateToken 建立 JWE Token (支援 RSA 與 AES/Direct)
func (_this *JWTProcessor) CreateToken(_method string, _root map[string]interface{}) string {
	_this._mu.RLock()
	defer _this._mu.RUnlock()

	_payload, _ := json.Marshal(_root)
	_headers := jwe.NewHeaders()
	_headers.Set("com", "mars-semi.com")

	var _token []byte
	var _err error

	if _method == TM_RSA.Value() && _this._PublicKey != nil {
		// 與舊 Java 相容：使用 RSA-OAEP + A128GCM
		_token, _err = jwe.Encrypt(
			_payload,
			jwe.WithKey(jwa.RSA_OAEP, _this._PublicKey),
			jwe.WithContentEncryption(jwa.A128GCM),
			jwe.WithProtectedHeaders(_headers),
		)
	} else if _method == TM_AES.Value() && len(_this._SecretKey) > 0 {
		// 直接使用 AES Key 加密 (Direct)
		_token, _err = jwe.Encrypt(_payload, jwe.WithKey(jwa.DIRECT, _this._SecretKey), jwe.WithContentEncryption(jwa.A128GCM), jwe.WithProtectedHeaders(_headers))
	} else if _method == TM_AES32.Value() && len(_this._SecretKey) > 0 {
		// 直接使用 AES Key 加密 (Direct)
		_token, _err = jwe.Encrypt(_payload, jwe.WithKey(jwa.DIRECT, _this._SecretKey), jwe.WithProtectedHeaders(_headers))
	}

	if _err != nil {
		Tools.Log.Print(Tools.LL_Debug, fmt.Sprintf("JWS Create Error : %v", _err))
		return ""
	}

	return string(_token)
}

// -------------------------------------------------------------------------------------
// CreateExtendedToken 建立 MARS Extended JWT：
// header.payload.signature[.encrypted-extension4][.encrypted-extension5]
//
// 第二段 payload 完整由呼叫端提供；SDK 僅負責 JSON 編碼、RSA-SHA256 簽章，
// 以及使用 AES-GCM 加密最多兩個 extension。
func (_this *JWTProcessor) CreateExtendedToken(_claims map[string]interface{}, _extensions []map[string]interface{}) (string, error) {
	if len(_claims) == 0 {
		return "", fmt.Errorf("extended JWT claims are required")
	}
	if len(_extensions) > extendedJWTMaxExtensions {
		return "", fmt.Errorf("extended JWT supports at most %d extensions", extendedJWTMaxExtensions)
	}

	_this._mu.RLock()
	defer _this._mu.RUnlock()

	if _this._PrivateKey == nil {
		return "", fmt.Errorf("extended JWT RSA private key is not ready")
	}
	if len(_extensions) > 0 && len(_this._SecretKey) == 0 {
		return "", fmt.Errorf("extended JWT AES key is not ready")
	}

	_header := map[string]interface{}{
		"alg":             "RS256",
		"typ":             "JWT",
		"com":             "mars-semi.com",
		extendedJWTMarker: extendedJWTVersion,
		extendedJWTCount:  len(_extensions),
	}
	_headerJSON, _err := json.Marshal(_header)
	if _err != nil {
		return "", fmt.Errorf("encode extended JWT header: %w", _err)
	}
	_payloadJSON, _err := json.Marshal(_claims)
	if _err != nil {
		return "", fmt.Errorf("encode extended JWT payload: %w", _err)
	}
	if len(_payloadJSON) > extendedJWTMaxPayload {
		return "", fmt.Errorf("extended JWT payload exceeds %d bytes", extendedJWTMaxPayload)
	}

	_base64 := base64.RawURLEncoding
	_signingInput := _base64.EncodeToString(_headerJSON) + "." + _base64.EncodeToString(_payloadJSON)
	_digest := sha256.Sum256([]byte(_signingInput))
	_signature, _err := rsa.SignPKCS1v15(rand.Reader, _this._PrivateKey, cryptoHashSHA256, _digest[:])
	if _err != nil {
		return "", fmt.Errorf("sign extended JWT: %w", _err)
	}
	_baseToken := _signingInput + "." + _base64.EncodeToString(_signature)
	if len(_extensions) == 0 {
		return _baseToken, nil
	}

	_extensionKey := deriveExtendedJWTKey(_this._SecretKey, _baseToken)
	_block, _err := aes.NewCipher(_extensionKey)
	if _err != nil {
		return "", fmt.Errorf("initialize extended JWT cipher: %w", _err)
	}
	_gcm, _err := cipher.NewGCMWithNonceSize(_block, extendedJWTNonceSize)
	if _err != nil {
		return "", fmt.Errorf("initialize extended JWT GCM: %w", _err)
	}

	_parts := []string{_baseToken}
	for _index, _extension := range _extensions {
		if _extension == nil {
			return "", fmt.Errorf("extended JWT extension %d is required", _index+1)
		}
		_plaintext, _err := json.Marshal(_extension)
		if _err != nil {
			return "", fmt.Errorf("encode extended JWT extension %d: %w", _index+1, _err)
		}
		if len(_plaintext) > extendedJWTMaxExtension {
			return "", fmt.Errorf("extended JWT extension %d exceeds %d bytes", _index+1, extendedJWTMaxExtension)
		}
		_nonce := make([]byte, _gcm.NonceSize())
		if _, _err = io.ReadFull(rand.Reader, _nonce); _err != nil {
			return "", fmt.Errorf("create extended JWT extension %d nonce: %w", _index+1, _err)
		}
		_aad := extendedJWTAAD(_baseToken, _index, len(_extensions))
		_ciphertext := _gcm.Seal(nil, _nonce, _plaintext, _aad)
		_envelope := make([]byte, 1+len(_nonce)+len(_ciphertext))
		_envelope[0] = extendedJWTVersion
		copy(_envelope[1:], _nonce)
		copy(_envelope[1+len(_nonce):], _ciphertext)
		_parts = append(_parts, _base64.EncodeToString(_envelope))
	}
	return strings.Join(_parts, "."), nil
}

// -------------------------------------------------------------------------------------
// IsExtendedToken 只辨識格式標記，真正的簽章與 extension 驗證由 VerifyExtendedToken 執行。
func (_this *JWTProcessor) IsExtendedToken(_token string) bool {
	_token = strings.TrimSpace(_token)
	if len(_token) > extendedJWTMaxToken {
		return false
	}
	_parts := strings.Split(_token, ".")
	if len(_parts) < 3 || len(_parts) > 5 {
		return false
	}
	_headerJSON, _err := base64.RawURLEncoding.DecodeString(_parts[0])
	if _err != nil {
		return false
	}
	var _header map[string]interface{}
	if json.Unmarshal(_headerJSON, &_header) != nil {
		return false
	}
	_version, _ok := numericJSONInt(_header[extendedJWTMarker])
	return _ok && _version == extendedJWTVersion
}

// -------------------------------------------------------------------------------------
// VerifyExtendedToken 驗證前三段 RSA-SHA256 簽章、有效期與 extension 數量，
// 並以 AES-GCM 驗證及解密第四、五段。任一段失敗時不回傳部分結果。
func (_this *JWTProcessor) VerifyExtendedToken(_token string, _ignoreExp bool) (*ExtendedJWT, error) {
	_token = strings.TrimSpace(_token)
	if len(_token) > extendedJWTMaxToken {
		return nil, fmt.Errorf("extended JWT exceeds %d bytes", extendedJWTMaxToken)
	}
	_parts := strings.Split(_token, ".")
	if len(_parts) < 3 || len(_parts) > 5 {
		return nil, fmt.Errorf("invalid extended JWT segment count")
	}

	_this._mu.RLock()
	defer _this._mu.RUnlock()

	if _this._PublicKey == nil {
		return nil, fmt.Errorf("extended JWT RSA public key is not ready")
	}
	_headerJSON, _err := base64.RawURLEncoding.DecodeString(_parts[0])
	if _err != nil {
		return nil, fmt.Errorf("decode extended JWT header: %w", _err)
	}
	var _header map[string]interface{}
	if _err = json.Unmarshal(_headerJSON, &_header); _err != nil {
		return nil, fmt.Errorf("decode extended JWT header JSON: %w", _err)
	}
	if _header["alg"] != "RS256" || _header["typ"] != "JWT" {
		return nil, fmt.Errorf("unsupported extended JWT header")
	}
	_version, _ok := numericJSONInt(_header[extendedJWTMarker])
	if !_ok || _version != extendedJWTVersion {
		return nil, fmt.Errorf("unsupported extended JWT version")
	}
	_expectedCount, _ok := numericJSONInt(_header[extendedJWTCount])
	if !_ok || _expectedCount < 0 || _expectedCount > extendedJWTMaxExtensions || _expectedCount != len(_parts)-3 {
		return nil, fmt.Errorf("extended JWT extension count mismatch")
	}

	_signingInput := _parts[0] + "." + _parts[1]
	_signature, _err := base64.RawURLEncoding.DecodeString(_parts[2])
	if _err != nil {
		return nil, fmt.Errorf("decode extended JWT signature: %w", _err)
	}
	_digest := sha256.Sum256([]byte(_signingInput))
	if _err = rsa.VerifyPKCS1v15(_this._PublicKey, cryptoHashSHA256, _digest[:], _signature); _err != nil {
		return nil, fmt.Errorf("verify extended JWT signature: %w", _err)
	}

	_payloadJSON, _err := base64.RawURLEncoding.DecodeString(_parts[1])
	if _err != nil || len(_payloadJSON) == 0 || len(_payloadJSON) > extendedJWTMaxPayload {
		return nil, fmt.Errorf("invalid extended JWT payload")
	}
	var _claims map[string]interface{}
	if _err = json.Unmarshal(_payloadJSON, &_claims); _err != nil || len(_claims) == 0 {
		return nil, fmt.Errorf("decode extended JWT payload")
	}
	if !_ignoreExp && !jsonClaimsTimeValid(_claims) {
		return nil, fmt.Errorf("extended JWT is expired or has an invalid exp claim")
	}

	_result := &ExtendedJWT{Header: _header, Claims: _claims}
	if _expectedCount == 0 {
		return _result, nil
	}
	if len(_this._SecretKey) == 0 {
		return nil, fmt.Errorf("extended JWT AES key is not ready")
	}
	_baseToken := strings.Join(_parts[:3], ".")
	_extensionKey := deriveExtendedJWTKey(_this._SecretKey, _baseToken)
	_block, _err := aes.NewCipher(_extensionKey)
	if _err != nil {
		return nil, fmt.Errorf("initialize extended JWT cipher: %w", _err)
	}
	_gcm, _err := cipher.NewGCMWithNonceSize(_block, extendedJWTNonceSize)
	if _err != nil {
		return nil, fmt.Errorf("initialize extended JWT GCM: %w", _err)
	}
	_result.Extensions = make([]map[string]interface{}, 0, _expectedCount)
	for _index := 0; _index < _expectedCount; _index++ {
		_envelope, _err := base64.RawURLEncoding.DecodeString(_parts[_index+3])
		if _err != nil || len(_envelope) <= 1+_gcm.NonceSize()+_gcm.Overhead() || _envelope[0] != extendedJWTVersion {
			return nil, fmt.Errorf("invalid extended JWT extension %d", _index+1)
		}
		_nonce := _envelope[1 : 1+_gcm.NonceSize()]
		_ciphertext := _envelope[1+_gcm.NonceSize():]
		_plaintext, _err := _gcm.Open(nil, _nonce, _ciphertext, extendedJWTAAD(_baseToken, _index, _expectedCount))
		if _err != nil || len(_plaintext) > extendedJWTMaxExtension {
			return nil, fmt.Errorf("decrypt extended JWT extension %d", _index+1)
		}
		var _extension map[string]interface{}
		if _err = json.Unmarshal(_plaintext, &_extension); _err != nil || _extension == nil {
			return nil, fmt.Errorf("decode extended JWT extension %d", _index+1)
		}
		_result.Extensions = append(_result.Extensions, _extension)
	}
	return _result, nil
}

// cryptoHashSHA256 避免呼叫端依賴實作細節，並與 RSA PKCS#1 v1.5 的 RS256 定義一致。
const cryptoHashSHA256 = crypto.Hash(crypto.SHA256)

func deriveExtendedJWTKey(_secret []byte, _baseToken string) []byte {
	_mac := hmac.New(sha256.New, _secret)
	_mac.Write([]byte("mars-sdk/extended-jwt/v1\x00"))
	_mac.Write([]byte(_baseToken))
	return _mac.Sum(nil)
}

func extendedJWTAAD(_baseToken string, _index int, _count int) []byte {
	return []byte(_baseToken + "\x00mars-ext-v1\x00" + strconv.Itoa(_index) + "\x00" + strconv.Itoa(_count))
}

func numericJSONInt(_value interface{}) (int, bool) {
	switch _typed := _value.(type) {
	case float64:
		_value := int(_typed)
		return _value, float64(_value) == _typed
	case json.Number:
		_value, _err := strconv.Atoi(_typed.String())
		return _value, _err == nil
	case int:
		return _typed, true
	case int64:
		return int(_typed), int64(int(_typed)) == _typed
	default:
		return 0, false
	}
}

func jsonClaimsTimeValid(_claims map[string]interface{}) bool {
	_expValue, _exists := _claims["exp"]
	if !_exists {
		return true
	}
	switch _exp := _expValue.(type) {
	case float64:
		return _exp >= float64(time.Now().Unix())
	case json.Number:
		_value, _err := _exp.Int64()
		return _err == nil && _value >= time.Now().Unix()
	case int64:
		return _exp >= time.Now().Unix()
	case int:
		return int64(_exp) >= time.Now().Unix()
	default:
		return false
	}
}

// -------------------------------------------------------------------------------------
// DecryptToken 解密並驗證 Token
func (_this *JWTProcessor) DecryptToken(_tokenStr string, _ignoreExp bool) *MarsJSON.JSONObject {
	if !isCompactJWEToken(_tokenStr) {
		return nil
	}

	_this._mu.RLock()
	defer _this._mu.RUnlock()

	var _decrypted []byte
	var _err error
	_attempted := false

	// 嘗試使用私鑰解密 (RSA)
	if _this._PrivateKey != nil {
		_attempted = true
		_decrypted, _err = jwe.Decrypt([]byte(_tokenStr), jwe.WithKey(jwa.RSA_OAEP, _this._PrivateKey))
		if _err != nil {
			_decrypted, _err = jwe.Decrypt([]byte(_tokenStr), jwe.WithKey(jwa.RSA_OAEP_256, _this._PrivateKey))
		}
	}

	// 若 RSA 失敗或無私鑰，嘗試 AES
	if (!_attempted || _err != nil) && len(_this._SecretKey) > 0 {
		_attempted = true
		_decrypted, _err = jwe.Decrypt([]byte(_tokenStr), jwe.WithKey(jwa.DIRECT, _this._SecretKey))
	}

	if !_attempted || _err != nil || len(_decrypted) == 0 {
		return nil
	}

	// 解析 Payload
	var _obj map[string]interface{}
	if _err = json.Unmarshal(_decrypted, &_obj); _err != nil || len(_obj) == 0 {
		return nil
	}

	// 驗證有效期 (exp)
	if _expValue, _exists := _obj["exp"]; _exists {
		_exp, _ok := _expValue.(float64)
		if !_ok {
			return nil
		}
		if !_ignoreExp && int64(_exp) < time.Now().Unix() {
			Tools.Log.Print(Tools.LL_Debug, "JWS token is out of time")
			return nil
		}
	}

	return MarsJSON.NewJSONObject(_obj)
}

// -------------------------------------------------------------------------------------
// AES Data 加解密 (CBC)
// -------------------------------------------------------------------------------------
func (_this *JWTProcessor) EncryptAESData(_id string) string {
	_this._mu.RLock()
	defer _this._mu.RUnlock()

	if _this._AESBlock == nil {
		return ""
	}
	_plaintext := []byte(_id)
	// PKCS7 Padding
	_padding := aes.BlockSize - (len(_plaintext) % aes.BlockSize)
	_padtext := append(_plaintext, bytes.Repeat([]byte{byte(_padding)}, _padding)...)

	_ciphertext := make([]byte, len(_padtext))
	_mode := cipher.NewCBCEncrypter(_this._AESBlock, _this._AES_IV)
	_mode.CryptBlocks(_ciphertext, _padtext)

	return base64.StdEncoding.EncodeToString(_ciphertext)
}

// -------------------------------------------------------------------------------------
func (_this *JWTProcessor) DecryptAESData(_data string) string {
	_this._mu.RLock()
	defer _this._mu.RUnlock()

	if _this._AESBlock == nil {
		return ""
	}
	_ciphertext, _ := base64.StdEncoding.DecodeString(_data)
	_mode := cipher.NewCBCDecrypter(_this._AESBlock, _this._AES_IV)
	_mode.CryptBlocks(_ciphertext, _ciphertext)

	// Unpadding
	_padding := int(_ciphertext[len(_ciphertext)-1])
	return string(_ciphertext[:len(_ciphertext)-_padding])
}
