package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// CarData struct
type CarData struct {
	Model          string  `json:"model"`
	Speed          int     `json:"speed"`
	FuelEfficiency float64 `json:"fuel_efficiency"`
	EngineTemp     float64 `json:"engine_temp"`
	Timestamp      string  `json:"timestamp"`
}

// PredictionRequest struct
type PredictionRequest struct {
	Instances [][]float64 `json:"instances"`
}

// GetPrediction calls AWS SageMaker for ML predictions
func GetPrediction(speed int, engineTemp float64) float64 {
	sagemakerEndpoint := os.Getenv("SAGEMAKER_ENDPOINT_URL")
	data := PredictionRequest{Instances: [][]float64{{float64(speed), engineTemp}}}
	payload, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", sagemakerEndpoint, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error calling SageMaker:", err)
		return -1
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string][]float64
	json.Unmarshal(body, &result)

	if len(result["predictions"]) > 0 {
		return result["predictions"][0]
	}
	return -1
}

// GetCarAnalytics generates real-time Porsche vehicle data with ML insights
func GetCarAnalytics(w http.ResponseWriter, r *http.Request) {
	models := []string{"911 RSR", "Cayenne Turbo GT", "Taycan Turbo S", "Macan GTS"}
	rand.Seed(time.Now().UnixNano())

	speed := rand.Intn(300)
	engineTemp := rand.Float64()*30 + 70
	predictedFuelEfficiency := GetPrediction(speed, engineTemp)

	data := CarData{
		Model:          models[rand.Intn(len(models))],
		Speed:          speed,
		FuelEfficiency: predictedFuelEfficiency,
		EngineTemp:     engineTemp,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/analytics", GetCarAnalytics)

	log.Println("Porsche Analytics with ML is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
