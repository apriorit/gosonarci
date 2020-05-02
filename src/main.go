package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type taskInfo struct {
	Organization       string `json:"organization"`
	ID                 string `json:"id"`
	TaskType           string `json:"taskType"`
	ComponentID        string `json:"componentId"`
	ComponentKey       string `json:"componentKey"`
	ComponentName      string `json:"componentName"`
	ComponentQualifier string `json:"componentQualifier"`
	Status             string `json:"status"`
	SubmittedAt        string `json:"submittedAt"`
	StartedAt          string `json:"startedAt"`
	ExecutedAt         string `json:"executedAt"`
	ExecutionTimeMs    int64  `json:"executionTimeMs"`
	Logs               bool   `json:"logs"`
	ErrorMessage       string `json:"errorMessage"`
	HasErrorStacktrace string `json:"hasErrorStacktrace"`
	HasScannerContext  bool   `json:"hasScannerContext"`
}

type taskStatusResponse struct {
	Tasks []taskInfo `json:"tasks"`
}

type projectStatusInfo struct {
	Status string `json:"status"`
}

type errorInfo struct {
	Message string `json:"msg"`
}

type errorResponse struct {
	Errors []errorInfo `json:"errors"`
}

type projectStatusResponse struct {
	ProjectStatus projectStatusInfo `json:"projectStatus"`
}

func sonarAPIRequest(
	sonarHostURL string, apiMethod string, httpMethod string, username string, password string) (
	[]byte, error) {

	var responseData []byte
	client := &http.Client{}
	var url string
	if strings.HasSuffix(sonarHostURL, "/") {
		url = sonarHostURL + apiMethod
	} else {
		url = sonarHostURL + "/" + apiMethod
	}
	componentRequest, err := http.NewRequest(
		httpMethod,
		url,
		nil)
	if err != nil {
		return responseData, err
	}

	componentRequest.SetBasicAuth(username, password)
	resp, err := client.Do(componentRequest)
	if err != nil {
		return responseData, err
	}
	defer resp.Body.Close()

	responseData, err = ioutil.ReadAll(resp.Body)
	return responseData, err
}

func apiQualityGatesProjectStatus(sonarHostURL string, projectKey string, token string) (projectStatusResponse, error) {
	var response projectStatusResponse
	var errorsResponse errorResponse

	body, err := sonarAPIRequest(
		sonarHostURL,
		"api/qualitygates/project_status?projectKey="+projectKey,
		"GET",
		token,
		"")

	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &errorsResponse)
	if err != nil || len(errorsResponse.Errors) > 0 {
		fmt.Println(string(body))
		return response, errors.New("failed to project status through API and parse response")
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func apiCeActivityStatus(
	sonarHostURL string, projectKey string, token string) (taskStatusResponse, error) {

	var response taskStatusResponse

	body, err := sonarAPIRequest(
		sonarHostURL,
		"api/ce/activity?status=PENDING,IN_PROGRESS&component="+projectKey,
		"GET",
		token,
		"")

	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err

	}
	return response, nil
}

func waitForPendingTasks(sonarHostURL string, projectKey string, token string, timeout time.Duration, refreshPeriod time.Duration) bool {
	var workTime time.Duration
	tasksCount := 1
	fmt.Println("\nWaiting for pending tasks to finish...")
	for workTime < timeout {
		response, err := apiCeActivityStatus(sonarHostURL, projectKey, token)
		if err != nil {
			fmt.Println("Failed to perform SonarQube API request for activities!")
			fmt.Println("Error: ", err)
			return false
		}
		tasksCount = len(response.Tasks)
		fmt.Printf("\r%d pending tasks remaining for %s component...", tasksCount, projectKey)
		if tasksCount == 0 {
			return true
		}
		time.Sleep(refreshPeriod)
		workTime += refreshPeriod
	}
	fmt.Println("\nTimeout reached!")
	return false
}

func isQualityGatePassed(sonarHostURL string, projectKey string, token string) bool {
	projectStatus, err := apiQualityGatesProjectStatus(sonarHostURL, projectKey, token)
	if err != nil {
		fmt.Printf("Failed to get project status for projectKey %s\n", projectKey)
		return false
	}

	fmt.Println("\n==============================================")
	fmt.Printf("Project Status: %s\n", projectStatus.ProjectStatus.Status)
	fmt.Println("==============================================")
	return projectStatus.ProjectStatus.Status == "OK"
}

func main() {
	fmt.Println("Running SonarQube Quality Gate checker!")
	serverPtr := flag.String("server", "http://localhost:9000/", "Sonar server address to use for API calls")
	projectKeyPtr := flag.String("project", "", "Sonar project (value from sonar.projectKey for your project) to check state of Quality Gate status")
	tokenPtr := flag.String("token", "", "User token for SonarQube to execute API requests. User has to have browse permission for the provided project")
	timeoutPtr := flag.Int64("timeout", 300, "Timeout in seconds to wait for pending tasks to finish execution")
	refreshPeriodPtr := flag.Int("refresh_period", 1, "Status refresh period")
	flag.Parse()

	if len(*projectKeyPtr) == 0 || len(*tokenPtr) == 0 {
		fmt.Println("Project key and token arguments are required!")
		os.Exit(-1)
	} else {
		fmt.Println("Checking if any tasks are running for the provided project...")
		timeout := time.Duration(*timeoutPtr) * time.Second
		period := time.Duration(*refreshPeriodPtr) * time.Second
		if waitForPendingTasks(*serverPtr, *projectKeyPtr, *tokenPtr, timeout, period) {
			fmt.Printf("\nAll tasks on project %s are finished!\n\n", *projectKeyPtr)
			fmt.Println("Checking Quality Gate status of the project...")
			if isQualityGatePassed(*serverPtr, *projectKeyPtr, *tokenPtr) {
				os.Exit(0)
			}
			os.Exit(1)
		} else {
			fmt.Printf("\nFailed to wait for project %s run outof tasks!\n\n", *projectKeyPtr)
			os.Exit(1)
		}
	}
}
