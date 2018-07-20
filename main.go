// Package main controls all features of the FastGate API Gateway.
// FasGate API Gateway is an application that  built with the Golang language
/*
You can run FastGate with the following command:
```
    fastgate -config ./path_to_config_file
```
  A sample to the configuration file can be found in config.model.json
 To manually register (and test) FastGate, Send a POST request to `yourip:yourport/fastgate/` with a JSON like follows:
```
{
  "address" : "https://yourEndpoint:8080"
  "uri"     : "/api/your_resource"
}
```
### Now send the desired request to `yourip:yourport/api/your_resource` and see it working !



*/
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/auyer/colab-dataset/config"
	"github.com/auyer/colab-dataset/db"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/color"
)

// confFlag stores the flags available when calling the program from the command line.
var confFlag = flag.String("config", "./config.json", "PATH to Configuration File. See docs for example config.")

const (
	version = "0.1.alpha"
	website = "github.com/auyer/fastgate/"
	banner  = "%s\nFast, light and Low Overhead API Gateway written in GO\n%s \nServing %s on port => %s"
	// logo built with http://www.patorjk.com/software and https://www.browserling.com/tools/utf8-encode
)

type Vote struct {
	Key  string `json:"Key"`
	Vote string `json:"Vote"`
}

func staticBuilder(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if f.IsDir() {
			staticBuilder(dir + "/" + f.Name())
		} else {

			_ = db.InsertResource(dir+"/"+f.Name(), 0)

			// fmt.Println(f.Name())
			// fmt.Println("TEST:")
			// fmt.Println(db.GetResourceValue(dir + "/" + f.Name()))
		}

	}
}

func main() {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	log.Printf("Starting FastGate APIGateway")
	flag.Parse()
	err := config.ReadConfig(*confFlag)
	if err != nil {
		fmt.Print("Error reading configuration file")
		log.Print(err.Error())
		return
	}
	server.Debug, _ = strconv.ParseBool(config.ConfigParams.Debug)
	if config.TLSEnabled {
		log.Printf(banner, color.Red("v"+version), color.Blue(website), color.Green("HTTPS"), color.Green(config.ConfigParams.HttpsPort))
	} else {
		log.Printf(banner, color.Red("v"+version), color.Blue(website), color.Red("HTTP"), color.Green(config.ConfigParams.HttpPort))
	}

	server.Logger.SetOutput(config.LogFile)
	log.SetOutput(config.LogFile)

	// Database Loading
	err = db.Init(config.ConfigParams.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.GetDB().Close()
	staticBuilder("." + config.ConfigParams.StaticFolder)
	db.DBSize = db.CountDBSize()

	server.Use(middleware.Logger())
	server.Use(middleware.Recover())
	server.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	// server.Use(middleware.Static(config.ConfigParams.LogLocation))
	server.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   ".",
		Browse: true,
		HTML5:  true,
	}))

	server.POST("/api/vote/", func(c echo.Context) error {
		var vote Vote
		err := c.Bind(&vote)
		if err != nil {
			server.Logger.Info(err)
			return c.String(http.StatusBadRequest, err.Error())
		}
		if vote.Vote == "true" {
			fmt.Print("POSITIVE VOTE")
			db.UpdateResource(vote.Key, 1)
			return c.String(200, " ")
		} else if vote.Vote == "false" {
			fmt.Print("NEGATIVE VOTE")
			db.UpdateResource(vote.Key, -1)
			return c.String(200, " ")
		}

		return c.String(500, " ")
	})

	server.GET("/api/getkey/", func(c echo.Context) error {
		// resource := c.Request().Header.Get("X-fastgate-resource")

		value, err := db.GetRandomKey()
		if err != nil {
			server.Logger.Info(err.Error())
			return c.String(http.StatusNotFound, err.Error())
		}
		return c.String(http.StatusAccepted, value) //c.Request().Host+
	})

	server.GET("/api/results/", func(c echo.Context) error {

		db.GetCurrentVotes()
		// value, err := db.GetCurrentVotes()
		// if err != nil {
		// 	server.Logger.Info(err.Error())
		// 	return c.String(http.StatusNotFound, err.Error())
		// }
		return c.String(http.StatusAccepted, " ") //c.Request().Host+
	})

	if config.TLSEnabled {
		go func() {
			if err := server.StartTLS(":"+config.ConfigParams.HttpsPort, config.ConfigParams.TLSCertLocation, config.ConfigParams.TLSKeyLocation); err != nil {
				server.Logger.Info("shutting down the server")
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 10 seconds.
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			server.Logger.Fatal(err)
		}
	} else {
		go func() {
			if err := server.Start(":" + config.ConfigParams.HttpPort); err != nil {
				server.Logger.Info("shutting down the server")
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 10 seconds.
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			server.Logger.Fatal(err)
		}
	}
}
