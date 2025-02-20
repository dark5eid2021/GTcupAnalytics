package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/session"
	_ "github.com/lib/pq"
)

// CarData represennts  Porsche car analytics
type CarData struct {
	Model          string  `json:"model"`
	Speed          int     `json:"speed"`           // in km/h
	FuelEfficiency float64 `json:"fuel_efficiency"` // km/l
	EngineTemp     float64 `json:"engine_temp"`     // in degrees C
	Timestamp      string  `json:"timestamp"`
}

// Wrap these in a struct
// Database connection
var db *sql.DB

// kinesis client
var (
	kinesisClient *kinesis.Kinesis
	streamName    = "porsch-analytics-stream"
)

func init() {
	// Move init contents to main
	// Load AWS Kinesis session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	kinesisClient = kinesis.New(sess)

	// load PostgreSQL connection
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// create the table if it does not exist
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS analytics (
		id SERIAL PRIMARY KEY,
		model TEXT,
		speed INT,
		fuel_efficiency FLOAT,
		engine_temp FLOAT,
		timestamp TIMESTAMP
		)`)
	if err != nil {
		log.Fatalf("Failed to create table: %v, err")
	}
}

// GenerateCarData simulates real-tiime vehicle analytics
func GenerateCarData() CarData {
	models := []string{"911 RSR", "Cayenne Turbo GT", "Taycan Turbo S", "Macan GTS"}
	rand.Seed(time.Now().UnixNano())

	return CarData{
		Model:          models[rand.Intn(len(models))],
		Speed:          rand.Intn(300),            // Random speed up to 300 km/h
		FuelEfficiency: rand.Float64()*10 + 5,     // between 5-15 km/l
		EngineTemp:     rand.ExpFloat64()*30 + 70, // between 70-100 degrees C
		Timestamp:      time.Now().Format(time.RFC3339),
	}
}

// Make this a method on the new struct
// Send data to Kinesis Stream
func sendToKinesis(data CarData) {
	jsonData, _ := json.Marshal(data)
	_, err := kinesisClient.PutRecord(&kinesis.PutRecordInput{
		Data:         jsonData,
		StreamName:   aws.String(streamName),
		PartitionKey: aws.String(data.Model),
	})
	if err != nil {
		log.Printf("Error sending to Kinesis: %v", err)
	}
}

// Make this a method on the new struct
// store data in PSQL
func storeInDatabase(data CarData) {
	_, err := db.Exec(`
	INSERT INTO analytics (model, speed, fuel_efficiency, engine_temp, timestamp)
	VALUES ($1, $2, $3, $4, $5)`,
		data.Model, data.Speed, data.FuelEfficiency, data.EngineTemp, data.Timestamp)
	if err != nil {
		log.Printf("Error inserting data: %v", err)
	}
}

// Make this a method on the new struct
// handle real-time analytics API
func GetCarAnalytics(w http.ResponseWriter, r *http.Request) {
	data := GenerateCarData()
	sendToKinesis(data)
	storeInDatabase(data)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Make this a method on the new struct
// retrieve analytics from PSQL
func GetStoredAnalytics(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT model, speed, fuel_efficiency, engine_temp, timestamp FROM analytics ORDER BY timestamp DESC LIMIT 10`)
	if err != nil {
		http.Error(w, "Failed to fetch analytics", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []CarData
	for rows.Next() {
		var data CarData
		err := rows.Scan(&data.Model, &data.Speed, &data.FuelEfficiency, &data.EngineTemp, &data.Timestamp)
		if err != nil {
			continue
		}
		results = append(results, data)
	}

	// Want to check row.Err here

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	http.HandleFunc("/analytics", GetCarAnalytics)
	http.HandleFunc("/analytics/history", GetStoredAnalytics)

	log.Println("Porsche Analytics Service Runninng on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
