package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var logFile *os.File

type logVerisi struct {
	Mesaj  string    `json:"message" binding:"required"`
	Seviye string    `json:"level" binding:"required"`
	Zaman  time.Time `json:"timestamp"`
	Servis string    `json:"service" binding:"required"`
}

func main() {
	var err error

	// app.json dosyasını açıyoruz veya yoksa oluşturuyoruz
	logFile, err = os.OpenFile("app.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Dosya açılırken hata çıktı: %v", err)
	}
	defer logFile.Close()

	// Gin motorunu başlatıyoruz.
	// gin.Default(): Logger + Recovery ile gelir (her isteği terminale basar).
	// gin.New(): sadece istediğimiz middleware'leri ekleriz — GET /get-logs loglarını gizleyebiliriz.
	// gin.New() kullanıyoruz: gin.Default()'un aksine Logger middleware'i yok.
	// Kendi filtreleyici logger'ımızı ekliyoruz: sadece POST isteklerini göster,
	// GET /get-logs gibi polling isteklerini terminali kirletmesin.
	r := gin.New()
	r.Use(gin.Recovery()) // sunucu çökmelerini yakalar (hava yastığı)
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/get-logs"}, // bu endpoint'in loglarını bastır
	}))

	//javascripttin istek atarken sorun yaşamaması için gerekli izin güncellemeleri yaptığım kısım:
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent) //tarayıcılar bazen kontrol amaçlı options isteği atıyormuş.Burda o isteğe karşı evet gelebilirsin onaylıyorum diyoruz ve options isteğini kapatıyoruz.Abort ile
			return
		}
		c.Next() //bu next ne yapıyor anlamadım bir daha bak
		//istek geldi mi ilk başta üsteki func e gidiyor bu nedenle alttaki collect ve get-logs a bakamıyor.Next() ise bu isteği alıp bir daha yolluyor ama alt tarafataki kodlar için
	})

	r.POST("/collect", logHandler)
	r.GET("/get-logs", getLogsHandler)

	fmt.Println("Log toplayici 8080 portunda baslatiliyor...")

	// Sunucuyu 8080 portundan ayağa kaldırıyoruz
	if err := r.Run(":8080"); err != nil { //portu dinliyor httpdeki listenserve gibi
		log.Fatalf("Bilgisayarın portunda problem çıktı: %v", err)
	}
}

func logHandler(c *gin.Context) {
	var yeniLogVerisi logVerisi

	//Kendime not:   c.ShouldBindJSON hem Body'yi okur (stream), hem çözümler (Decode), hem de struct'a bağlar.
	if err := c.ShouldBindJSON(&yeniLogVerisi); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Generatorden gelen json formatı veya zorunlu alanlar hatalı"})
		return
	}

	jsonBytes, err := json.Marshal(yeniLogVerisi) //
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Log verisi işlenirken bir hata oluştu"})
		return
	}

	logSatiri := string(jsonBytes) + "\n"
	fmt.Print(logSatiri) // terminale yazdır

	// app.json dosyasına yazıyoruz
	_, err = logFile.WriteString(logSatiri)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Log dosyaya yazılırken bir hata oluştu"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Log başarılı bir şekilde alındı"})
}

func getLogsHandler(c *gin.Context) {
	data, err := os.ReadFile("app.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Log dosyası okunurken hata ile karşılaştık"})
		return
	}

	//kendime not:
	// Okunan veri zaten ham JSON byte dizisi olduğu için, c.JSON yerine c.Data kullanarak

	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
	//charset=utf-8: dünyadaki tüm harf emojileri kapsar benim jsonumda türkçe karakterler olduğu için bunu belirtiyorum.
}
