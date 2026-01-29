package Security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Tools"
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
}

// -------------------------------------------------------------------------------------
// 全域實例 (模擬 Java 的靜態成員)
var JWT = &JWTProcessor{
	_AES_IV: []byte("0102070408070209"), // 與 Java 相同的 IV
}

// -------------------------------------------------------------------------------------
// RSA 相關功能
// -------------------------------------------------------------------------------------
// NewRSAKey 重新產生並儲存 RSA 金鑰
func (_j *JWTProcessor) NewRSAKey(_method string, _pubPath string, _priPath string) bool {
	os.Remove(_pubPath)
	os.Remove(_priPath)
	return _j.LoadRSAKeyFromFile(_pubPath, _priPath)
}

// LoadRSAKey 從位元組載入 RSA 金鑰
func (_j *JWTProcessor) LoadRSAKey(_pubKey []byte, _priKey []byte) bool {
	if len(_pubKey) > 0 && len(_priKey) > 0 {
		// 解析公鑰 (X509)
		_blockPub, _ := pem.Decode(_pubKey)
		if _blockPub != nil {
			_pub, _err := x509.ParsePKIXPublicKey(_blockPub.Bytes)
			if _err == nil {
				_j._PublicKey = _pub.(*rsa.PublicKey)
			}
		}
		// 解析私鑰 (PKCS8)
		_blockPri, _ := pem.Decode(_priKey)
		if _blockPri != nil {
			_pri, _err := x509.ParsePKCS8PrivateKey(_blockPri.Bytes)
			if _err == nil {
				_j._PrivateKey = _pri.(*rsa.PrivateKey)
			}
		}
	}

	if _j._PublicKey == nil || _j._PrivateKey == nil {
		Tools.Log.Print(Tools.LL_Warning, "JWS RSA Key is empty, dynamic generating ...")
		_pri, _err := rsa.GenerateKey(rand.Reader, 2048) // Go 建議至少 2048
		if _err != nil {
			return false
		}
		_j._PrivateKey = _pri
		_j._PublicKey = &_pri.PublicKey
	}

	Tools.Log.Print(Tools.LL_Info, "JWS RSA Key is ready")
	return true
}

// -------------------------------------------------------------------------------------
func (_j *JWTProcessor) LoadRSAKeyFromFile(_pubPath string, _priPath string) bool {
	_pubBinary, _ := os.ReadFile(_pubPath)
	_priBinary, _ := os.ReadFile(_priPath)
	_resp := _j.LoadRSAKey(_pubBinary, _priBinary)

	if _pubBinary == nil || _priBinary == nil {
		_j.SaveRSAKey(_pubPath, _priPath)
	}
	return _resp
}

