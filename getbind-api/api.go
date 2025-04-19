package getbind_api

import (
	"fmt"
	"getbind2api/common"
	"getbind2api/common/config"
	logger "getbind2api/common/loggger"
	"getbind2api/cycletls"
	"github.com/gin-gonic/gin"
	"math/rand"
	"strings"
	"time"
)

const (
	baseURL      = "https://api.getbind.co"
	chatEndpoint = baseURL + "/chatbot/stream"
)

func MakeStreamChatRequest(c *gin.Context, client cycletls.CycleTLS, requestBody map[string]interface{}, cookie string, modelInfo common.ModelInfo) (<-chan cycletls.SSEResponse, error) {
	split := strings.Split(cookie, "=")
	if len(split) >= 2 {
		cookie = split[0]
	}

	// 生成随机边界字符串
	boundary := "----WebKitFormBoundary" + generateRandomString(16)

	// 构建multipart/form-data请求体
	var formData strings.Builder

	// 添加model字段
	formData.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	formData.WriteString("Content-Disposition: form-data; name=\"model\"\r\n\r\n")
	formData.WriteString(fmt.Sprintf("%s\r\n", requestBody["model"]))

	// 添加query字段
	formData.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	formData.WriteString("Content-Disposition: form-data; name=\"query\"\r\n\r\n")
	formData.WriteString(fmt.Sprintf("%s\r\n", requestBody["query"]))

	// 添加bot_id字段
	formData.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	formData.WriteString("Content-Disposition: form-data; name=\"bot_id\"\r\n\r\n")
	formData.WriteString(fmt.Sprintf("%s\r\n", requestBody["bot_id"]))

	// 添加session_id字段
	formData.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	formData.WriteString("Content-Disposition: form-data; name=\"session_id\"\r\n\r\n")
	formData.WriteString(fmt.Sprintf("%s\r\n", requestBody["session_id"]))

	// 添加user_id字段
	formData.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	formData.WriteString("Content-Disposition: form-data; name=\"user_id\"\r\n\r\n")
	formData.WriteString(fmt.Sprintf("%s\r\n", requestBody["user_id"]))

	// 添加files字段
	formData.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	formData.WriteString("Content-Disposition: form-data; name=\"files\"\r\n\r\n")
	formData.WriteString(fmt.Sprintf("%s\r\n", requestBody["files"]))

	// 如果有context字段，也添加
	if context, ok := requestBody["context"]; ok {
		formData.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		formData.WriteString("Content-Disposition: form-data; name=\"context\"\r\n\r\n")
		formData.WriteString(fmt.Sprintf("%s\r\n", context))
	}

	// 结束边界
	formData.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	headers := map[string]string{
		"accept":             "text/event-stream",
		"accept-language":    "zh-CN,zh;q=0.9,en;q=0.8",
		"content-type":       fmt.Sprintf("multipart/form-data; boundary=%s", boundary),
		"origin":             "https://copilot.getbind.co",
		"priority":           "u=1, i",
		"referer":            "https://copilot.getbind.co/",
		"sec-ch-ua":          "\"Google Chrome\";v=\"135\", \"Not-A.Brand\";v=\"8\", \"Chromium\";v=\"135\"",
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"macOS\"",
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"user-agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
		//"cookie":             cookie,
	}

	options := cycletls.Options{
		Timeout: 10 * 60 * 60,
		Proxy:   config.ProxyUrl, // 在每个请求中设置代理
		Body:    formData.String(),
		Method:  "POST",
		Headers: headers,
	}

	logger.Debug(c.Request.Context(), fmt.Sprintf("cookie: %v", cookie))

	sseChan, err := client.DoSSE(chatEndpoint, options, "POST")
	if err != nil {
		logger.Errorf(c, "Failed to make stream request: %v", err)
		return nil, fmt.Errorf("Failed to make stream request: %v", err)
	}
	return sseChan, nil
}

// 生成随机字符串，用于表单边界
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}
