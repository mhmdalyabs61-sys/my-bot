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

// لستة بروكسيات مدمجة للاستخدام السريع
func getEmbeddedProxies() []string {
	return []string{
		"20.205.243.166:3128",
		"34.195.154.123:3128",
		"54.210.36.19:3128",
		"104.21.78.12:80",
		"172.67.150.11:80",
		"13.248.169.48:80",
		"54.230.156.40:80",
		"18.155.67.23:3128",
		"52.216.164.2:80",
		"3.125.18.90:8080",
	}
}

// توليد يوزرات رباعية عشوائية تلقائياً (حروف وأرقام)
func generateRandomUsernames(count int) []string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	var usernames []string
	seen := make(map[string]bool)

	rand.Seed(time.Now().UnixNano())
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

// فحص وتصفية البروكسيات السليمة فقط
func filterWorkingProxies(rawProxies []string) []string {
	fmt.Println("[*] جاري فحص وتصفية البروكسيات...")
	var working []string
	var wg sync.WaitGroup
	mu := &sync.Mutex{}

	for _, p := range rawProxies {
		wg.Add(1)
		go func(proxyStr string) {
			defer wg.Done()

			proxyURL, err := url.Parse("http://" + proxyStr)
			if err != nil {
				return
			}

			client := &http.Client{
				Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
				Timeout:   3 * time.Second,
			}

			resp, err := client.Get("https://httpbin.org/ip")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				mu.Lock()
				working = append(working, proxyStr)
				mu.Unlock()
			}
		}(p)
	}
	wg.Wait()
	fmt.Printf("[+] تم العثور على %d بروكسي شغال.\n", len(working))
	return working
}

func getRandomProxy(proxies []string) string {
	if len(proxies) == 0 {
		return ""
	}
	return proxies[rand.Intn(len(proxies))]
}

// فحص اليوزر باستخدام التوكن والبروكسي
func checkUsernameWithToken(username string, proxies []string, token string, wg *sync.WaitGroup) {
	defer wg.Done()

	apiURL := "https://discord.com/api/v10/users/@me/username"
	client := &http.Client{Timeout: 5 * time.Second}

	proxyURLStr := getRandomProxy(proxies)
	if proxyURLStr != "" {
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
		fmt.Printf("[+] مبروك! اليوزر متاح: %s\n", username)
		saveResult("available.txt", username)

	case 400:
		fmt.Printf("[-] مستخدم أو غير متاح: %s\n", username)

	case 429:
		fmt.Printf("[!] تم رصد حظر مؤقت (Rate Limit) على اليوزر: %s\n", username)
		time.Sleep(2 * time.Second)

	case 401, 403:
		fmt.Println("[x] تحذير: التوكن غير صالح أو يتطلب تحقق (Captcha).")
		time.Sleep(3 * time.Second)
	}
}

func saveResult(filename string, username string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
    f.WriteString(username + "\n")
}

func main() {
	fmt.Println("=== أداة توليد وفحص اليوزرات التلقائية (Go) ===")

	// ضع توكن الحساب الوهمي هنا
	token := "YOUR_USER_TOKEN_HERE"
	if token == "YOUR_USER_TOKEN_HERE" {
		fmt.Println("[!] الرجاء وضع توكن الحساب الوهمي في متغير token داخل الكود.")
		return
	}

	// توليد 50 يوزر رباعي عشوائي تلقائياً (تقدر تغير العدد لو تبي أكثر)
	usernames := generateRandomUsernames(50)
	fmt.Printf("[*] تم توليد %d يوزر رباعي عشوائي بنجاح!\n", len(usernames))

	// تصفية البروكسيات
	workingProxies := filterWorkingProxies(getEmbeddedProxies())
	if len(workingProxies) == 0 {
		fmt.Println("[-] لم يتم العثور على بروكسي نشط.")
		return
	}

	fmt.Printf("[*] بدء الفحص...\n")

	var wg sync.WaitGroup
	limiter := make(chan struct{}, 5)

	for _, user := range usernames {
		wg.Add(1)
		limiter <- struct{}{}

		go func(u string) {
			defer func() {
				<-limiter
				wg.Done()
			}()
			checkUsernameWithToken(u, workingProxies, token, &wg)
		}(user)

		time.Sleep(200 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println("[*] انتهى الفحص!")
}
