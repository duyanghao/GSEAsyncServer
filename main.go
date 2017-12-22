package main

import (
	"Configuration"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"handle"
	"net/http"
	"time"
)

// args
var (
	argPort              = flag.Int("port", 80, "The port to listen to for incoming HTTP requests.")
	argBindAddress       = flag.String("bind-address", "0.0.0.0", "The IP address on which to serve the --port (set to 0.0.0.0 for all interfaces).")
	argConfigPath        = flag.String("config-file-path", "", "The configuration file path for server.")
	argWorkChannelLen    = flag.Int("work-channel-len", 10, "The maximum limit on concurrent task requests")
	argProcessMaxTimeout = flag.Int("process-max-timeout", 30, "Connection timeout(s) for task requests.")
)

type Server struct {
	twClient *handle.TaskWork
}

var server Server

// create the server
func newServer() error {
	// configurate handle WORK_CHANNEL_LEN and PROCESS_MAX_TIMEOUT
	handle.WORK_CHANNEL_LEN = *argWorkChannelLen
	handle.PROCESS_MAX_TIMEOUT = time.Second * time.Duration(*argProcessMaxTimeout)
	// loading configuration
	glog.V(5).Infof("Starting Loading configuration: %s.", *argConfigPath)
	config, err := Configuration.LoadConfig(*argConfigPath)
	if err != nil {
		return err
	}
	glog.V(5).Infof("Loading configuration done.")
	// create task work
	tw, err := handle.NewTaskWork(config.Taskconfig)
	if err != nil {
		return err
	}
	// start task work
	go tw.Run()

	server = Server{tw}
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.RequestURI {
	case "/task/v1/message":
		var reqBody struct {
			handle.Message
		}
		json_decoder := json.NewDecoder(r.Body)
		err := json_decoder.Decode(&reqBody)
		if err != nil {
			glog.Errorf("Failed to decode request body: %s.", err)
			msg := fmt.Sprintf("Failed to decode request body: %s.", err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		// handle task
		taskChan := &handle.TaskChan{
			MessageTask: reqBody.Message,
			ErrChan:     make(chan error),
		}
		if err := s.twClient.AsyncTask(*taskChan); err != nil {
			glog.Errorf("Message: %+v handle failed,error: %s.", reqBody.Message, err)
			msg := fmt.Sprintf("Message: %+v handle failed,error: %s.", reqBody.Message, err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		http.Error(w, "Message successfully received!", http.StatusOK)

	default:
		glog.Errorf("bad request.Got bad uri: \"%s\".", r.RequestURI)
		msg := fmt.Sprintf("bad request.Got invalid uri: \"%s\".", r.RequestURI)
		http.Error(w, msg, http.StatusBadRequest)
	}
}

func main() {
	flag.Parse()
	// generate the server
	glog.V(5).Infof("Starting Generating Server...")
	err := newServer()
	if err != nil {
		glog.Errorf("Generating Server failure: %s", err)
		return
	}
	glog.V(5).Infof("Generating Server done")

	// Setup HTTP endpoint
	glog.V(5).Infof("Listening on %s:%d ...\n", *argBindAddress, *argPort)

	// http server
	glog.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *argBindAddress, *argPort), &server))
}
