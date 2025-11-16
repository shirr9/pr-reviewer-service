package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL        = "http://localhost:8080"
	numTeams       = 5
	usersPerTeam   = 10
	numPRs         = 50
	concurrentReqs = 10
	testDuration   = 30 * time.Second
)

type Stats struct {
	mu              sync.Mutex
	totalRequests   int
	successRequests int
	failedRequests  int
	totalLatency    time.Duration
	minLatency      time.Duration
	maxLatency      time.Duration
	latencies       []time.Duration
}

func (s *Stats) recordRequest(latency time.Duration, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalRequests++
	s.totalLatency += latency
	s.latencies = append(s.latencies, latency)

	if success {
		s.successRequests++
	} else {
		s.failedRequests++
	}

	if s.minLatency == 0 || latency < s.minLatency {
		s.minLatency = latency
	}
	if latency > s.maxLatency {
		s.maxLatency = latency
	}
}

func (s *Stats) printReport() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.totalRequests == 0 {
		fmt.Println("No requests recorded")
		return
	}

	avgLatency := s.totalLatency / time.Duration(s.totalRequests)
	successRate := float64(s.successRequests) / float64(s.totalRequests) * 100

	fmt.Println("\n=== Load Test Results ===")
	fmt.Printf("Total Requests:    %d\n", s.totalRequests)
	fmt.Printf("Success:           %d (%.2f%%)\n", s.successRequests, successRate)
	fmt.Printf("Failed:            %d\n", s.failedRequests)
	fmt.Printf("Min Latency:       %v\n", s.minLatency)
	fmt.Printf("Max Latency:       %v\n", s.maxLatency)
	fmt.Printf("Avg Latency:       %v\n", avgLatency)
	fmt.Printf("P95 Latency:       %v\n", s.percentile(0.95))
	fmt.Printf("P99 Latency:       %v\n", s.percentile(0.99))
	fmt.Println("========================")
}

func (s *Stats) percentile(p float64) time.Duration {
	if len(s.latencies) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(s.latencies))
	copy(sorted, s.latencies)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	idx := int(float64(len(sorted)) * p)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}

	return sorted[idx]
}

func main() {
	fmt.Println("Starting load test...")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Concurrent requests: %d\n", concurrentReqs)
	fmt.Printf("Test duration: %v\n", testDuration)

	stats := &Stats{
		minLatency: time.Hour,
	}

	setupTestData()
	time.Sleep(2 * time.Second)

	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	for i := 0; i < concurrentReqs; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runWorker(workerID, stats, stopCh)
		}(i)
	}

	time.Sleep(testDuration)
	close(stopCh)
	wg.Wait()

	stats.printReport()
}

func setupTestData() {
	fmt.Println("Setting up test data...")

	for i := 1; i <= numTeams; i++ {
		teamName := fmt.Sprintf("team-%d", i)
		members := []map[string]interface{}{}

		for j := 1; j <= usersPerTeam; j++ {
			userID := fmt.Sprintf("user-%d-%d", i, j)
			members = append(members, map[string]interface{}{
				"user_id":   userID,
				"username":  fmt.Sprintf("User %d-%d", i, j),
				"is_active": true,
			})
		}

		payload := map[string]interface{}{
			"team_name": teamName,
			"members":   members,
		}

		makeRequest("POST", "/team/add", payload)
	}

	fmt.Printf("Created %d teams with %d users each\n", numTeams, usersPerTeam)
}

func runWorker(workerID int, stats *Stats, stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
			operation := rand.Intn(5)

			switch operation {
			case 0:
				createPR(stats)
			case 1:
				getUserReviews(stats)
			case 2:
				getStatistics(stats)
			case 3:
				getTeam(stats)
			case 4:
				mergePR(stats)
			}

			time.Sleep(time.Millisecond * time.Duration(10+rand.Intn(90)))
		}
	}
}

func createPR(stats *Stats) {
	prID := fmt.Sprintf("pr-%d-%d", time.Now().Unix(), rand.Intn(10000))
	teamID := rand.Intn(numTeams) + 1
	userID := fmt.Sprintf("user-%d-%d", teamID, rand.Intn(usersPerTeam)+1)

	payload := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": fmt.Sprintf("Feature %s", prID),
		"author_id":         userID,
	}

	start := time.Now()
	resp := makeRequest("POST", "/pullRequest/create", payload)
	latency := time.Since(start)

	success := resp != nil && (resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict)
	stats.recordRequest(latency, success)

	if resp != nil {
		resp.Body.Close()
	}
}

func getUserReviews(stats *Stats) {
	teamID := rand.Intn(numTeams) + 1
	userID := fmt.Sprintf("user-%d-%d", teamID, rand.Intn(usersPerTeam)+1)

	start := time.Now()
	resp := makeRequest("GET", fmt.Sprintf("/users/getReview?user_id=%s", userID), nil)
	latency := time.Since(start)

	success := resp != nil && resp.StatusCode == http.StatusOK
	stats.recordRequest(latency, success)

	if resp != nil {
		resp.Body.Close()
	}
}

func getStatistics(stats *Stats) {
	start := time.Now()
	resp := makeRequest("GET", "/statistics", nil)
	latency := time.Since(start)

	success := resp != nil && resp.StatusCode == http.StatusOK
	stats.recordRequest(latency, success)

	if resp != nil {
		resp.Body.Close()
	}
}

func getTeam(stats *Stats) {
	teamID := rand.Intn(numTeams) + 1
	teamName := fmt.Sprintf("team-%d", teamID)

	start := time.Now()
	resp := makeRequest("GET", fmt.Sprintf("/team/get?team_name=%s", teamName), nil)
	latency := time.Since(start)

	success := resp != nil && resp.StatusCode == http.StatusOK
	stats.recordRequest(latency, success)

	if resp != nil {
		resp.Body.Close()
	}
}

func mergePR(stats *Stats) {
	prID := fmt.Sprintf("pr-merge-%d-%d", time.Now().Unix(), rand.Intn(1000))

	start := time.Now()
	resp := makeRequest("POST", "/pullRequest/merge", map[string]interface{}{
		"pull_request_id": prID,
	})
	latency := time.Since(start)

	success := resp != nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound)
	stats.recordRequest(latency, success)

	if resp != nil {
		resp.Body.Close()
	}
}

func makeRequest(method, path string, payload interface{}) *http.Response {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}

	return resp
}
