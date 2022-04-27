package main

import (
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

//добавить функцию, которая будет считывать это все говно из конфиг файла

var ipAddress = ""
var iterationsNumber = ""
var targetDir = ""
var logFolder = ""
var maxSize float64
var flagNetworkCheck bool

/*
readConfig() reading some keys and values from config file named "recorder-file-handler.yaml"
*/
func readConfig() {
	//configYamlFile, readErr := ioutil.ReadFile("/home/jupyter" + "/" + "recorder-file-handler.yaml")
	configYamlFile, readErr := ioutil.ReadFile("recorder-file-handler.yaml")
	if readErr != nil {
		panic(readErr)
	}
	data := make(map[string]string)

	unmarshErr := yaml.Unmarshal(configYamlFile, &data)
	if unmarshErr != nil {
		panic(unmarshErr)
	}
	ipAddress = data["ipAddress"]
	iterationsNumber = data["iterationsNumber"]
	targetDir = data["targetDir"]
	logFolder = data["logFolder"]
	maxSize, _ = strconv.ParseFloat(data["maxSize"], 64)
	flagNetworkCheck, _ = strconv.ParseBool(data["flagNetworkCheck"])

}

/*
func for logging events. logFolder - folder CreateOpenWriteRead logfile, logLevel - "info", "panic",
"fatal", logMessage - phrase'in'log
*/
func logger(logFolder string, logLevel string, logMessage string) {
	file, err := os.OpenFile(logFolder+"/"+"logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {

		log.Fatalf("error opening file: %v", err)

	}
	log.SetOutput(file)
	switch logLevel {
	case "info":
		log.Printf("%s: %s\n", logLevel, logMessage)
	case "fatal":
		log.Fatalf("%s: %s\n", logLevel, logMessage)
	case "panic":
		log.Panicf("%s: %s\n", logLevel, logMessage)
	default:
		log.Printf("%s - %s", logLevel, "is not a logger level!")
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger(logFolder, "fatal", "cannot close the log file!")
		}
	}(file)
}

/*checking availability of address of private network, determined by input var ipAddress, number
of check iterations determined by input var iterationsNumber. Connection's check provided by unix
"ping". If private network is unavailable - func rebooting the linux device */

func checkinPrivateNetwork(iterationsNumber string, ipAddress string) {
	//я не нашел лучшего способа проверить интернет-соединение для приватной сети
	//в которой закрыты порты, net.Dial работает только при указании порта,
	// поэтому в этой ситуации его применять не получится, хотя, она работает на порядок быстрее,
	// нежели обычный пинг
	cmd := exec.Command("ping", "-c "+iterationsNumber, ipAddress)
	_, err := cmd.Output()
	if err != nil {
		logger(logFolder, "info", "no private network connection, rebooting device")
		reboot := exec.Command("reboot")
		err := reboot.Run()
		if err != nil {
			logger(logFolder, "panic", "no way to reboot!")
			return
		}
	}
}

/*
counting size of all files situated in targetDir, returns size in Mibs in float64 val SizeCount and the
oldest file in targetDir
*/
func dirSizeTheOldestFile(rootDir string) (float64, string) {
	var sizeCount int64
	//пока не понял, как сделать без указания начального значения даты создания файла
	//по умолчанию я ставлю заведомо бОльшую дату, с которой
	//будут сравниваться даты
	//создания файлов
	var dateFile = time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
	var fileName = ""

	err := filepath.WalkDir(rootDir, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}
		size, _ := file.Info()
		if !file.IsDir() {
			sizeCount = sizeCount + size.Size()
			if size.ModTime().Before(dateFile) {
				dateFile = size.ModTime()
				fileName = file.Name()
			}
		}

		return nil
	})
	if err != nil {
		return 0, ""
	}
	return float64(sizeCount / 1000 / 1000), fileName
}

/*
function checks input size, compares with input val maxSize, if size > maxSize, removes filename,
(in loop removes until size < maxSize)
*/
func deletingOldFiles(maxSize float64, size float64, fileName string) {
	if size > maxSize {
		deleteOldFile := os.Remove(targetDir + "/" + fileName)
		if deleteOldFile != nil {
			logger(logFolder, "panic", "cannot remove file")
		}
		//прикрутить лог с инфо и именем удаленного файла
		logger(logFolder, "info", fileName+" was removed")
	}
}

func main() {
	//чтение параметров из конфиг-файла
	readConfig()
	//начало сессии
	defer logger(logFolder, "info", "session closed")
	logger(logFolder, "info", "session started")

	for {
		time.Sleep(5 * time.Second)
		//определяем размер указанной в конфиге директории и самый старый файл в ней
		size, name := dirSizeTheOldestFile(targetDir)
		//если размер директории превышает установленный пользователем в конфиге, то удаляется самый
		//старый файл
		deletingOldFiles(maxSize, size, name)
		//проверка соединения к впн сети
		checkinPrivateNetwork(iterationsNumber, ipAddress)
	}
}
