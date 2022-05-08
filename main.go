package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type RepoResults struct {
	User      string `json:"user"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    int    `json:"status"`
	StarCount int    `json:"star_count"`
	PullCount int    `json:"pull_count"`
}

type RepoList struct {
	Count    int `json:"count"`
	Next     int `json:"next"`
	Previous int `json:"previous"`
	Results  []RepoResults
}

var repolist RepoList
var repores RepoResults

func getJson(url string, target interface{}) error {
	var myClient = &http.Client{Timeout: 20 * time.Second}
	r, err := myClient.Get(url)
	if err != nil {
		fmt.Println("ERROR: in getJson function")
		return err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("ERROR: in getJson function")
		return err
	}

	json.Unmarshal(body, &repolist)
	fmt.Println(repolist)

	return json.NewDecoder(r.Body).Decode(target)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func recordMetrics() {
	go func() {
		for {
			getJson("https://hub.docker.com/v2/repositories/"+getEnv("DOCKERHUB_ORGANIZATION", "github")+"/?page_size=25&page=1", nil)
			for i := 0; i < len(repolist.Results); i++ {
				repores = repolist.Results[i]
				dockerImagePulls.WithLabelValues(repores.Name, repores.User).Set(float64(repores.PullCount))
			}

			time.Sleep(30 * time.Second)
		}
	}()
}

var metricsLabel = []string{"image", "organization"}
var (
	dockerImagePulls = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_image_pulls",
		Help: "The total number of Docker image pulls",
	},
		metricsLabel,
	)
)

func main() {
	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	recordMetrics()
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2113", nil)
}
