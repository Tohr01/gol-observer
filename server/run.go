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
	"strings"
)

type LogFiles struct {
	Logs []LogFile `json:"logFiles"`
}

type LogFile struct {
	LogName     string `json:"logName"`
	LogPath     string `json:"logPath"`
	LogEndpoint string `json:"logEndpoint"`
	LogChannel  chan string
	Clients     map[*websocket.Conn]bool
}

var upgrader = websocket.Upgrader{}
var logFiles LogFiles
var logFilesJson []byte
var mux *http.ServeMux

func main() {
	jsonFile, openErr := os.Open("./log_files.json")
	if openErr != nil {
		panic(openErr)
	}
	defer jsonFile.Close()

	byteValue, readErr := io.ReadAll(jsonFile)
	if readErr != nil {
		panic(readErr)
	}

	unmarshalErr := json.Unmarshal(byteValue, &logFiles)
	if unmarshalErr != nil {
		panic(unmarshalErr)
	}

	mux = http.NewServeMux()
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	setupWebsockets()

	logFilesMap := make(map[string]string)
	for _, logConfig := range logFiles.Logs {
		logFilesMap[logConfig.LogEndpoint] = logConfig.LogName
	}
	logFilesJson, _ = json.Marshal(logFilesMap)

	mux.HandleFunc("/available-logs", handleAvailableLogs)

	handler := cors.Default().Handler(mux)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://127.0.0.1:5500"},
	})

	handler = c.Handler(handler)

	err := http.ListenAndServe(":6666", handler)
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

	logFile.Clients[ws] = true
	defer delete(logFile.Clients, ws)

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
				delete(logFile.Clients, client)
			}
		}
	}
}

func tailWatch(logFile *LogFile) {
	cmd := exec.Command("tail", "-f", logFile.LogPath)
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
		logFile.LogChannel <- line
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
