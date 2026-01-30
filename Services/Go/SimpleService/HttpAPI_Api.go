package main

import (
	"net/http"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
)

// -------------------------------------------------------------------------------------
type HttpAPI_API struct{}

// -------------------------------------------------------------------------------------
func (_h *HttpAPI_API) Process(_w http.ResponseWriter, _r *http.Request, _jwt *MarsJSON.JSONObject, _path []string, _params *MarsJSON.JSONObject, _body string) []byte {

	_cmd := _path[len(_path)-1]

	switch _cmd {

	case "hello":
		return []byte("hello")

	}

	return nil
}

// -------------------------------------------------------------------------------------
