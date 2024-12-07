package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LogFiles struct {
	Logs []LogFile
}

type LogFile struct {
	LogName     string
	LogPath     string
	LogEndpoint string
	LogChannel  chan string
	Clients     map[*websocket.Conn]bool
	LastLines   []string
}

var upgrader = websocket.Upgrader{}
var logFiles LogFiles
var logFilesJson []byte
var mux *http.ServeMux

func main() {
	log.Println("Loading log files config")
	jsonFile, openErr := os.Open("./log_files.json")
	if openErr != nil {
		panic(openErr)
	}
	defer jsonFile.Close()

	byteValue, readErr := io.ReadAll(jsonFile)
	if readErr != nil {
		panic(readErr)
	}

	type LogFilesGlobs struct {
		Globs []string `json:"logFilesGlob"`
	}
	var logFilesGlobs LogFilesGlobs

	unmarshalErr := json.Unmarshal(byteValue, &logFilesGlobs)
	if unmarshalErr != nil {
		panic(unmarshalErr)
	}

	logFilesGlobs.Globs = removeDuplicateStr(logFilesGlobs.Globs)

	for _, logGlob := range logFilesGlobs.Globs {
		files, err := filepath.Glob(logGlob)
		if err != nil {
			log.Printf("Skipping %s\n", logGlob)
			continue
		}
		for _, path := range files {
			filebase := filepath.Base(path)
			suffix := filepath.Ext(path)
			filebase = strings.TrimSuffix(filebase, suffix)
			logFiles.Logs = append(logFiles.Logs, LogFile{
				LogName: filebase,
				LogPath: path,
			})
		}
	}

	log.Println("Creating new http serve mux")
	mux = http.NewServeMux()
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	log.Println("Setting up websocket endpoints")
	setupWebsockets()

	logFilesMap := make(map[string]string)
	for _, logConfig := range logFiles.Logs {
		logFilesMap[logConfig.LogEndpoint] = logConfig.LogName
	}
	logFilesJson, _ = json.Marshal(logFilesMap)

	mux.HandleFunc("/available-logs", handleAvailableLogs)

	handler := cors.Default().Handler(mux)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://127.0.0.1:5500",
			"https://api.pixeldogs.de",
		},
	})

	handler = c.Handler(handler)

	log.Println("Listening on 127.0.0.1:8888")
	err := http.ListenAndServe(":8888", handler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleAvailableLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(logFilesJson)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func setupWebsockets() {
	for i := range logFiles.Logs {
		logConfig := &logFiles.Logs[i] // Use fresh instance
		logConfig.LogEndpoint = strings.ReplaceAll(logConfig.LogName, " ", "-")
		wsEndpoint := "/ws/" + logConfig.LogEndpoint
		logConfig.Clients = make(map[*websocket.Conn]bool)
		logConfig.LogChannel = make(chan string)

		log.Println("Registering ws endpoint with name " + logConfig.LogEndpoint)

		go tailWatch(logConfig) // Start log watching
		go handleSendWsMessages(logConfig)

		mux.HandleFunc(wsEndpoint, func(w http.ResponseWriter, r *http.Request) {
			handleConnections(w, r, logConfig)
		})
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request, logFile *LogFile) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	log.Println("New client connected to " + logFile.LogEndpoint)
	logFile.Clients[ws] = true
	defer delete(logFile.Clients, ws)

	for _, line := range logFile.LastLines {
		err := ws.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			log.Println("Error writing to client:", err)
			break
		}
	}

	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			break
		}
	}
}

func handleSendWsMessages(logFile *LogFile) {
	for logLine := range logFile.LogChannel {
		for client := range logFile.Clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(logLine))
			if err != nil {
				log.Println("Error writing to client:", err)
				err := client.Close()
				if err != nil {
					log.Println("Could not close client:", err)
					return
				}
				log.Println("Closing ws connection to client for log file" + logFile.LogName)
				delete(logFile.Clients, client)
			}
		}
	}
}

func tailWatch(logFile *LogFile) {
	const N = 20
	log.Println("Starting tail -F for log file at path " + logFile.LogPath)
	cmd := exec.Command("tail", "-F", logFile.LogPath)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(cmdReader)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		logFile.LogChannel <- line

		if len(logFile.LastLines) >= N {
			logFile.LastLines = logFile.LastLines[1:]
		}
		logFile.LastLines = append(logFile.LastLines, line)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

}

// https://stackoverflow.com/questions/66643946/how-to-remove-duplicates-strings-or-int-from-slice-in-go
func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
