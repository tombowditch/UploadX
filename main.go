package main

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"fmt"
	"os"
	"io"
	"gopkg.in/h2non/filetype.v1"
	"io/ioutil"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/elgs/gostrgen"
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

func main() {
	app := iris.New()

	err := godotenv.Load()
	if err != nil {
		panic("Cannot load .env file, please make sure it is created.")
	}

	app.Post("/upload", func(ctx context.Context) {
		logrus.Info("Handling upload request")
		key := ctx.PostValue("key")
		if key != os.Getenv("UPLOAD_KEY") {
			logrus.Error("Invalid application key provided")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.JSON(SimpleResponse{Success: false, Message: "Invalid application key"})
			return
		}

		file, _, err := ctx.FormFile("img")
		if err != nil {
			logrus.Error("No file was given")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.JSON(SimpleResponse{Success: false, Message: "Could not upload [no file]"})
			return
		}

		defer file.Close()
		randName := randString(6)

		out, err := os.OpenFile("./files/"+randName+".png", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			logrus.Error("Could not write the file")
			fmt.Println("Could't write the file, please make sure the 'files' directory is created.")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.JSON(SimpleResponse{Success: false, Message: "Could not upload [cf]"})
			return
		}
		defer out.Close()
		io.Copy(out, file)
		buf, _ := ioutil.ReadFile("./files/" + randName+".png")

		if filetype.IsImage(buf) {
			logrus.Info("Uploaded file and served response: " + randName)
			ctx.StatusCode(iris.StatusOK)
			ctx.JSON(Response{Success: true, Message: "Uploaded", Name: randName})
		} else {
			logrus.Error("Uploaded file was not an image")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.JSON(SimpleResponse{Success: false, Message: "Uploaded file is not an image"})
			os.Remove("./files/" + randName) // lol doesnt work "HaHa The process cannot access the file because it is being used by another process. :angry:
		}
	})

	app.Get("/{imgName}", func(ctx context.Context) {
		imgName := ctx.Params().Get("imgName")
		filePath := "./files/" + imgName + ".png"
		_, err := ioutil.ReadFile(filePath)
		if err != nil {
			logrus.Error("Serving file '" + imgName + "' - unknown file")
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.JSON(SimpleResponse{Success: false, Message: " Unknown file"})
			return
		}

		logrus.Info("Serving file '" + imgName + "' - success")
		ctx.StatusCode(iris.StatusOK)
		ctx.ServeFile(filePath, true)

	})

	app.Get("/", func(ctx context.Context) {
		logrus.Info("Serving index page")
		ctx.StatusCode(iris.StatusOK)
		ctx.JSON(SimpleResponse{Success: true, Message: os.Getenv("INDEX_PAGE_TEXT")})
	})

	app.Run(iris.Addr(os.Getenv("BIND_HOST") + ":" + os.Getenv("BIND_PORT")))
}

func randString(n int) string {
	r, e := gostrgen.RandGen(n, gostrgen.Lower|gostrgen.Upper, "", "")
	if e != nil {
		logrus.Error("Could not generate random string")
	}
	return r
}

