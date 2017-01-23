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
	"mime"
)

var pending = make(map[string]bool)
var mtx sync.Mutex

type request struct {
	ID          string      `json:"id"`
	Dimensions  []dimension `json:"dimensions"`
	FileFormats []string    `json:"fileFormats"`
	// FIXME this can be removed once metadata api etc is working
	S3URL string `json:"s3url"`
}
type dimension struct {
	ID      string   `json:"id"`
	Options []string `json:"options"`
}
type createdResponse struct {
	ID string `json:"id"`
}
type statusResponse struct {
	ID     string                `json:"id"`
	Status string                `json:"status"`
	Files  []*fileStatusResponse `json:"files"`
}
type fileStatusResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

var BindAddr = ":20100"

const jsonMediaType = "application/json"
const contentTypeHeader = "Content-Type"

func main() {
	if v := os.Getenv("BIND_ADDR"); len(v) > 0 {
		BindAddr = v
	}

	log.Namespace = "dp-dd-job-creator-api-stub"

	router := pat.New()
	alice := alice.New(
		corsHandler,
		timeout.Handler(10*time.Second),
		log.Handler,
		requestID.Handler(16),
	).Then(router)

	router.Post("/job", createHandler)
	router.Options("/job", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
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
	contentType, _, err := mime.ParseMediaType(req.Header.Get(contentTypeHeader))
	if err != nil || contentType != jsonMediaType {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		// Copy of the error message produced by the real API
		w.Write([]byte(`{"timestamp":1484652342702,"status":415,"error":"Unsupported Media Type",` +
			`"exception":"org.springframework.web.HttpMediaTypeNotSupportedException","message":"Content type '` +
			contentType + `' not supported","path":"/job"}`))
		return
	}

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
		<-time.NewTimer(time.Second * 4).C
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
		Files: []*fileStatusResponse{
			{Name: "example.csv", Status: "Pending"},
			//{Name: "example.xls", Status: "Pending"},
			//{Name: "example.json", Status: "Pending"},
		},
	}

	log.Debug("test", log.Data{"response": response})

	if _, ok := pending[id]; !ok {
		response.Status = "Complete"
		for _, v := range response.Files {
			v.Status = "Complete"
			v.URL = "https://www.ons.gov.uk"
		}
	}

	log.Debug("test", log.Data{"response": response})

	b, err := json.Marshal(&response)
	if err != nil {
		w.WriteHeader(500)
		w.Write(b)
		log.ErrorR(req, err, nil)
		return
	}

	log.Debug("test", log.Data{"response": string(b)})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(b)
}

func corsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		h.ServeHTTP(w, req)
	})
}
