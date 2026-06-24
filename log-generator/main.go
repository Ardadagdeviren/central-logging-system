package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

// fallbackLogUret: Gemini api çalışmadığında hazır şablonlardan rastgele log üretir
// Bu sayede kota bittiğinde bile dashboard boş kalmaz
func fallbackLogUret() []LogVerisi {
	type sablon struct {
		mesaj  string
		seviye string
		servis string
	}

	sablonlar := []sablon{

		{"Kullanıcı oturumu başarıyla açıldı (kullanici_id: 4821)", "INFO", "auth-service"},
		{"Kampanya sayfası yüklendi, anlık ziyaretçi sayısı: 12.450", "INFO", "gateway-service"},
		{"Sepete ürün eklendi (urun_id: SKU-7823, adet: 2)", "INFO", "cart-service"},
		{"Ödeme işlemi başarıyla tamamlandı (sipariş_no: ORD-99821)", "INFO", "payment-service"},
		{"Kullanıcı profil bilgileri güncellendi", "INFO", "user-profile"},
		{"Yeni üye kaydı oluşturuldu (e-posta doğrulanmamış)", "INFO", "auth-service"},
		{"Stok güncelleme işlemi tamamlandı (234 ürün)", "INFO", "stock-service"},
		{"E-posta bildirimi kuyruğa alındı (şablon: siparis_onayi)", "INFO", "notification-service"},
		{"CDN önbellek temizleme işlemi tamamlandı", "INFO", "cdn-service"},
		{"Veritabanı yedekleme işlemi başarıyla bitti (boyut: 2.3GB)", "INFO", "db-backup-service"},
		{"Kupon kodu başarıyla uygulandı (BLACKFRIDAY50)", "INFO", "coupon-service"},
		{"Kargo takip numarası oluşturuldu (TRK-887654)", "INFO", "shipping-service"},
		{"API istek sınırı güncellendi (yeni limit: 1000/dk)", "INFO", "gateway-service"},
		{"Önbellek isabet oranı: %94.2 (son 5 dakika)", "INFO", "cache-service"},

		{"Veritabanı bağlantı havuzu %85 doluluk oranına ulaştı (34/40)", "WARN", "payment-service"},
		{"Kupon servisi yanıt süresi yüksek (1850ms, eşik: 1000ms)", "WARN", "coupon-service"},
		{"Disk kullanımı %78'e ulaştı (/var/log bölümü)", "WARN", "monitoring-service"},
		{"Aynı IP'den kısa sürede 15 başarısız giriş denemesi", "WARN", "auth-service"},
		{"Stok miktarı kritik seviyeye düştü (urun_id: SKU-1122, kalan: 3)", "WARN", "stock-service"},
		{"Bellek kullanımı %82'ye çıktı (pod: payment-service-7b9d4)", "WARN", "monitoring-service"},
		{"SSL sertifika süresi dolmak üzere (kalan: 12 gün)", "WARN", "gateway-service"},
		{"Mesaj kuyruğu birikmesi tespit edildi (bekleyen: 847 mesaj)", "WARN", "notification-service"},
		{"Yavaş sorgu tespit edildi (2340ms, tablo: siparisler)", "WARN", "payment-service"},

		{"Ödeme ağ geçidi bağlantı zaman aşımı hatası (timeout: 30s)", "ERROR", "payment-service"},
		{"Veritabanı replikasyon gecikmesi kritik seviyede (gecikme: 45s)", "ERROR", "db-backup-service"},
		{"Kullanıcı oturum doğrulama hatası: JWT token süresi dolmuş", "ERROR", "auth-service"},
		{"Stok servisi yanıt vermiyor, devre kesici devreye girdi", "ERROR", "stock-service"},
		{"E-posta gönderim hatası: SMTP sunucusu reddetti (hata: 550)", "ERROR", "notification-service"},
		{"Sipariş oluşturma başarısız: veritabanı kilitlenmesi (deadlock)", "ERROR", "payment-service"},
		{"WAF engelleme: SQL enjeksiyonu denemesi tespit edildi (IP: 185.143.xx.xx)", "ERROR", "gateway-service"},
		{"Dosya yükleme hatası: maksimum boyut aşıldı (25MB > 10MB)", "ERROR", "user-profile"},

		{"Veritabanı bağlantı havuzu durumu: aktif=28, boşta=12, maks=40", "DEBUG", "payment-service"},
		{"HTTP isteği detayı: POST /api/v1/orders, süre: 234ms, boyut: 1.2KB", "DEBUG", "gateway-service"},
		{"Önbellek anahtarı oluşturuldu: user:4821:cart:v3", "DEBUG", "cache-service"},
		{"Goroutine sayısı: 847, bellek: 156MB, GC döngüsü: 12ms", "DEBUG", "monitoring-service"},
		{"JWT token yenileme işlemi başlatıldı (kullanici_id: 4821)", "DEBUG", "auth-service"},
		{"Kafka tüketici grubu offset bilgisi: partition=3, offset=128847", "DEBUG", "notification-service"},
		{"Redis PING yanıtı: PONG (gecikme: 0.3ms)", "DEBUG", "cache-service"},
		{"Bağlantı havuzu metrikleri: bekleme_suresi=2ms, edinme_suresi=0.5ms", "DEBUG", "payment-service"},
	}

	// 25 adet rastgele log seç
	loglar := make([]LogVerisi, 25) //make() ne işe yarıyor:25 tane logverisi alan slice dizi oluşturuyoruz bu sayede her eleman için bellekte yer ayırılıyor ve loglar[5] e falan ulaşıp içini doldurabiliyoruz
	for i := 0; i < 25; i++ {
		s := sablonlar[rand.Intn(len(sablonlar))]
		loglar[i] = LogVerisi{
			Mesaj:  s.mesaj,
			Seviye: s.seviye,
			Servis: s.servis,
		}
	}

	fmt.Println("[FALLBACK] Gemini kullanılamıyor, hazır şablonlardan 25 log üretildi.")
	return loglar
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
	defer client.Close()                                     //açılan bağlantıları kapatmayı alışkanlık haline getir.
	model := client.GenerativeModel("gemini-2.5-flash-lite") //daha hafif model, kota dostu

	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type:        genai.TypeArray,
		Description: "Üretilen 25 adet sıralı ve mantıklı sistem logu", //tarif
		Items: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"message": {Type: genai.TypeString, Description: "Mikroservis mimarisine uygun, KESİNLİKLE TÜRKÇE yazılmış, gerçekçi bir olay veya hata mesajı."},
				"level":   {Type: genai.TypeString, Description: "DEBUG, INFO ,WARN ,ERROR seviyelerinden biri olsun sadece ,başka bir seviye asla ekleme"},
				"service": {Type: genai.TypeString, Description: "auth-service, payment-service, user-profile servislerin biri veya mantıklı yeni bir mikroservis adı"},
			},
			Required: []string{"message", "level", "service"}, //zorunluluk koşuyorsun.Bunlar bulunmassa çöpe at.
		},
	}
	model.SystemInstruction = genai.NewUserContent(genai.Text("Sen bir backend sistem log üreticisisin. Senden istenen senaryoya göre birbirini tetikleyen, mantıklı ardışık loglar üretirsin. Ürettiğin tüm 'message' (mesaj) içerikleri KESİNLİKLE TÜRKÇE olmalıdır. İngilizce terimler yerine (örneğin 'Connection timeout' yerine 'Bağlantı zaman aşımı') Türkçe teknik ifadeler kullan."))
	//model.SystemInstruction: yapay zekanın karakteri verirsin.  genai.NewUserContent:yapay zekanın diline çevirir.   genai.Text:yapay zekaya vericeğin metni giremek için
	var logHavuzu []LogVerisi
	retryBeklemeSuresi := 10 * time.Second // başlangıç bekleme süresi
	ardisikHataSayisi := 0                 // geminiye ardışık olarak 3 defa ulaşamassam geminiyi bırakıp hazır loglardan üretiyor
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
				ardisikHataSayisi++
				fmt.Printf("Gemini hata verdi (kota bitmiş olabilir): %v\n", err)

				//BU KISMI YAPAY ZEKAYA YAPTIRDIM TEKRAR BAKICAM!
				// 3 ardışık hatadan sonra fallback moduna geç
				if ardisikHataSayisi >= 3 {
					fmt.Println("\n  3 ardışık Gemini hatası! Fallback log üreticisine geçiliyor...")
					logHavuzu = fallbackLogUret()
					ardisikHataSayisi = 0                 // sayacı sıfırla, bir sonraki turda tekrar Gemini denenecek
					retryBeklemeSuresi = 45 * time.Second // bekleme süresini sıfırla
				} else {
					fmt.Printf("%.0f saniye sonra tekrar denenecek... (deneme: %d/3)\n", retryBeklemeSuresi.Seconds(), ardisikHataSayisi)
					time.Sleep(retryBeklemeSuresi)
					// Exponential backoff: her hatada bekleme süresini 2 katına çıkar (maks 5 dakika)
					if retryBeklemeSuresi < 60*time.Second {
						retryBeklemeSuresi *= 2
					}
				}
				continue
			}
			ardisikHataSayisi = 0 // Gemini başarılı oldu, hata sayacını sıfırla

			//kendime not :Katmanlı Yapı Analojisi
			//Gemini'ın bize gönderdiği paketin hiyerarşik yapısı tam olarak şöyledir:

			//resp (En dış kutu): API'den gelen genel yanıt paketi.

			//Candidates (Adaylar Dizisi): Yapay zekanın ürettiği alternatif cevapların listesi. (Gemini bazen 1'den fazla alternatif cevap üretebilir).

			//Content (İçerik): Seçtiğimiz cevabın ana gövdesi.

			//Parts (Parçalar Dizisi): Cevabın içindeki elemanlar. (Cevap hem metin, hem görsel, hem de fonksiyon çağrısı içerebileceği için dizi olarak tutulur).
			//aşağıda kod çökmemesi için veri var mı kondtrolü yapıyoruz

			if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 { //candidates :Adaylar/seçenkler demek
				jsonStr := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]) //burda geminiden gelenleri string olarak json formatında alıyoruz.

				if err := json.Unmarshal([]byte(jsonStr), &logHavuzu); err != nil { //unmarshal parametre olarak byte türü bekliyor bu nedenle string json jsonStr yi byte çeviriyotuz
					fmt.Printf("JSON okunamadı: %v", err)
					time.Sleep(45 * time.Second)
					continue
				}

				fmt.Printf("Yeni %d  loglar alındı \n", len(logHavuzu))
				retryBeklemeSuresi = 45 * time.Second // başarılı olunca bekleme süresini sıfırla

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
