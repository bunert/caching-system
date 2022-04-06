package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// set logrus settings for orchestrator/gateway servers
func SetupLogger(isDebug bool) {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&Formatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FieldsOrder:     []string{"component"},
		HideKeys:        true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Define the level of logging
	if isDebug {
		logrus.SetLevel(logrus.DebugLevel)
		return
	}
	logrus.SetLevel(logrus.InfoLevel)

}

// set logrus settings for lambda-runtime
func SetupLambdaLogger(id string) {
	logrus.SetFormatter(&Formatter{
		defaultField:    id,
		NoColors:        true,
		TimestampFormat: "-",
		FieldsOrder:     []string{"id", "component"},
		HideKeys:        true,
	})
	logrus.SetLevel(logrus.InfoLevel)
}
