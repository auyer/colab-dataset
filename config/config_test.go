package config

import (
	"fmt"
	"log"
	"os"
	"testing"
)

const (
	testConfigLogPath = "./test_server.config_test.go.log"
	testConfigPath    = "./testconfig.config_test.go.json"
	testConfig        = `{
		"Debug": "false",
		"LogLocation": "./test_server.config_test.go.log",
		"HttpAddress": "8888test",
		"HttpsAddress": "8443test",
		"TLSKeyLocation": "./tls_test.go.key",
		"TLSCertLocation": "./tls.test.go.pem",
		"DatabasePath" : "./test_fastgate.db",
		"StaticFolder" : "./static"
}`
)

func TestConfRead(t *testing.T) {
	//Creating Test Structure
	if _, err := os.Stat(testConfigPath); !os.IsNotExist(err) {
		err = os.Remove(testConfigPath)
		if err != nil {
			log.Fatal("Unable to clean Test Settings. Check for permissions.")
		}
	}
	// TEST with Default Configs
	err := ReadConfig(testConfigPath)
	if err != nil {
		t.Errorf("Unable to Read Configuration: " + err.Error())
	}
	tmp := ConfigParams
	fmt.Printf(tmp.Debug)
	if ConfigParams.Debug != "true" && ConfigParams.LogLocation != "os.Stdout" && ConfigParams.HttpAddress != "8080" && ConfigParams.HttpsAddress != "8443" && ConfigParams.TLSKeyLocation != "./devssl/server.key" &&
		ConfigParams.TLSCertLocation != "./devssl/server.pem" && ConfigParams.DatabasePath != "./fastgate.db" {
		t.Errorf("Default Configuration read wrongly.")
	}
	// CREATING TestFile
	fileConf, fileErr := os.Create(testConfigPath)
	if fileErr != nil {
		log.Fatal("Unable to create Test Settings. Check for permissions.")
	}
	fileConf.WriteString(testConfig)
	fileConf.Close()

	//TESTING with Config File
	err = ReadConfig(testConfigPath)
	if err != nil {
		t.Errorf("Unable to Read Configuration: " + err.Error())
	}
	if ConfigParams.TLSKeyLocation == "./tls_test.go.key" && ConfigParams.TLSCertLocation == "./tls.test.go.pem" {
		if _, err := os.Stat(ConfigParams.TLSCertLocation); !os.IsNotExist(err) {
			err = os.Remove(ConfigParams.TLSCertLocation)
			if err != nil {
				log.Fatal("Unable to clean Test Certificate. Check for permissions.")
			}
		}
		fileConf, fileErr = os.Create(ConfigParams.TLSCertLocation)
		if fileErr != nil {
			log.Fatal("Unable to create Test Certificate. Check for permissions.")
		}
		fileConf.WriteString("TEST")
		fileConf.Close()

		if _, err := os.Stat(ConfigParams.TLSKeyLocation); !os.IsNotExist(err) {
			err = os.Remove(ConfigParams.TLSKeyLocation)
			if err != nil {
				log.Fatal("Unable to clean Test Key. Check for permissions.")
			}
		}
		fileConf, fileErr = os.Create(ConfigParams.TLSKeyLocation)
		if fileErr != nil {
			log.Fatal("Unable to create Test Key. Check for permissions.")
		}
		fileConf.WriteString("TEST")
		fileConf.Close()
		//CREATE cert Test Files
	}
	if ConfigParams.Debug != "false" && ConfigParams.LogLocation != "./test_server.config_test.go.log" && ConfigParams.HttpAddress != "8888test" && ConfigParams.HttpsAddress != "88443test" && ConfigParams.DatabasePath != "./test_fastgate.db" && TLSEnabled != true {
		t.Errorf("Configuration read from file interpreted wrongly.")
	} else if _, err := os.Stat(ConfigParams.TLSCertLocation); os.IsNotExist(err) {
		t.Errorf("Unable to Create Log File at" + ConfigParams.LogLocation)
	}

	// CLEAN files
	err = os.Remove(testConfigPath)
	if err != nil {
		log.Fatal("Unable to clean Test ConfigFile. Check for permissions.")
	}
	err = os.Remove(ConfigParams.TLSKeyLocation)
	if err != nil {
		log.Fatal("Unable to clean Test Key. Check for permissions.")
	}
	err = os.Remove(ConfigParams.TLSCertLocation)
	if err != nil {
		log.Fatal("Unable to clean Test Certificate. Check for permissions.")
	}
	err = os.Remove(ConfigParams.LogLocation)
	if err != nil {
		log.Fatal("Unable to clean Test Log. Check for permissions.")
	}
}

func TestLogCreator(t *testing.T) {
	file := createLogfile(testConfigLogPath)
	if file == nil {
		t.Errorf("Unable to create logfile")
	} else {
		err := os.Remove(testConfigLogPath)
		if err != nil {
			t.Errorf("Unable to delete test logfile")
		}
	}
}
