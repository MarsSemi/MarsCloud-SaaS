package main

import (
	"net/http"
	"time"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
)

// -------------------------------------------------------------------------------------
type HttpAPI_API struct{}

// -------------------------------------------------------------------------------------
func (_h *HttpAPI_API) Process(_w http.ResponseWriter, _r *http.Request, _jwt *MarsJSON.JSONObject, _path []string, _params *MarsJSON.JSONObject, _body string) []byte {

	_cmd := _path[len(_path)-1]

	if _jwt != nil {
		// With Avalible Auth Token

		_usrID := _jwt.OptString("iss", "")
		_userGroup := _jwt.OptString("group", "")

		switch _cmd {

		case "hello":
			return []byte("hello : " + _usrID + " @ " + _userGroup + " > " + time.Now().Format("[01/02] 15:04:05"))

		}

	}

	switch _cmd {

	case "hello":
		return []byte("hello : " + time.Now().Format("[01/02] 15:04:05"))

	}

	return nil
}

// -------------------------------------------------------------------------------------
