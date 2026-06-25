package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type logVerisi struct {
	ID     int       `json:"id"`
	Mesaj  string    `json:"message" binding:"required"`
	Seviye string    `json:"level" binding:"required"`
	Zaman  time.Time `json:"timestamp"`
	Servis string    `json:"service" binding:"required"`
}

func main() {
	var err error

	dsn := "log_user:log_password@tcp(mysql-db:3306)/log_db?parseTime=true"

	for i := 0; i < 30; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping() //db ye ping atıyoruz bağlantı kurulmuş mu check ettik
		}
		if err == nil {
			break
		}
		fmt.Printf("MySQL bağlantısı tekrar deneniyor %d/30 %v\n", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("MySQL'e bağlanılamadı: %v", err)
	}
	defer db.Close()

	fmt.Println("mysql bağlantısı başarılı.")

	// Tablo yoksa oluştur
	tabloOlustur := `CREATE TABLE IF NOT EXISTS application_logs (
		id INT AUTO_INCREMENT PRIMARY KEY,
		service_name VARCHAR(255) NOT NULL,
		log_level VARCHAR(50) NOT NULL,
		message TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := db.Exec(tabloOlustur); err != nil {
		log.Fatalf("Tablo oluşturulurken hata: %v", err)
	}
	fmt.Println("application_logs tablosu hazır.")

	// Gin motorunu başlatıyoruz.
	// gin.Default(): Logger + Recovery ile gelir (her isteği terminale basar).
	// gin.New(): sadece istediğimiz middleware'leri ekleriz — GET /get-logs loglarını gizleyebiliriz.
	// gin.New() kullanıyoruz: gin.Default()'un aksine Logger middleware'i yok.
	// Kendi filtreleyici logger'ımızı ekliyoruz: sadece POST isteklerini göster,
	// GET /get-logs gibi polling isteklerini terminali kirletmesin.
	r := gin.New()
	r.Use(gin.Recovery()) // sunucu çökmelerini yakalar (hava yastığı)
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/get-logs"}, // bu endpoint'in loglarını bastırma
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

	sorgu := "INSERT INTO application_logs (service_name,log_level,message) VALUES (?,?,?)"
	_, err := db.Exec(sorgu, yeniLogVerisi.Servis, yeniLogVerisi.Seviye, yeniLogVerisi.Mesaj)

	if err != nil {
		fmt.Print("Veritabanına logu yazarken hata oluştu")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "log veritabanına yazılırken hata oluştu"})
		return
	}
	fmt.Printf("[KAYDEDİLDİ] Servis: %s | Seviye: %s | Mesaj: %s\n", yeniLogVerisi.Servis, yeniLogVerisi.Seviye, yeniLogVerisi.Mesaj)

	c.JSON(http.StatusCreated, gin.H{"message": "Log başarılı bir şekilde alındı"})
}

func getLogsHandler(c *gin.Context) {

	sorgu := "SELECT id ,service_name,log_level,message,created_at FROM application_logs ORDER BY created_at DESC" // ORBER BY:en güncel olandan eski olana doğru sıralı getir   DESC : azalan sırada
	satir, err := db.Query(sorgu)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Veritabanında log okunurken hata ile karşılaştık"})
		return
	}
	defer satir.Close()

	var loglar []logVerisi

	for satir.Next() { //NEXT.() imleci bir sağa kaydır diyormuş okuncak bir şey kalamdığında false döndürüp döngüden çıkıyormuş.
		var temp logVerisi
		if err := satir.Scan(&temp.ID, &temp.Servis, &temp.Seviye, &temp.Mesaj, &temp.Zaman); err != nil { //rows'dakileri tempin içine yazıyoruz
			fmt.Printf("Satır okunurken hata oluştu:%v", err)
			continue
		}

		loglar = append(loglar, temp)

	}
	c.JSON(http.StatusOK, loglar)

	//kendime not:
	// Okunan veri zaten ham JSON byte dizisi olduğu için, c.JSON yerine c.Data kullanarak

	//charset=utf-8: dünyadaki tüm harf emojileri kapsar benim jsonumda türkçe karakterler olduğu için bunu belirtiyorum.
}
