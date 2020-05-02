# golang Sonar QG checker scanner
## General

This tool is designed to check SQ Quality Gate status of the project to be used
in CI pipelines.

## Usage 
```
Running SonarQube Quality Gate checker!
Usage of ./sqqg:
  -project string
        Sonar project (value from sonar.projectKey for your project) to check state of Quality Gate status
  -refresh_period int
        Status refresh period (default 1)
  -server string
        Sonar server address to use for API calls (default "http://localhost:9000/")
  -timeout int
        Timeout in seconds to wait for pending tasks to finish execution (default 300)
  -token string
        User token for SonarQube to execute API requests. User has to have browse permission for the provided project
```
