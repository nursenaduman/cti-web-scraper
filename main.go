package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// URL komut satırından alınacak
	if len(os.Args) < 2 {
		fmt.Println("Kullanim: go run main.go <url>")
		return
	}

	url := os.Args[1]

	// Çıktılar karışmasın diye site adına göre klasör 
	u, err := neturl.Parse(url)
	klasor := "site"
	if err == nil && u.Host != "" {
		klasor = u.Host
	}
	// Windows için sorunlu karakterleri temizle
	klasor = strings.Map(func(r rune) rune {
		switch r {
		case ':', '*', '?', '"', '<', '>', '|':
			return -1
		}
		return r
	}, klasor)

	_ = os.Mkdir(klasor, 0755)

	// HTML çekme
	html, err := fetchHTML(url)
	if err != nil {
		fmt.Println("HTML cekme hatasi:", err)
		return
	}

	_ = os.WriteFile(klasor+"/site_data.html", []byte(html), 0644)

	// Linkleri çek (ek puan)
	linkler, err := extractLinks(url)
	if err == nil {
		_ = os.WriteFile(klasor+"/links.txt", []byte(strings.Join(linkler, "\n")), 0644)
	}

	// Ekran görüntüsü al
	err = takeScreenshot(url, klasor+"/screenshot.png")
	if err != nil {
		fmt.Println("Screenshot alinamadi:", err)
		return
	}

	fmt.Println("Tamamlandi:", url)
}

func fetchHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP kodu %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func takeScreenshot(url, dosya string) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var img []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.FullScreenshot(&img, 90),
	)
	if err != nil {
		return err
	}

	return os.WriteFile(dosya, img, 0644)
}

func extractLinks(url string) ([]string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var sonuc []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &sonuc),
	)

	return sonuc, err
}

