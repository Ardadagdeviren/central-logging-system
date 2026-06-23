package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai" //burayı yapay zekaya yaptrıdım dönüp bakıcam
	"google.golang.org/api/option"
)

type LogVerisi struct {
	Mesaj  string    `json:"message"`
	Seviye string    `json:"level"`
	Zaman  time.Time `json:"timestamp"`
	Servis string    `json:"service"`
}

type generativeLogs struct { //yapay zekaya her log için istek atmıyorum 1 istekte 100 log çekiyorum.
	Logs []LogVerisi `json:"logs"`
}

func main() {

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Gemini api keyi yazmamışsın")
	}

	ctx := context.Background() // kendiem not: context :Go dilinde context, mikroservisler arası veya API'lerle yapılan uzak ağ isteklerinde (Network Requests) isteklerin ömrünü, zaman aşımlarını (timeout) ve iptal sinyallerini yönetmek için kullanılan standart bir araçtır.
	//benim kodda bulunmasının sebebi gemini nin standart olarak istemesi.
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal("Gemini istemcisi başlamadı %v", err)

	}
	defer client.Close()                                //açılan bağlantıları kapatmayı alışkanlık haline getir.
	model := client.GenerativeModel("gemini-3.5-flash") //gemini-2.5-flash daha ucuz ve daha hızlı

	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"logs": {
				Type:        genai.TypeArray,
				Description: "Üretilen 100 adet sıralı ve mantıklı sistem logu", //tarif
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"message": {Type: genai.TypeString, Description: "Mikroservis mimarisine uygun, KESİNLİKLE TÜRKÇE yazılmış, gerçekçi bir olay veya hata mesajı."},
						"level":   {Type: genai.TypeString, Description: "DEBUG, INFO ,WARN ,ERROR seviyelerinden biri olsun sadece ,başka bir seviye asla ekleme"},
						"service": {Type: genai.TypeString, Description: "auth-service, payment-service, user-profile servislerin biri veya mantıklı yeni bir mikroservis adı"},
					},
					Required: []string{"message", "level", "service"}, //zorunluluk koşuyorsun.Bunlar bulunmassa çöpe at.
				},
			},
		},
		Required: []string{"logs"}, //gene bir zorunluluk koşuyorsun logs isimli string bir dizi yoksa çöpe at.
	}
	model.SystemInstruction = genai.NewUserContent(genai.Text("Sen bir backend sistem log üreticisisin. Senden istenen senaryoya göre birbirini tetikleyen, mantıklı ardışık loglar üretirsin. Ürettiğin tüm 'message' (mesaj) içerikleri KESİNLİKLE TÜRKÇE olmalıdır. İngilizce terimler yerine (örneğin 'Connection timeout' yerine 'Bağlantı zaman aşımı') Türkçe teknik ifadeler kullan."))
	//model.SystemInstruction: yapay zekanın karakteri verirsin.  genai.NewUserContent:yapay zekanın diline çevirir.   genai.Text:yapay zekaya vericeğin metni giremek için
	var logHavuzu []LogVerisi
	for {
		if len(logHavuzu) == 0 {
			fmt.Println("\nHavuzda log kalmadı. Gemini'dan 100 adet yeni log talep ediliyor...") //logların tekrar etmemesi için üretilen loglar bitince yeni log isteği yolluyorum
			prompt := `Bir e-ticaret platformunda büyük bir 'Kara Cuma (Black Friday) İndirim Gecesi' senaryosu işlet. Üreteceğin 100 adet sıralı log, saat tam 23:55'te başlasın ve gece yarısı indirimlerin patladığı, sunucuların zorlandığı, siber saldırganların açık aradığı ve ekibin müdahale ettiği dinamik bir olay örgüsünü (hikayeyi) anlatsın.

Bu 100 logluk paket içinde şu kurallara KESİNLİKLE uy:
1. ÇEŞİTLİLİK: Paket içinde en az 15 adet DEBUG, 40 adet INFO, 25 adet WARN ve 20 adet ERROR/CRITICAL logu bulunmak zorunda.
2. ZİNCİRLEME OLAYLAR: Loglar bağımsız olmasın, birbirini tetiklesin. Örneğin:
   - [INFO] Kampanya başladı, trafik anlık %500 arttı.
   - [DEBUG] Ödeme servisi veritabanı bağlantı havuzu sınırına yaklaşıyor (38/40).
   - [WARN] Kupon kodu servisinde yavaşlama tespit edildi (Response time: 1800ms).
   - [ERROR] Ödeme servisi SQL bağlantı zaman aşımı hatası verdi!
   - [INFO] Yedek veritabanı replikasyonu devreye alındı.
3. GERÇEKÇİLİK VE AKSİYON: Kullanıcı girişleri, sepete ürün eklemeler, stok tükenme uyarıları, sahte kupon denemesi yapan siber saldırganların engellenmesi (WAF bloklamaları) ve sistem mimarisinin (auth, payment, user-profile, stock-service) birbiriyle konuşmasını tamamen TÜRKÇE teknik terimlerle yansıt.

Bana bu 100 adet ardışık ve heyecanlı olay örgüsüne sahip log listesini şemaya uygun şekilde teslim et.`

			resp, err := model.GenerateContent(ctx, genai.Text(prompt)) //ctx:internet falan koparsa sonsuza kadar beklememesini sağlar. genai.text promptu yapay zekanın anlayacağı bir dile dönüştürür.
			if err != nil {
				fmt.Printf("Yapay zeka logu üretemedi 5 saniye sonra tekrar denenecek: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			//kendime not :Katmanlı Yapı Analojisi
			//Gemini'ın bize gönderdiği paketin hiyerarşik yapısı tam olarak şöyledir:

			//resp (En dış kutu): API'den gelen genel yanıt paketi.

			//Candidates (Adaylar Dizisi): Yapay zekanın ürettiği alternatif cevapların listesi. (Gemini bazen 1'den fazla alternatif cevap üretebilir).

			//Content (İçerik): Seçtiğimiz cevabın ana gövdesi.

			//Parts (Parçalar Dizisi): Cevabın içindeki elemanlar. (Cevap hem metin, hem görsel, hem de fonksiyon çağrısı içerebileceği için dizi olarak tutulur).
			//aşağıda kod çökmemesi için veri var mı kondtrolü yapıyoruz

			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
				jsonStr := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

				var log generativeLogs
				if err := json.Unmarshal([]byte(jsonStr), &log); err != nil {
					fmt.Printf("JSON okunamadı: %v", err)
					time.Sleep(45 * time.Second)
					continue
				}

				logHavuzu = log.Logs
				fmt.Printf("Yeni %d  loglar alındı ", len(logHavuzu))

			}
		}
		if len(logHavuzu) > 0 {
			gecerliLog := logHavuzu[0]
			logHavuzu = logHavuzu[1:]
			gecerliLog.Zaman = time.Now() //zamanı türkiye saaatine göre ayarla

			jsonDilindeVeri, err := json.Marshal(gecerliLog)
			if err != nil {
				fmt.Println("log verilerini jsona çeviremedi", err)
			}

			url := "http://log-collector-service:8080/collect"
			contentType := "application/json"
			body := bytes.NewBuffer(jsonDilindeVeri) //veriyi parça parça http ye göre aktarıyor newbuffer

			cevap, err := http.Post(url, contentType, body)
			if err != nil {
				fmt.Println("Collectora post isteği atarken sorun oldu :", err)

				time.Sleep(2 * time.Second)
				continue
			}
			cevap.Body.Close()
			fmt.Printf("[GÖNDERİLDİ] Servis: %s | Seviye: %s | Mesaj: %s\n", gecerliLog.Servis, gecerliLog.Seviye, gecerliLog.Mesaj)
		}
		time.Sleep(7 * time.Second)

		//Kendime not:
		// Bu satırın yaptığı iş tam olarak şudur: Kovadaki suyun (statik bayt dizisinin) altına bir musluk (Buffer) takmak.
		//bytes.NewBuffer, senin oluşturduğun o statik []byte verisini içine alır.
		//Onu ağ üzerinden parça parça okunmaya hazır, akışkan bir yapıya dönüştürür.
		//Artık elindeki body değişkeni sıradan bir veri değil, http.Post fonksiyonunun ucunu bağlayıp veriyi hüpletilerek çekebileceği bir okuma borusudur.

	}
}
