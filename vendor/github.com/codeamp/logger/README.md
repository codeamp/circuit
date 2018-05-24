# Logger [![CircleCI](https://circleci.com/gh/codeamp/logger.svg?style=svg)](https://circleci.com/gh/codeamp/logger) [![godoc](https://img.shields.io/badge/go-documentation-blue.svg)](https://godoc.org/github.com/codeamp/logger) [![Coverage Status](https://coveralls.io/repos/github/codeamp/logger/badge.svg?branch=master)](https://coveralls.io/github/codeamp/logger?branch=master) [![Go Report Card](https://goreportcard.com/badge/codeamp/logger)](https://goreportcard.com/report/codeamp/logger) [![codebeat badge](https://codebeat.co/badges/1c3bae3f-fb77-437e-93cb-7e07869898b5)](https://codebeat.co/projects/github-com-codeamp-logger-master)
Simple logrus wrapper

### Example Usage
```
package main

import (
	log "github.com/codeamp/logger"
)

func main() {
  log.InfoWithFields("hello", log.Fields{
    "world": "earth",
  })

  log.Println("Hello World")
  
  log.Debug("Hello World")
}
```
