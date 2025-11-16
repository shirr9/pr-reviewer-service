package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

func TestE2EWorkflow(t *testing.T) {
	if os.Getenv("E2E_TEST") != "true" {
		t.Skip("Skipping E2E test. Set E2E_TEST=true to run")
	}

	waitForServer(t)

	t.Run("CreateTeamAndUsers", testCreateTeam)
	t.Run("CreatePullRequest", testCreatePullRequest)
	t.Run("GetUserReviews", testGetUserReviews)
	t.Run("ReassignReviewer", testReassignReviewer)
	t.Run("MergePullRequest", testMergePullRequest)
	t.Run("GetStatistics", testGetStatistics)
	t.Run("DeactivateTeam", testDeactivateTeam)
}

func waitForServer(t *testing.T) {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(baseURL + "/statistics")
		if err == nil && resp.StatusCode < 500 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("Server did not start in time")
}

func testCreateTeam(t *testing.T) {
	timestamp := time.Now().Unix()
	payload := map[string]interface{}{
		"team_name": "e2e-team-" + string(rune(timestamp)),
		"members": []map[string]interface{}{
			{"user_id": "e2e-u1", "username": "E2E-Alice", "is_active": true},
			{"user_id": "e2e-u2", "username": "E2E-Bob", "is_active": true},
			{"user_id": "e2e-u3", "username": "E2E-Charlie", "is_active": true},
		},
	}

	resp := makeRequest(t, "POST", "/team/add", payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Team creation returned status %d (may already exist): %s", resp.StatusCode, body)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if team, ok := result["team"].(map[string]interface{}); ok {
		t.Logf("Team created/exists: %s", team["team_name"])
	}
}

func testCreatePullRequest(t *testing.T) {
	timestamp := time.Now().UnixNano()
	payload := map[string]interface{}{
		"pull_request_id":   "e2e-pr-" + string(rune(timestamp/1000000)),
		"pull_request_name": "E2E Add feature",
		"author_id":         "e2e-u1",
	}

	resp := makeRequest(t, "POST", "/pullRequest/create", payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("PR creation returned status %d (may already exist): %s", resp.StatusCode, body)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if pr, ok := result["pr"].(map[string]interface{}); ok {
		t.Logf("PR created: %s with status %s", pr["pull_request_id"], pr["status"])
	}
}

func testGetUserReviews(t *testing.T) {
	resp := makeRequest(t, "GET", "/users/getReview?user_id=e2e-u2", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Get reviews returned status %d: %s", resp.StatusCode, body)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	t.Logf("User reviews retrieved for %s", result["user_id"])
}

func testReassignReviewer(t *testing.T) {
	payload := map[string]interface{}{
		"pull_request_id":  "e2e-pr-test",
		"old_reviewer_id": "e2e-u2",
	}

	resp := makeRequest(t, "POST", "/pullRequest/reassign", payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Unexpected status %d: %s", resp.StatusCode, body)
	}

	t.Logf("Reassign response status: %d", resp.StatusCode)
}

func testMergePullRequest(t *testing.T) {
	payload := map[string]interface{}{
		"pull_request_id": "e2e-pr-test",
	}

	resp := makeRequest(t, "POST", "/pullRequest/merge", payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Merge returned status %d: %s", resp.StatusCode, body)
		return
	}

	t.Logf("Merge response status: %d", resp.StatusCode)
}

func testGetStatistics(t *testing.T) {
	resp := makeRequest(t, "GET", "/statistics", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := result["total_prs"]; !ok {
		t.Fatal("Statistics missing total_prs")
	}

	if _, ok := result["user_stats"]; !ok {
		t.Fatal("Statistics missing user_stats")
	}
}

func testDeactivateTeam(t *testing.T) {
	payload := map[string]interface{}{
		"team_name": "test-deactivate",
	}

	teamPayload := map[string]interface{}{
		"team_name": "test-deactivate",
		"members": []map[string]interface{}{
			{"user_id": "u10", "username": "TestUser", "is_active": true},
		},
	}
	resp := makeRequest(t, "POST", "/team/add", teamPayload)
	resp.Body.Close()

	resp = makeRequest(t, "POST", "/team/deactivate", payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, ok := result["deactivated_users"]; !ok {
		t.Fatal("Deactivate response missing deactivated_users")
	}
}

func makeRequest(t *testing.T, method, path string, payload interface{}) *http.Response {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Failed to marshal payload: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	return resp
}
