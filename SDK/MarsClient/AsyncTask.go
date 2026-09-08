package MarsClient

import (
	"context"
	"fmt"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/MarsJSON"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Tools"
	"strings"
	"time"
)

// 以下相容設定僅可於初始化時修改。
var AsyncTaskTimeout = 5 * time.Minute
var AsyncTaskPollInterval = 2 * time.Second

// RunAsyncTask 保留舊入口，預設五分鐘總期限。
func (c *MarsClient) RunAsyncTask(id, service, api string, payload *MarsJSON.JSONObject, callback AsyncTaskCallback) {
	c.RunAsyncTaskWithContext(context.Background(), id, service, api, payload, callback)
}

// RunAsyncTaskWithContext 可取消等待與正在進行的 HTTP 請求。
// 結果通道送出一次結果後關閉；自訂 callback 必須自行保證返回。
func (c *MarsClient) RunAsyncTaskWithContext(ctx context.Context, id, service, api string, payload *MarsJSON.JSONObject, callback AsyncTaskCallback) <-chan error {
	result := make(chan error, 1)
	timeout, interval := AsyncTaskTimeout, AsyncTaskPollInterval
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	if interval <= 0 {
		interval = 2 * time.Second
	}
	body := "{}"
	if payload != nil {
		body = payload.ToString()
	}
	go func() {
		defer close(result)
		defer func() {
			if r := recover(); r != nil {
				result <- fmt.Errorf("非同步 callback 異常: %v", r)
			}
		}()
		if id == "" || api == "" || callback == nil || ctx == nil {
			result <- fmt.Errorf("非同步任務參數不完整")
			return
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		address := fmt.Sprintf("%s/services-async/run/%s/%s/%s", c.GetServerURL(), id, service, strings.ReplaceAll(api, "/", "+"))
		resp, err := Tools.HttpRequestContext(ctx, "POST", address, c.GetAuthToken(), "application/json", body, 0)
		if err != nil {
			result <- err
			return
		}
		if resp == "" {
			result <- fmt.Errorf("任務回應為空")
			return
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				result <- ctx.Err()
				return
			case <-ticker.C:
				address = fmt.Sprintf("%s/services-async/check/%s", c.GetServerURL(), id)
				resp, err = Tools.HttpRequestContext(ctx, "GET", address, c.GetAuthToken(), "", "", 0)
				if err != nil {
					result <- err
					return
				}
				if resp == "" {
					result <- fmt.Errorf("任務狀態為空")
					return
				}
				callback(resp)
				if MarsJSON.NewJSONObject(resp).OptBoolean("done", false) {
					result <- nil
					return
				}
			}
		}
	}()
	return result
}
