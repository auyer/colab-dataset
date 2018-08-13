// Package main controls all features of the Voting Helper Web App and API.
// colab-dataset is a simple website to help calssify images accordignly.
package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/auyer/colab-dataset/config"
	"github.com/auyer/colab-dataset/db"
	"github.com/dgraph-io/badger"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/color"
	"golang.org/x/crypto/acme/autocert"
)

// confFlag stores the flags available when calling the program from the command line.
var confFlag = flag.String("config", "./config.json", "PATH to Configuration File. See docs for example config.")

var builddb = flag.Bool("builddb", false, "use this flag if DB shoud be built")

// database stores the pointer for the voting databse
var database *badger.DB

// counterdb stores the pointer for the vote counter databse
var counterdb *badger.DB

// databasesize stores the amount of items found while scanning the folder
var databasesize int

// staticBuilder function reads through the provided directory and populates the databases
func staticBuilder(dir string, dbpointer *badger.DB, counterdb *badger.DB) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {

		if f.IsDir() {
			go staticBuilder(dir+"/"+f.Name(), dbpointer, counterdb)
			log.Println(color.Blue("[BUILDDB]") + " Navigating into folder " + f.Name())
		} else {
			db.InsertResource(dir+"/"+f.Name(), 0, dbpointer)
			db.InsertResource(dir+"/"+f.Name(), 0, counterdb)
			log.Println(color.Blue("[BUILDDB]") + " Inserted " + f.Name())
		}
	}
}

func main() {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	flag.Parse()
	err := config.ReadConfig(*confFlag)
	if err != nil {
		log.Print("Error reading configuration file")
		log.Print(err.Error())
		return
	}
	server.Debug, _ = strconv.ParseBool(config.ConfigParams.Debug)

	server.Logger.SetOutput(config.LogFile)
	log.SetOutput(config.LogFile)

	// Database Loading

	database, err = db.Init(config.ConfigParams.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	counterdb, err = db.Init(config.ConfigParams.DatabasePath + ".count")
	if err != nil {
		log.Fatal(err)
	}
	defer counterdb.Close()
	if *builddb {
		log.Println(color.Red("[WORKING]") + "Building database")
		staticBuilder("."+config.ConfigParams.StaticFolder, database, counterdb)
		log.Println(color.Green("[DONE]") + "Database Built.")
	}
	databasesize = db.CountDBSize(database)
	log.Println(color.Green(strconv.Itoa(databasesize)) + " entries in the Database")

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
		var vote db.Vote
		err := c.Bind(&vote)
		if err != nil {
			server.Logger.Info(err)
			return c.String(http.StatusBadRequest, err.Error())
		}
		if vote.Vote == "true" {
			// log.Print("POSITIVE VOTE")
			db.UpdateResource(vote.Key, 1, counterdb)
			db.UpdateResource(vote.Key, 1, database)
			return c.String(200, " ")
		} else if vote.Vote == "false" {
			// log.Print("NEGATIVE VOTE")
			db.UpdateResource(vote.Key, 1, counterdb)
			db.UpdateResource(vote.Key, -1, database)
			return c.String(200, " ")
		}
		return c.String(500, " ")
	})
	server.POST("/api/unvote/", func(c echo.Context) error {
		var vote db.Vote
		err := c.Bind(&vote)
		if err != nil {
			server.Logger.Info(err)
			return c.String(http.StatusBadRequest, err.Error())
		}
		if vote.Vote == "true" {
			// log.Print("Undoing POSITIVE VOTE")
			db.UpdateResource(vote.Key, -1, counterdb)
			db.UpdateResource(vote.Key, -1, database)
			return c.String(200, " ")
		} else if vote.Vote == "false" {
			// log.Print("Undoing NEGATIVE VOTE")
			db.UpdateResource(vote.Key, -1, counterdb)
			db.UpdateResource(vote.Key, 1, database)
			return c.String(200, " ")
		}
		return c.String(500, " ")
	})

	server.GET("/api/getkey/", func(c echo.Context) error {
		value, err := db.GetSortedKey(counterdb)
		if err != nil {
			server.Logger.Info(err.Error())
			return c.String(http.StatusNotFound, err.Error())
		}
		return c.String(http.StatusAccepted, value) //c.Request().Host+
	})
	server.GET("/api/getTotalSize/", func(c echo.Context) error {
		// log.Println(databasesize)
		return c.String(http.StatusAccepted, strconv.Itoa(databasesize)) //c.Request().Host+
	})

	server.GET("/api/results/", func(c echo.Context) error {
		var countedList []db.VoteIntAmt
		value, err := db.GetCurrentVotes(database)
		if err != nil {
			server.Logger.Info(err.Error())
			return c.String(http.StatusNotFound, err.Error())
		}
		counter, err := db.GetCurrentVotes(counterdb)
		if err != nil {
			server.Logger.Info(err.Error())
			return c.String(http.StatusNotFound, err.Error())
		}
		for index, item := range value {
			if item.Key == counter[index].Key {
				countedList = append(countedList, db.VoteIntAmt{item.Key, item.Vote, counter[index].Vote})
			} else {
				log.Fatal("Unmatching Database")
			}
		}
		return c.JSON(http.StatusAccepted, countedList) //c.Request().Host+
	})

	if config.AutoTLS {
		server.AutoTLSManager.Cache = autocert.DirCache("./cert/")
		log.Printf("Serving Auto %s on address => %s", color.Green("HTTPS"), color.Green(config.ConfigParams.HttpsAddress))
		go func() {
			if err := server.StartAutoTLS(config.ConfigParams.Adress + ":" + config.ConfigParams.HttpsAddress); err != nil {
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
	} else if config.TLSEnabled {
		log.Printf("Serving %s on address => %s", color.Green("HTTPS"), color.Green(config.ConfigParams.HttpsAddress))
		go func() {
			if err := server.StartTLS(config.ConfigParams.HttpsAddress, config.ConfigParams.TLSCertLocation, config.ConfigParams.TLSKeyLocation); err != nil {
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
		log.Printf("Serving %s on address => %s", color.Red("HTTP"), color.Green(config.ConfigParams.HttpAddress))
		go func() {
			// if err := server.StartAutoTLS(config.ConfigParams.HttpAddress); err != nil {
			if err := server.Start(config.ConfigParams.Adress + ":" + config.ConfigParams.HttpAddress); err != nil {
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
