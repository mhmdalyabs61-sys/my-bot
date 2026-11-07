package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func logMessage(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}

func getEmbeddedProxies() []string {
	return []string{
		"20.205.243.166:3128",
		"34.195.154.123:3128",
		"54.210.36.19:3128",
		"104.21.78.12:80",
		"172.67.150.11:80",
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

	apiURL := "https://discord.com/api/v10/users/@me/username"
	client := &http.Client{Timeout: 5 * time.Second}

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
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		logMessage(fmt.Sprintf("[!] خطأ في الاتصال باليوزر %s: %v", username, err))
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 201:
		logMessage(fmt.Sprintf("[+] مبروك! اليوزر متاح: %s", username))
	case 400:
		logMessage(fmt.Sprintf("[-] غير متاح: %s", username))
	case 429:
		logMessage("[!] لمتد (Rate Limit)، جاري الانتظار...")
		time.Sleep(3 * time.Second)
	default:
		logMessage(fmt.Sprintf("[?] حالة غير معروفة (%d) لليوزر: %s", resp.StatusCode, username))
	}
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logMessage("[!] خطأ: لم يتم العثور على التوكن في متغيرات البيئة (DISCORD_TOKEN).")
		return
	}

	logMessage("=== بدء أداة الفحص على Railway ===")
	
	// حلقة مستمرة للفحص عشان الأداة ما توقف وتستمر بتوليد وفحص اليوزرات
	for {
		usernames := generateRandomUsernames(100)
		proxies := getEmbeddedProxies()

		var wg sync.WaitGroup
		for _, u := range usernames {
			wg.Add(1)
			go checkUsernameWithToken(u, proxies, token, &wg)
			time.Sleep(150 * time.Millisecond)
		}
		wg.Wait()
		
		logMessage("[*] انتهت الدفعة، جاري توليد دفعة جديدة...")
		time.Sleep(2 * time.Second)
	}
}
