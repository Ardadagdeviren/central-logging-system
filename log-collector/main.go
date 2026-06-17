package main

import (
	"encoding/json"
	"fmt"
	"net/http" //gerekli kütüphaneleri yapay zekaya yaptırdım geri bakıcam
	"os"
	"time"
)

var logFile *os.File

type logVerisi struct {
	Mesaj  string    `json:"message"` //logu tanımladım
	Seviye string    `json:"level"`
	Zaman  time.Time `json:"timestamp"`
	Servis string    `json:"service"`
}

func main() {
	var err error

	logFile, err = os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		fmt.Println("Dosya açılırken hata çıktı:", err) //dosyayı oluşturdum varsa üzerine yazdım, izinleri verdim
		return
	}

	http.HandleFunc("/logs", logHandler) // logs portuna bakanı logHandlera yolladım

	fmt.Println("Log toplayici 8080 portunda baslatiliyor...")
	http.ListenAndServe(":8080", nil)

}

func logHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed) //Sadece POST ları aldım güvenlik amaçlı
		w.Write([]byte("Sadece post istekleri kabul edilir"))

		return
	}
	defer r.Body.Close() //açılan Bodyi kapattım

	var yeniLogVerisi logVerisi
	decoder := json.NewDecoder(r.Body) //bodyi okuyacak yeni decoder oluşturdum
	err := decoder.Decode(&yeniLogVerisi)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Gönderdiğiniz JSON formatı hatalı"))
		return
	}
	logSatiri := fmt.Sprintf("[%v] [%s] [%s] %s\n", yeniLogVerisi.Zaman, yeniLogVerisi.Servis, yeniLogVerisi.Seviye, yeniLogVerisi.Mesaj)
	//log satını string yazdım ilerde rahat okumak için

	fmt.Println(logSatiri)

	_, err = logFile.WriteString(logSatiri) //logFile dosyasına yazdım
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Log dosyaya yazilirken bir hata oluştu."))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Log başarılı bir şekilde alındı"))

}
