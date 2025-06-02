package logger

import (
    "io"
    "log"
    "os"
)

var Logger *log.Logger

func InitLogger() {
    file, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatalf("Could not open log file: %v", err)
    }
    Logger = log.New(io.MultiWriter(os.Stdout, file), "", log.LstdFlags)
}
