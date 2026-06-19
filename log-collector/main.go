package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http" //gerekli kütüphaneleri yapay zekaya yaptırdım geri bakıcam
	"os"
	"time"
)

var logFile *os.File

type logVerisi struct {
	Mesaj  string    `json:"message"` //logu tanımladım.    //kendime not: "Sevgili Go, benim dilimin kuralları gereği bu alanın adı içeride büyük harfle başlayan Mesaj olmak zorunda. Ama sen bunu JSON'a çevirirken (Marshal yaparken) ya da gelen JSON'ı okurken (Decode ederken) dış dünyaya message olarak göster/ara."
	Seviye string    `json:"level"`
	Zaman  time.Time `json:"timestamp"`
	Servis string    `json:"service"`
}

func main() {
	var err error

	logFile, err = os.OpenFile("app.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		fmt.Println("Dosya açılırken hata çıktı:", err) //dosyayı oluşturdum varsa üzerine yazdım, izinleri verdim
		return
	}

	http.HandleFunc("/collect", logHandler)      // logs portuna bakanı logHandlera yolladım
	http.HandleFunc("/get-logs", getLogsHandler) // javascriptin get ile logları alacağı func

	fmt.Println("Log toplayici 8080 portunda baslatiliyor...")
	err = http.ListenAndServe(":8080", nil)
	//kendime not:Gerçek Hayattan Bir Benzetme 🏨
	//Bir otel açtığını düşün:

	//":8080": Otelinin sokaktaki kapı numarasıdır.

	//nil (DefaultServeMux): Resepsiyondaki görevlidir. Gelen müşteriye (isteğe) bakar, "Sen giriş mi yapacaksın? 1. kata (logHandler) geç. Sen çıkış mı yapacaksın? 2. kata (getLogsHandler) geç" der.

	//ListenAndServe: Otelin kapılarını müşterilere açmak ve resepsiyonisti işe başlatmaktır. Eğer kapı açılabilirse otel sonsuza kadar açık kalır; ama kapı kilitliyse veya arızalıysa otel daha açılmadan kapanır (err).

	if err != nil {
		log.Fatalf("Sunucu portunda problem çıktı: %v", err) //log.Fatalf ile hata yazılımcıya programı yapan adama hata döndürdürmektir baya kritk hatadır.Sistemin asla çalışamıyacağı direkt kapanması gereken durumlar içiin
		//Ama WriteHeader ile gönderilen hatalar kullanıcıya örnek vermek gerekirse karşı tarafa benim sistememe farklı formatta dosya gönderdi veya süslü parantezi eksik koydu vs.

	}

}

func logHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed) //Sadece POST ları aldım güvenlik amaçlı
		w.Write([]byte("Sadece post istekleri kabul edilir"))

		return
	}
	defer r.Body.Close() //açılan Bodyi kapattım

	var yeniLogVerisi logVerisi
	decoder := json.NewDecoder(r.Body)    //decoder:Akan json formatındaki veriyi goloangin diline çevirir.Http istekleri komple bi anda gelmez.Akan su gibi geldiği için json.marshal kullanamayız json.marshal sonu belli olan komple indirilmiş veriler içindir.
	err := decoder.Decode(&yeniLogVerisi) //burda da okunan veriyi yenilogverisine işliyoruz.Bu sayede gelen veri benim istedğim logVerisi türüne uygun mu onuda görmüş oluyoruz.Güvenlik sağlanmış oluyor.

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Gönderdiğiniz JSON formatı hatalı"))
		return
	}

	jsonBytes, err := json.Marshal(yeniLogVerisi)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("log dosyası oluşurken hata oluşt"))
	}

	logSatiri := string(jsonBytes) + "\n"
	fmt.Println(logSatiri)
	//logu satırlara ayırdım ,string olarak yazdım ilerde rahat okumak için

	_, err = logFile.WriteString(logSatiri) //logFile dosyasına yazdım.  baştaki altçizgi kısmına normalde doğru yazılan bayt sayısını veriyormuş ama gerek olmadığı için altçizgi koydum
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Log dosyaya yazilirken bir hata oluştu."))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Log başarılı bir şekilde alındı"))

}
func getLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") //erişim
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")

	//kendime not:
	//Neden application/json direkt json değil
	//MIME standartlarında veriler önce ana kategorilere ayrılır. Dünyadaki popüler ana kategorilerden bazıları şunlardır:

	//text/: İnsanlar tarafından doğrudan açılıp okunabilecek ham metinler için (Örn: text/html, text/css, text/plain).

	//image/: Resim formatları için (Örn: image/png, image/jpeg).

	//audio/ ve video/: Ses ve video içerikleri için.

	//  application/: Belirli bir program, algoritma veya uygulama tarafından işlenmesi/çözümlenmesi (parse edilmesi) gereken kurala bağlı veriler veya ikili (binary) dosyalar için.
	//Başlık geçersiz olunca tarayıcı "MIME Sniffing" denilen bir sürece girer. Yani gelen ham verinin ilk birkaç byte'ına bakarak onun ne olduğunu kendisi tahmin etmeye çalışır.

	//Eğer tahmin edemezse, gelen veriyi en temel/en düz format olan text/plain (düz metin) veya application/octet-stream (bilinmeyen indirilebilir dosya) olarak kabul eder.

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Javascripten gelen istek GET olmalı"))
		return
	}

	data, err := os.ReadFile("app.json") //burda neden app.jsonı okuyup javascripite gönderiyorum.Çünkü logFile ı oluşturuken üzerine eklenebilir ve sadece yazılabilir izinlerini verdim bu sebebple imleç logFileın sonunda olucaktır.Ben okumak istediğimde sondan okuyacaktır.Kod karmaşası olmaması için belleğimdeki app.json dosyasını os.ReadFile ile okuyorum.Bu tek hamlede açıp okuyor.Formatı neyse ona göre okuyor json sa byte olarak okuyup dataya kaydettim ,açtığı dosyayıda geri kapatıyor..
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("log dosyası okunurken hata ile karşılaştık"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
