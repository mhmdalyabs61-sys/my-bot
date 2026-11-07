package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// رابط الويب هوك الخاص بك لإرسال النتائج إليه
const webhookURL = "https://discord.com/api/webhooks/1568426822142656582/KFYMMVFpAgSEAm3E5A4PJ9WLpQoBCoq7SHukUTzaYiZcEVx_LBpdArWdhZhZOIr-o6Zy"

func logMessage(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}

// دالة لإرسال إشعار للويب هوك في ديسكورد
func sendToWebhook(content string) {
	payloadMap := map[string]string{"content": content}
	jsonValue, err := json.Marshal(payloadMap)
	if err != nil {
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func getEmbeddedProxies() []string {
	return []string{
		"20.205.243.166:3128",
		"34.195.154.123:3128",
		"54.210.36.19:3128",
	}
}

func generateRandomUsernames(count int) []string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	var usernames []string
	seen := make(map[string]bool)

	for len(usernames) < count {
		sb := strings.Builder{}
		for i := 0; i < 4; i++ {
			sb.WriteByte(chars[rand.Intn(len(chars))])
		}
		u := sb.String()
		if !seen[u] {
			seen[u] = true
			usernames = append(usernames, u)
		}
	}
	return usernames
}

func checkUsernameWithToken(username string, proxies []string, token string, wg *sync.WaitGroup) {
	defer wg.Done()

	apiURL := "https://discord.com/api/v10/users/@me"
	client := &http.Client{Timeout: 10 * time.Second}

	if len(proxies) > 0 {
		proxyURLStr := proxies[rand.Intn(len(proxies))]
		proxyURL, err := url.Parse("http://" + proxyURLStr)
		if err == nil {
			client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		}
	}

	payload := fmt.Sprintf(`{"username":"%s"}`, username)
	req, err := http.NewRequest("PATCH", apiURL, strings.NewReader(payload))
	if err != nil {
		return
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 201:
		successMsg := fmt.Sprintf("🎉 **مبروك! تم صيد يوزر متاح:** `%s`", username)
		logMessage(successMsg)
		// إرسال اليوزر المتاح مباشرة إلى روم ديسكورد عبر الويب هوك
		sendToWebhook(successMsg)
	case 400:
		logMessage(fmt.Sprintf("[-] غير متاح: %s", username))
	case 429:
		logMessage("[!] لمتد (Rate Limit)، جاري الانتظار...")
		time.Sleep(5 * time.Second)
	}
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logMessage("[!] خطأ: لم يتم العثور على التوكن في متغيرات البيئة (DISCORD_TOKEN).")
		return
	}

	logMessage("=== بدء أداة الفحص على Railway وإرسال النتائج عبر الويب هوك ===")
	
	for {
		usernames := generateRandomUsernames(50)
		proxies := getEmbeddedProxies()

		var wg sync.WaitGroup
		for _, u := range usernames {
			wg.Add(1)
			go checkUsernameWithToken(u, proxies, token, &wg)
			time.Sleep(200 * time.Millisecond)
		}
		wg.Wait()
		
		logMessage("[*] انتهت الدفعة، جاري توليد دفعة جديدة...")
		time.Sleep(2 * time.Second)
	}
}
