package logger

import (
	"fmt"
)

func Info(args ...interface{}) {
	fmt.Print("info ")
	fmt.Println(args...)
}

func Infof(template string, args ...interface{}) {
	fmt.Printf("info "+template+"\n", args...)
}

func Error(args ...interface{}) {
	fmt.Print("error ")
	fmt.Println(args...)
}

func Errorf(template string, args ...interface{}) {
	fmt.Printf("error "+template+"\n", args...)
}
