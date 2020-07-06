package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	temperature = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sensor_temperature",
		Help: "Temperature",
	})
	humidity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sensor_humidity",
		Help: "Humidity",
	})
)

type Config struct {
	Secret string `yaml:"secret"`
}

type Secrets struct {
	path string
}

func init() {
	// log as JSON
	log.SetFormatter(&log.JSONFormatter{})
	// Output everything including stderr to stdout
	log.SetOutput(os.Stdout)
	// set level
	log.SetLevel(log.InfoLevel)

	// register Prometheus metrics
	prometheus.MustRegister(temperature)
	prometheus.MustRegister(humidity)
}

func (s *Secrets) secretsHandler(w http.ResponseWriter, req *http.Request) {

	yamlFile, err := ioutil.ReadFile((*s).path)
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	log.Info(config)
	w.Write([]byte("The secret is: " + config.Secret))
	log.Info(req.Method + req.RequestURI + " " + req.Proto)
}

func headersHandler(w http.ResponseWriter, req *http.Request) {
	jsonHeaders, err := json.Marshal((*req).Header)
	if err != nil {
		log.Error(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonHeaders)
	log.Info(req.Method + req.RequestURI + " " + req.Proto)
}

func recordMetrics() {
	go func() {
		for {
			temperature.Inc()
			humidity.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

func main() {

	port := ":8080"

	s := Secrets{
		path: "resources/config.yaml",
	}

	recordMetrics()
	http.HandleFunc("/secrets", s.secretsHandler)
	http.Handle("/metrics", promhttp.Handler())

	log.Info("Listening on port ", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
