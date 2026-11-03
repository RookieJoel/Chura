package main

import "log"

func main() {
	log.Fatal("cmd/api is deprecated: run cmd/project, cmd/sprint, cmd/analytics, and cmd/workitem (or use deploy/docker-compose.yml with Kong on :8083)")
}