// -------------------------------------------------------------------------------------
func (_j *JWTProcessor) SaveRSAKey(_pubPath string, _priPath string) bool {
	if _j._PublicKey != nil && _j._PrivateKey != nil {
		_pubASN1, _ := x509.MarshalPKIXPublicKey(_j._PublicKey)
		_priASN1, _ := x509.MarshalPKCS8PrivateKey(_j._PrivateKey)

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

func (_j *JWTProcessor) LoadAESKey(_key []byte) bool {
	if len(_key) > 0 {
		_j._SecretKey = _key
	} else {
		Tools.Log.Print(Tools.LL_Warning, "JWS AES Key is empty, generating ...")
		_j._SecretKey = make([]byte, 16) // AES-128
		io.ReadFull(rand.Reader, _j._SecretKey)
	}

	_block, _err := aes.NewCipher(_j._SecretKey)
	if _err != nil {
		return false
	}
	_j._AESBlock = _block
	Tools.Log.Print(Tools.LL_Info, "JWS AES Key is ready")
	return true
}

// -------------------------------------------------------------------------------------
func (_j *JWTProcessor) LoadAESKey32(_key []byte) bool {
	if len(_key) > 0 {
		_j._SecretKey = _key
	} else {
		Tools.Log.Print(Tools.LL_Warning, "JWS AES32 Key is empty, generating ...")
		_j._SecretKey = make([]byte, 32) // AES-256
		io.ReadFull(rand.Reader, _j._SecretKey)
	}

	_block, _err := aes.NewCipher(_j._SecretKey)
	if _err != nil {
		return false
	}
	_j._AESBlock = _block
	Tools.Log.Print(Tools.LL_Info, "JWS AES32 Key is ready")
	return true
}

// -------------------------------------------------------------------------------------
func (_j *JWTProcessor) LoadAESKeyFromFile(_path string) bool {
	_key, _ := os.ReadFile(_path)
	_resp := _j.LoadAESKey(_key)
	if _key == nil {
		os.WriteFile(_path, _j._SecretKey, 0600)
	}
	return _resp
}

// -------------------------------------------------------------------------------------
// Token (JWE) 加解密
// -------------------------------------------------------------------------------------
// CreateToken 建立 JWE Token (支援 RSA 與 AES/Direct)
func (_j *JWTProcessor) CreateToken(_method string, _root map[string]interface{}) string {
	_payload, _ := json.Marshal(_root)
	_headers := jwe.NewHeaders()
	_headers.Set("com", "mars-semi.com")

	var _token []byte
	var _err error

	if _method == TM_RSA.Value() && _j._PublicKey != nil {
		// RSA-OAEP 加密
		_token, _err = jwe.Encrypt(_payload, jwe.WithKey(jwa.RSA_OAEP_256, _j._PublicKey), jwe.WithProtectedHeaders(_headers))
	} else if _method == TM_AES.Value() && len(_j._SecretKey) > 0 {
		// 直接使用 AES Key 加密 (Direct)
		_token, _err = jwe.Encrypt(_payload, jwe.WithKey(jwa.DIRECT, _j._SecretKey), jwe.WithContentEncryption(jwa.A128GCM), jwe.WithProtectedHeaders(_headers))
	} else if _method == TM_AES32.Value() && len(_j._SecretKey) > 0 {
		// 直接使用 AES Key 加密 (Direct)
		_token, _err = jwe.Encrypt(_payload, jwe.WithKey(jwa.DIRECT, _j._SecretKey), jwe.WithProtectedHeaders(_headers))
	}

	if _err != nil {
		Tools.Log.Print(Tools.LL_Debug, fmt.Sprintf("JWS Create Error : %v", _err))
		return ""
	}

	return string(_token)
}

// -------------------------------------------------------------------------------------
// DecryptToken 解密並驗證 Token
func (_j *JWTProcessor) DecryptToken(_tokenStr string, _ignoreExp bool) *MarsJSON.JSONObject {
	if _tokenStr == "" {
		return nil
	}

	var _decrypted []byte
	var _err error

	// 嘗試使用私鑰解密 (RSA)
	if _j._PrivateKey != nil {
		_decrypted, _err = jwe.Decrypt([]byte(_tokenStr), jwe.WithKey(jwa.RSA_OAEP_256, _j._PrivateKey))
	}

	// 若 RSA 失敗或無私鑰，嘗試 AES
	if _err != nil && len(_j._SecretKey) > 0 {
		_decrypted, _err = jwe.Decrypt([]byte(_tokenStr), jwe.WithKey(jwa.DIRECT, _j._SecretKey))
	}

	if _err != nil {
		return nil
	}

	// 解析 Payload
	var _obj map[string]interface{}
	json.Unmarshal(_decrypted, &_obj)

	// 驗證有效期 (exp)
	if _exp, _ok := _obj["exp"].(float64); _ok {
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
func (_j *JWTProcessor) EncryptAESData(_id string) string {
	if _j._AESBlock == nil {
		return ""
	}
	_plaintext := []byte(_id)
	// PKCS7 Padding
	_padding := aes.BlockSize - (len(_plaintext) % aes.BlockSize)
	_padtext := append(_plaintext, bytes.Repeat([]byte{byte(_padding)}, _padding)...)

	_ciphertext := make([]byte, len(_padtext))
	_mode := cipher.NewCBCEncrypter(_j._AESBlock, _j._AES_IV)
	_mode.CryptBlocks(_ciphertext, _padtext)

	return base64.StdEncoding.EncodeToString(_ciphertext)
}

// -------------------------------------------------------------------------------------
func (_j *JWTProcessor) DecryptAESData(_data string) string {
	if _j._AESBlock == nil {
		return ""
	}
	_ciphertext, _ := base64.StdEncoding.DecodeString(_data)
	_mode := cipher.NewCBCDecrypter(_j._AESBlock, _j._AES_IV)
	_mode.CryptBlocks(_ciphertext, _ciphertext)

	// Unpadding
	_padding := int(_ciphertext[len(_ciphertext)-1])
	return string(_ciphertext[:len(_ciphertext)-_padding])
}
