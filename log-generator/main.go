package main

import (
	//"encoding/json"
	//"fmt"
	//"net/http"
	//"os"
	"bytes"
	"encoding/json"
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

func main() {
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

		jsonDilindeVeri, err := json.Marshal(yeniLog)
		if err != nil {
			fmt.Print("Log verilerini jsona çeviremedi")
			return
		}

		url := "http://log-collector-container:8080/logs"
		contentType := "application/json"
		body := bytes.NewBuffer(jsonDilindeVeri)
		cevap, err := http.Post(url, contentType, body)

		if err != nil {
			fmt.Println("post isteği atarken sorun oldu")
		}
		cevap.Body.Close()

		time.Sleep(2 * time.Second)
	}

}

func rastgeleVeriSec(dizi []string) string {
	uzunluk := len(dizi)

	randIndex := rand.IntN(uzunluk)

	return dizi[randIndex]

}
