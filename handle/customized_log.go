package handle

import (
	"github.com/golang/glog"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

// Construct struct log for subsequent collection and analysis(eg: kafka)
// Add any field you would like to collect(eg: taskid) ...
func CustomizeLog(loglevel int, message string) {
	switch loglevel {
	case 0:
		log.WithFields(log.Fields{
			"message": message,
		}).Error("")
	case 1:
		log.WithFields(log.Fields{
			"message": message,
		}).Warn("")
	case 2:
		log.WithFields(log.Fields{
			"message": message,
		}).Info("")
	default:
		glog.Errorf("invalid loglevel: %d for message: %s", loglevel, message)
	}
}
