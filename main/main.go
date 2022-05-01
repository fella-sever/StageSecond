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

// vals for setting up parsed info
var ipAddress = ""        // set addrr for pingmachine
var iterationsNumber = "" // how many "ping -c"
var targetDir = ""        //sets up control dir for tracking dir size and old files
var logFolder = ""        //were'll be yr logs
var maxSize float64       //max capacity of control dir in float for convenience work with return types
var flagNetworkCheck bool //defines use networkchecker or not

/*
readConfig() reading some keys and values from config file named "recorder-file-handler.yaml"
*/
func readConfig() {
	var configYamlFile []byte
	var readConfErr error
	configYamlFile, readConfErr = ioutil.ReadFile("recorder-file-handler.yaml")
	if readConfErr != nil {
		//log.Printf("trying another path to parse config file")
		configYamlFile, readConfErr = ioutil.ReadFile("/usr/local/etc/recorder-file-handler.yaml")
		if readConfErr != nil {
			log.Printf("configFile: %v", readConfErr)
			os.Exit(1)
		}

	}
	parsedConfig := make(map[string]string) //map for containing data parsed from config file
	// parsing all info from config file to map
	parseErr := yaml.Unmarshal(configYamlFile, &parsedConfig)
	if parseErr != nil {
		log.Printf("configFile: %v", parseErr)
		os.Exit(1)
	}
	ipAddress = parsedConfig["ipAddress"]
	iterationsNumber = parsedConfig["iterationsNumber"]
	targetDir = parsedConfig["targetDir"]
	logFolder = parsedConfig["logFolder"]
	maxSize, _ = strconv.ParseFloat(parsedConfig["maxSize"], 64)
	flagNetworkCheck, _ = strconv.ParseBool(parsedConfig["flagNetworkCheck"])
	/* default config set
	# checking network section#
	# if true - starts checking network, if false - starts without connection check
	flagNetworkCheck: false
	# ipAddress defines ip of network without port for checking network connection (based on unix ping)
	ipAddress: 10.0.0.1
	# iterationsNumber sets number of ping iterations of cheking ip address
	iterationsNumber: "3"
	# recorder section #
	# directory in targetDir will be cheking out for it's size and possibly removing the oldest files
	targetDir: /home/jupyter/testFolder
	# max size of targetDir directory, if dir's size > maxSize, the oldest file in targetDir will be removed
	maxSize: 250
	# logging #
	# directory for logfile storage
	logFolder: /home/jupyter
	*/

}

/*
func for logging events. logFolder - folder CreateOpenWriteRead logfile, logLevel - "info", "panic",
"fatal", logMessage - phrase'in'log
*/
func logger(logFolder string, logLevel string, logMessage string) {
	// open/create/append to existing file
	file, errFile := os.OpenFile(logFolder+"/"+"logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errFile != nil {

		log.Printf("configFile: %v", errFile)
		os.Exit(1)

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

func checkinPrivateNetwork(iterationsNumber string, ipAddress string, flagNetworkCheck bool) {
	//я не нашел лучшего способа проверить интернет-соединение для приватной сети
	//в которой закрыты порты, net.Dial работает только при указании порта,
	// поэтому в этой ситуации его применять не получится, хотя, она работает на порядок быстрее,
	// нежели обычный пинг
	if flagNetworkCheck {
		cmd := exec.Command("ping", "-c "+iterationsNumber, ipAddress)
		_, err := cmd.Output()
		if err != nil {
			logger(logFolder, "info", "no private network connection, rebooting device")
			reboot := exec.Command("reboot")
			err := reboot.Run()
			if err != nil {
				logger(logFolder, "info", "no way to reboot!")
				return
			}
		}
	}

}

/*
counting size of all files situated in targetDir, returns size in Mibs in float64 val SizeCount and the
oldest file in targetDir
*/
func dirSizeTheOldestFile(rootDir string) (float64, string) {
	var sizeCount int64
	// date for comparing date of files in dir. dateFile
	// was defined obviously later than any datefile
	// for right result
	var dateFile = time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
	var fileName = ""
	// parsing dir definded in rooDir for its size and old files
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
		logger(logFolder, "info", fileName+" was removed")
	}
}

func main() {
	//чтение параметров из конфиг-файла
	readConfig()
	//начало сессии
	logger(logFolder, "info", "session started with params: ")
	logger(logFolder, "info", "-- check network: "+strconv.FormatBool(flagNetworkCheck))
	logger(logFolder, "info", "-- target directory: "+targetDir)
	logger(logFolder, "info", "-- max size of target dir: "+
		strconv.FormatFloat(maxSize, 'f', -1, 64)+" Mib")

	for {
		time.Sleep(5 * time.Second)
		//определяем размер указанной в конфиге директории и самый старый файл в ней
		size, name := dirSizeTheOldestFile(targetDir)
		//если размер директории превышает установленный пользователем в конфиге, то удаляется самый
		//старый файл
		deletingOldFiles(maxSize, size, name)
		//проверка соединения к впн сети
		checkinPrivateNetwork(iterationsNumber, ipAddress, flagNetworkCheck)
	}
}
