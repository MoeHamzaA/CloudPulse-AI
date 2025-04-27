package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	logsTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type PredictionResponse struct {
	Prediction   string    `json:"prediction"`
	ResponseTime float64   `json:"response_time"`
	Timestamp    time.Time `json:"timestamp"`
	Confidence   float64   `json:"confidence,omitempty"`
	ModelVersion string    `json:"model_version,omitempty"`
	Error        string    `json:"error,omitempty"`
}

type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	Uptime      float64 `json:"uptime"`
}

type DriftMetrics struct {
	PredictionChanges int     `json:"prediction_changes"`
	ConfidenceDrop    float64 `json:"confidence_drop"`
	ResponseTimeSpike float64 `json:"response_time_spike"`
}

var (
	logGroup        = "CloudPulseLogs"
	logStream       = "monitoring-agent-stream"
	logClient       *cloudwatchlogs.Client
	sequenceToken   *string
	lastPredictions []string
	maxPredictions  = 10

	logs []PredictionResponse
	logsMutex sync.Mutex

	// System metrics
	systemMetrics SystemMetrics
	metricsMutex  sync.Mutex

	// Drift detection
	driftMetrics DriftMetrics
	driftMutex   sync.Mutex

	// Configuration
	appConfig = struct {
		ModelEndpoint       string
		PollingInterval     time.Duration
		DriftThreshold      float64
		ConfidenceThreshold float64
		MaxResponseTime     float64
	}{
		ModelEndpoint:       "http://127.0.0.1:5000/predict",
		PollingInterval:     5 * time.Second,
		DriftThreshold:      0.5,
		ConfidenceThreshold: 0.8,
		MaxResponseTime:     1.0,
	}

	startTime = time.Now()
)

func initAWS() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("Unable to load AWS SDK config: " + err.Error())
	}

	logClient = cloudwatchlogs.NewFromConfig(cfg)

	_, err = logClient.CreateLogStream(context.TODO(), &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroup),
		LogStreamName: aws.String(logStream),
	})
	if err != nil {
		fmt.Println("Log stream exists or another issue:", err)
	}
}

func sendToCloudWatch(message string, level string) {
	input := &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(logGroup),
		LogStreamName: aws.String(logStream),
		LogEvents: []logsTypes.InputLogEvent{
			{
				Message:   aws.String(fmt.Sprintf("[%s] %s", level, message)),
				Timestamp: aws.Int64(time.Now().Unix() * 1000),
			},
		},
		SequenceToken: sequenceToken,
	}

	result, err := logClient.PutLogEvents(context.TODO(), input)
	if err != nil {
		fmt.Println("❌ Failed to send log to CloudWatch:", err)
		return
	}

	sequenceToken = result.NextSequenceToken
	fmt.Printf("✅ Sent to AWS CloudWatch: [%s] %s\n", level, message)
}

func updateSystemMetrics() {
    metricsMutex.Lock()
    defer metricsMutex.Unlock()

    // Get CPU usage (percent over 1 second)
    cpuPercents, err := cpu.Percent(0, false)
    if err == nil && len(cpuPercents) > 0 {
        systemMetrics.CPUUsage = cpuPercents[0]
    } else {
        systemMetrics.CPUUsage = 0
    }

    // Get Memory usage (percent)
    vmStat, err := mem.VirtualMemory()
    if err == nil {
        systemMetrics.MemoryUsage = vmStat.UsedPercent
    } else {
        systemMetrics.MemoryUsage = 0
    }

    // Uptime stays the same
    systemMetrics.Uptime = time.Since(startTime).Seconds()
}


func detectDrift(currentPrediction string, confidence float64, responseTime float64) {
	driftMutex.Lock()
	defer driftMutex.Unlock()

	lastPredictions = append(lastPredictions, currentPrediction)
	if len(lastPredictions) > maxPredictions {
		lastPredictions = lastPredictions[1:]
	}

	// Calculate prediction changes
	flips := 0
	for i := 1; i < len(lastPredictions); i++ {
		if lastPredictions[i] != lastPredictions[i-1] {
			flips++
		}
	}
	driftMetrics.PredictionChanges = flips

	// Check for confidence drop
	if confidence < appConfig.ConfidenceThreshold {
		driftMetrics.ConfidenceDrop = appConfig.ConfidenceThreshold - confidence
	}

	// Check for response time spike
	if responseTime > appConfig.MaxResponseTime {
		driftMetrics.ResponseTimeSpike = responseTime - appConfig.MaxResponseTime
	}

	// Determine if drift is detected
	if flips >= int(float64(maxPredictions)*appConfig.DriftThreshold) ||
		driftMetrics.ConfidenceDrop > 0 ||
		driftMetrics.ResponseTimeSpike > 0 {
		
		driftWarning := fmt.Sprintf("⚠️ Drift Detected: Changes=%d, Confidence Drop=%.2f, Response Time Spike=%.2fs",
			flips, driftMetrics.ConfidenceDrop, driftMetrics.ResponseTimeSpike)
		
		fmt.Println(driftWarning)
		sendToCloudWatch(driftWarning, "WARNING")

		logsMutex.Lock()
		logs = append(logs, PredictionResponse{
			Prediction:   driftWarning,
			ResponseTime: 0,
			Timestamp:    time.Now(),
		})
		logsMutex.Unlock()
	}
}

func sendRequest() {
	payload := map[string]string{"text": "Hello model!"}
	jsonData, _ := json.Marshal(payload)

	start := time.Now()
	resp, err := http.Post(appConfig.ModelEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Request error: %v", err)
		fmt.Println(errorMsg)
		sendToCloudWatch(errorMsg, "ERROR")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	responseTime := time.Since(start).Seconds()

	var result PredictionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		errorMsg := fmt.Sprintf("❌ Parse error: %v", err)
		fmt.Println(errorMsg)
		sendToCloudWatch(errorMsg, "ERROR")
		return
	}

	result.Timestamp = time.Now()
	result.ResponseTime = responseTime

	msg := fmt.Sprintf("Prediction: %s | Response Time: %.4f sec | Confidence: %.2f",
		result.Prediction, result.ResponseTime, result.Confidence)
	fmt.Println(msg)

	sendToCloudWatch(msg, "INFO")

	logsMutex.Lock()
	logs = append(logs, result)
	logsMutex.Unlock()

	detectDrift(result.Prediction, result.Confidence, result.ResponseTime)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		http.Error(w, "Failed to load dashboard", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	logsMutex.Lock()
	defer logsMutex.Unlock()

	json.NewEncoder(w).Encode(logs)
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	json.NewEncoder(w).Encode(systemMetrics)
}

func handleDriftMetrics(w http.ResponseWriter, r *http.Request) {
	driftMutex.Lock()
	defer driftMutex.Unlock()

	json.NewEncoder(w).Encode(driftMetrics)
}

func main() {
	fmt.Println("🚀 Starting CloudPulse Monitoring Agent...")
	initAWS()

	// Run monitoring agent in background
	go func() {
		for {
			sendRequest()
			time.Sleep(appConfig.PollingInterval)
		}
	}()

	// Run system metrics collection in background
	go func() {
		for {
			updateSystemMetrics()
			time.Sleep(1 * time.Second)
		}
	}()

	// Web server routes
	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/logs", handleLogs)
	http.HandleFunc("/metrics", handleMetrics)
	http.HandleFunc("/drift-metrics", handleDriftMetrics)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("🌐 Serving dashboard on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
