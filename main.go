package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/elgs/gostrgen"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

type Response struct {
	Success bool
	Message string
	Name    string
}

type SimpleResponse struct {
	Success bool
	Message string
}

var peopleServed = 0
var successfulServed = 0
var unsuccessfulServed = 0

type StatsResponse struct {
	PeopleServed             int
	SuccessfulPeopleServed   int
	UnsuccessfulPeopleServed int
}

func upload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logrus.Info("Handling upload request")
	if r.Method == "POST" {
		r.ParseMultipartForm(10000000)
		key := r.Form.Get("key")

		if key != os.Getenv("UPLOAD_KEY") {
			json.NewEncoder(w).Encode(SimpleResponse{Success: false, Message: "Invalid upload key"})
			return
		}

		file, _, err := r.FormFile("img")

		if err != nil {
			json.NewEncoder(w).Encode(SimpleResponse{Success: false, Message: "Could not upload: " + err.Error()})
			return
		}

		defer file.Close()

		randName := randString(6)

		f, err := os.OpenFile(os.Getenv("UPLOAD_LOCATION")+"/"+randName+".png", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			json.NewEncoder(w).Encode(SimpleResponse{Success: false, Message: "Could not open new file"})
			return
		}
		defer f.Close()
		io.Copy(f, file)

		json.NewEncoder(w).Encode(Response{Success: true, Message: "File uploaded", Name: randName})

	} else {
		json.NewEncoder(w).Encode(SimpleResponse{Success: false, Message: "Invalid method"})
	}

}

func serveImage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	imgName := ps.ByName("imgname")

	if imgName == "stats" {
		// lol thanks httprouter
		stats(w, r, ps)
		return
	}
	peopleServed += 1

	filePath := os.Getenv("UPLOAD_LOCATION") + "/" + imgName + ".png"
	_, err := ioutil.ReadFile(filePath)
	if err != nil {
		unsuccessfulServed += 1
		json.NewEncoder(w).Encode(SimpleResponse{Success: false, Message: "Unknown file"})
		return
	}

	logrus.Info("serving " + imgName)
	successfulServed += 1

	http.ServeFile(w, r, filePath)
}

func stats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	json.NewEncoder(w).Encode(StatsResponse{PeopleServed: peopleServed, SuccessfulPeopleServed: successfulServed, UnsuccessfulPeopleServed: unsuccessfulServed})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Cannot load .env file, please make sure it is created.")
	}

	logrus.Info("starting...")

	r := httprouter.New()

	r.POST("/upload", upload)
	r.GET("/:imgname", serveImage)

	if err := http.ListenAndServe(os.Getenv("BIND_HOST")+":"+os.Getenv("BIND_PORT"), r); err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

}

func randString(n int) string {
	r, e := gostrgen.RandGen(n, gostrgen.Lower|gostrgen.Upper, "", "")
	if e != nil {
		logrus.Error("Could not generate random string")
	}
	return r
}
