package main

import (
	"cloud-storage/core/config"
	"cloud-storage/core/utility/logger"
	"flag"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// cmd-line flag
	isDebug     = flag.Bool("debug", false, "specifiy log level (debug/production), default production")
	rate        = flag.Float64("rate", 3.0, "usert rate input, expected average of requests per minute: (0.0 - 10.0]")
	sensitivity = flag.Int("sensitivity", 5, "user sensitivity input for the weight on the latency part of the system, higher implies more weight on latency: [1-5]")

	log = logrus.WithField("component", "configurator")

	minSensitivity = 1
	maxSensitivity = 5
)

func deriveThresholds(rate float64, sensitivity int) {
	log.Info("Compute system parameters based on user input:")
	log.Infof("\t\t%-26s%d", "sensitivity:", sensitivity)
	log.Infof("\t\t%-26s%.2f", "rate:", rate)

	if !((0.0 < rate) && (rate < 10.0)) {
		log.Fatal("choose valid rate value: (0.0 - 10.0]")
	}
	if !((minSensitivity <= sensitivity) && (sensitivity <= maxSensitivity)) {
		log.Fatalf("choose valid sensitivity value:  [%d-%d]", minSensitivity, maxSensitivity)
	}

	thresholdHeuristics(rate, sensitivity)
}

// heuristic algorithm to determine the thresholds based on rate parameter and sensitivity value
func thresholdHeuristics(rate float64, s int) {

	avg_inbetween := 60.0 / rate //seconds between two requests on average
	log.Infof("time between requests (avg): %.3f seconds", avg_inbetween)

	tick := int(avg_inbetween - (float64(maxSensitivity-s) * (avg_inbetween / float64(maxSensitivity))))
	lambdaWindowElements := 2
	lambdaThreshold := time.Second * time.Duration(tick*lambdaWindowElements)
	redisThreshold := 4
	redisUtilization := time.Second * time.Duration(5*avg_inbetween+(avg_inbetween*float64(s)))

	// old values:
	// tick = 10 seconds
	// lambdaWindowElements = 2
	// lambdaThreshold = 25 seconds
	// redisThreshold = 8
	// redisUtilization = 2 minutes

	cfg := config.Conf{
		Tick:                 tick,
		LambdaWindowElements: lambdaWindowElements,
		LambdaThreshold:      lambdaThreshold,
		RedisThreshold:       redisThreshold,
		RedisUtilization:     redisUtilization,
	}
	cfg.PrintConfig()

	config.WriteConf(&cfg)

}

func main() {
	flag.Parse()

	logger.SetupLogger(*isDebug)

	// set thresholds for ObjectManager
	deriveThresholds(*rate, *sensitivity)

}
