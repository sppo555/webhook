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
	WebhookPath    = "/webhook"    // Webhook path
	HeartcheckPath = "/heartcheck" // Heartbeat check path
	Port           = ":8080"       // Server port
)

var (
	TGAPIToken   = os.Getenv("TG_API_TOKEN") // Telegram Bot API token
	TGChatID     = os.Getenv("TG_CHAT_ID")   // Telegram chat ID
	URLPath      = os.Getenv("URL_PATH")     // Path configuration
	PathHandlers = make(map[string]http.HandlerFunc)
)

func main() {
	initLogger()

	http.HandleFunc(WebhookPath, handleWebhook)       // Set the handler for the webhook path
	http.HandleFunc(HeartcheckPath, handleHeartcheck) // Set the handler for the heartbeat check path

	paths := append(strings.Split(URLPath, ","), WebhookPath, HeartcheckPath)
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path != "" {
			filterKeys := strings.Split(os.Getenv(strings.ToUpper(path)+"_FILTER_KEY"), ",")
			PathHandlers["/"+path] = createDynamicHandler(filterKeys, path) // Set the handler for dynamic paths
		}
	}

	http.HandleFunc("/", handleNotFound) // Set the handler for not found paths

	for path, handler := range PathHandlers {
		http.HandleFunc(path, handler) // Set handlers for all paths
	}

	log.Printf("Webhook server started on port %s", Port)
	log.Fatal(http.ListenAndServe(Port, nil)) // Start the server
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

	message := processJSONData(data) // Process JSON data and convert it to a message

	sendToTelegram(WebhookPath, message) // Send the message to Telegram

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

		message := processJSONDataWithFilterKeys(data, filterKeys) // Process JSON data with filter keys and convert it to a message

		sendToTelegram(path, message) // Send the message to Telegram

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
		message += processJSONKeyValue(key, value, 0) // Convert JSON key-value pairs to a message
	}
	return message
}

func processJSONDataWithFilterKeys(data map[string]interface{}, filterKeys []string) string {
	message := ""
	for _, key := range filterKeys {
		if val, ok := data[key]; ok {
			message += processJSONKeyValue(key, val, 0) // Convert JSON data based on filter keys to a message
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
			message += processJSONKeyValue(nestedKey, nestedValue, level+1) // Recursively process nested JSON data
		}
	case []interface{}:
		message += "\n"
		for _, nestedValue := range v {
			message += processJSONKeyValue(key, nestedValue, level+1) // Recursively process array-type JSON data
		}
	default:
		message += fmt.Sprintf("%v\n", v) // Process other value types and format the value as a message
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

	resp, err := http.Post(apiURL, "application/json", strings.NewReader(payload)) // Send a POST request to the Telegram API
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
