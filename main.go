package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Gophers-FUTA/futa-tech-bot/client"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("An error occured while loading environment variables")
	}

	rt := mux.NewRouter()

	rt.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "server is up and running")
	})

	//listen to crc check
	rt.HandleFunc("/webhook/twitter", CrcCheck).Methods(http.MethodGet)
	//listen to webhook event
	rt.HandleFunc("/webhook/twitter", WebhookHandler).Methods(http.MethodPost)

	sv := &http.Server{
		Handler: rt,
		Addr:    ":8080",
	}

	if args := os.Args; len(args) > 1 && args[1] == "-register" {
		go client.RegisterWebhook()
	}

	fmt.Printf("server running at port %s", sv.Addr)
	log.Fatal(sv.ListenAndServe())
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	b, _ := ioutil.ReadAll(r.Body)

	//init webhook load object
	var load client.WebhookLoad

	err := json.Unmarshal(b, &load)
	if err != nil {
		fmt.Println("An error occured while reading request body: ", err.Error())
	}

	if len(load.TweetCreateEvent) < 1 || load.UserId == load.TweetCreateEvent[0].User.IdStr {
		return
	}

	_, err = client.SendTweet("@"+load.TweetCreateEvent[0].User.Handle+" Hello World", load.TweetCreateEvent[0].IdStr)

	if err != nil {
		fmt.Println("An error occured: ", err.Error())
	} else {
		fmt.Println("Tweet sent successfully!")
	}
}

func CrcCheck(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	token := r.URL.Query()["crc_token"]

	if len(token) < 1 {
		fmt.Fprintf(w, "crc token not sent")
		return
	}

	h := hmac.New(sha256.New, []byte(os.Getenv("CONSUMER_SECRET")))

	h.Write([]byte(token[0]))

	encoded := base64.StdEncoding.EncodeToString(h.Sum(nil))

	response := make(map[string]string)

	response["response_token"] = "sha256=" + encoded

	rsp, _ := json.Marshal(response)
	fmt.Fprintf(w, string(rsp))

}
