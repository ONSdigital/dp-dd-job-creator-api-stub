package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/handlers/timeout"
	"github.com/ONSdigital/go-ns/log"
	"github.com/google/uuid"
	"github.com/gorilla/pat"
	"github.com/justinas/alice"
)

var pending = make(map[string]bool)
var mtx sync.Mutex

type request struct {
	ID          string      `json:"id"`
	Dimensions  []dimension `json:"dimensions"`
	FileFormats []string    `json:"fileFormats"`
}
type dimension struct {
	ID      string   `json:"id"`
	Options []string `json:"options"`
}
type createdResponse struct {
	ID string `json:"id"`
}
type statusResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

var BindAddr = ":20100"

func main() {
	if v := os.Getenv("BIND_ADDR"); len(v) > 0 {
		BindAddr = v
	}

	log.Namespace = "dp-dd-job-creator-api-stub"

	router := pat.New()
	alice := alice.New(
		timeout.Handler(10*time.Second),
		log.Handler,
		requestID.Handler(16),
	).Then(router)

	router.Post("/job", createHandler)
	router.Get("/job/{id}", statusHandler)

	log.Debug("Starting server", log.Data{
		"bind_addr": BindAddr,
	})

	server := &http.Server{
		Addr:         BindAddr,
		Handler:      alice,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error(err, nil)
		os.Exit(2)
	}
}

func createHandler(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		log.ErrorR(req, err, nil)
		return
	}

	var input request
	err = json.Unmarshal(b, &input)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		log.ErrorR(req, err, nil)
		return
	}

	response := createdResponse{
		ID: uuid.New().String(),
	}

	b, err = json.Marshal(&response)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		log.ErrorR(req, err, nil)
		return
	}

	mtx.Lock()
	pending[response.ID] = true
	mtx.Unlock()
	go func() {
		<-time.NewTimer(time.Second * 10).C
		mtx.Lock()
		defer mtx.Unlock()
		delete(pending, response.ID)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(b)
}

func statusHandler(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")
	response := statusResponse{
		ID:     id,
		Status: "Pending",
	}

	if _, ok := pending[id]; !ok {
		response.Status = "Complete"
		response.URL = "https://www.ons.gov.uk"
	}

	b, err := json.Marshal(&response)
	if err != nil {
		w.WriteHeader(500)
		w.Write(b)
		log.ErrorR(req, err, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(b)
}
