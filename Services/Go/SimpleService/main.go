package main

//-------------------------------------------------------------------------------------
import (
	"runtime"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsService"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Tools"
)

// -------------------------------------------------------------------------------------
func RunService() {

	_service := MarsService.Create("agent.properties", &MyCloudService{Counter: 0})

	_service.AddRestfulAPI("/api", &HttpAPI_API{})

	//單純啟動 HTTP Server，執行微服務功能
	//_service.HTTPServer.Start()

	//向 MarsCloud 服務註冊、並啟動 MarsClient 與 服務相關 連線
	_service.RegistryServerInfo("1.0.0", "pack", true)
	_service.Start()

	// 保持程式不結束
	select {}
}

// -------------------------------------------------------------------------------------
//
// -------------------------------------------------------------------------------------
func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	Tools.EnableUncaughtExceptionHandler("My Service", 3, func() { Tools.Log.Print(Tools.LL_Info, "System Error !!") })
	Tools.Log.SetDisplayLevel(Tools.LL_Info)

	RunService()
}

//-------------------------------------------------------------------------------------
