package main

import (
    "fmt"
    "time"
	"net/http"
    // İleride buraya "os", "net/http" ve "encoding/json" da eklenecek
)

type logVerisi{
	Mesaj  string 'json:"message"'
	Seviye string 'json:"level"'
	Zaman  time.Time 'json:"timestamp"'
	Servis string 'json:"service"'
}


func main(){

	http.HandleFunc("/logs", LogHandler)


	func logHandler(w http.ResponseWriter ,r *http.Request){

		if r.Method != http.MethodPost{
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Sadece post istekleri kabul edilir"))

			return
		}
		defer r.Body.Close()

		var yeniLogVerisi logVerisi
		decoder :=json.NewDecoder(r.body)
		err :=decoder.Decode(&logVerisi)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Gönderdiğiniz JSON formatı hatalı"))
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Log başarılı bir şekilde alındı"))



	}


}