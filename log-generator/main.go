package main

import (
	"bytes"
	"encoding/json" //gerekli kütüphaneleri yapay zekaya yaptırdım geri bakıcam
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

type LogVerisi struct {
	Mesaj  string    `json:"message"`
	Seviye string    `json:"level"`
	Zaman  time.Time `json:"timestamp"`
	Servis string    `json:"service"`
}

func main() { //Hazır bilgiler burda ,belki ileride bu dizilere yapay zekayı bağlayıp kendisine log bilgilerini üretirebilirim log çeşidimiz artsın diye
	servisler := []string{"auth-service", "payment-service", "user-profile"}
	seviyeler := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	mesajlar := []string{"Kullanıcı giriş yaptı", "Veritabanı bağlantısı yavaşladı", "Ödeme başarısız oldu!"}

	for {
		randServis := rastgeleVeriSec(servisler)
		randSeviye := rastgeleVeriSec(seviyeler)
		randMesaj := rastgeleVeriSec(mesajlar)
		randZaman := time.Now()
		fmt.Print("Log üretici başlatıldı...")

		yeniLog := LogVerisi{
			Mesaj:  randMesaj,
			Seviye: randSeviye,
			Zaman:  randZaman,
			Servis: randServis,
		}

		jsonDilindeVeri, err := json.Marshal(yeniLog) //marshal json diline çeviriyor unmarshal geri koda dönüştürüyor
		if err != nil {
			fmt.Print("Log verilerini jsona çeviremedi")
			return
		}

		url := "http://log-collector-container:8080/logs"
		contentType := "application/json"
		body := bytes.NewBuffer(jsonDilindeVeri) //burayı tam anlamadım bir daha bakıcam yapay zekaya yaptırdım
		cevap, err := http.Post(url, contentType, body)

		if err != nil {
			fmt.Println("post isteği atarken sorun oldu")
		}
		cevap.Body.Close()

		time.Sleep(2 * time.Second)
	}

}

func rastgeleVeriSec(dizi []string) string { //Dizideki verileri rastgele loglara yerleştiren func
	uzunluk := len(dizi)

	randIndex := rand.IntN(uzunluk)

	return dizi[randIndex]

}
