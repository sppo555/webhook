package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	WebhookPath    = "/webhook"      // Webhook 路徑
	HeartcheckPath = "/heartcheck"   // 心跳檢查路徑
	Port           = ":8080"         // 伺服器埠號
)

var (
	TGAPIToken   = os.Getenv("TG_API_TOKEN") // Telegram Bot API 金鑰
	TGChatID     = os.Getenv("TG_CHAT_ID")   // Telegram 聊天 ID
	URLPath      = os.Getenv("URL_PATH")     // 路徑設定
	PathHandlers = make(map[string]http.HandlerFunc)
)

func main() {
	initLogger()

	http.HandleFunc(WebhookPath, handleWebhook)     // 設定 Webhook 路徑的處理函式
	http.HandleFunc(HeartcheckPath, handleHeartcheck) // 設定心跳檢查路徑的處理函式

	paths := append(strings.Split(URLPath, ","), WebhookPath, HeartcheckPath)
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path != "" {
			filterKeys := strings.Split(os.Getenv(strings.ToUpper(path)+"_FILTER_KEY"), ",")
			PathHandlers["/"+path] = createDynamicHandler(filterKeys, path) // 設定動態路徑的處理函式
		}
	}

	http.HandleFunc("/", handleNotFound) // 設定找不到路徑的處理函式

	for path, handler := range PathHandlers {
		http.HandleFunc(path, handler) // 設定所有路徑的處理函式
	}

	log.Printf("Webhook server started on port %s", Port)
	log.Fatal(http.ListenAndServe(Port, nil)) // 啟動伺服器
}

func initLogger() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Failed to decode JSON payload", http.StatusBadRequest)
		return
	}

	message := processJSONData(data) // 處理 JSON 資料，轉換為訊息

	sendToTelegram(WebhookPath, message) // 將訊息傳送到 Telegram

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Webhook request processed")
}

func handleHeartcheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != HeartcheckPath {
		http.NotFound(w, r)
		log.Printf("404 Page Not Found: %s %s", r.Method, r.URL.Path)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Heartcheck request processed")
}

func createDynamicHandler(filterKeys []string, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid HTTP method", http.StatusMethodNotAllowed)
			return
		}

		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Failed to decode JSON payload", http.StatusBadRequest)
			return
		}

		message := processJSONDataWithFilterKeys(data, filterKeys) // 處理 JSON 資料，根據篩選鍵值轉換為訊息

		sendToTelegram(path, message) // 將訊息傳送到 Telegram

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Dynamic request processed")
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
	log.Printf("404 Page Not Found: %s %s", r.Method, r.URL.Path)
}

func processJSONData(data map[string]interface{}) string {
	message := ""
	for key, value := range data {
		message += processJSONKeyValue(key, value, 0) // 將 JSON 資料鍵值對轉換為訊息
	}
	return message
}

func processJSONDataWithFilterKeys(data map[string]interface{}, filterKeys []string) string {
	message := ""
	for _, key := range filterKeys {
		if val, ok := data[key]; ok {
			message += processJSONKeyValue(key, val, 0) // 根據篩選鍵值將 JSON 資料轉換為訊息
		}
	}
	return message
}

func processJSONKeyValue(key string, value interface{}, level int) string {
	indent := strings.Repeat("  ", level)
	message := fmt.Sprintf("%s%s: ", indent, key)
	switch v := value.(type) {
	case map[string]interface{}:
		message += "\n"
		for nestedKey, nestedValue := range v {
			message += processJSONKeyValue(nestedKey, nestedValue, level+1) // 遞迴處理巢狀 JSON 資料
		}
	case []interface{}:
		message += "\n"
		for _, nestedValue := range v {
			message += processJSONKeyValue(key, nestedValue, level+1) // 遞迴處理陣列型態的 JSON 資料
		}
	default:
		message += fmt.Sprintf("%v\n", v) // 處理其他值型態，將值格式化為訊息
	}
	return message
}

func sendToTelegram(path, message string) {
	if TGAPIToken == "" || TGChatID == "" {
		log.Println("Telegram API token or chat ID is missing.")
		return
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TGAPIToken)
	payload := fmt.Sprintf(`{"chat_id": "%s", "text": "%s"}`, TGChatID, message)

	resp, err := http.Post(apiURL, "application/json", strings.NewReader(payload)) // 發送 POST 請求到 Telegram API
	if err != nil {
		log.Printf("Failed to send message to Telegram: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respData, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Failed to send message to Telegram. Status code: %d\nResponse Body: %s", resp.StatusCode, respData)
		return
	}

	log.Printf("Message sent to Telegram for path: %s", path)
}

