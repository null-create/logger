# About

This is a general purpose logger module I've been using for various projects and
figured I'd just place it in it's own thing.

Outputs logs in a .csv file using the filename format `log-dd-mm-yyyy.csv`

Set the optional `LOG_DIR` environment variable to specifcy a directory for the log file to live, otherwise it will try to create a new log directory in the current working directory.

## Install

```
go get github.com/null-create/logger
```

## Example

```go
package main

import (
  "github.com/google/uuid"
  "github.com/null-create/logger"
)

func main() {
  log := logger.NewLogger("My Component", uuid.NewString())
  log.Info("Hello")
  log.Warn("Uh oh")
  log.Error("Oh no")
}
```
