package model

import (
	"fmt"
	"log"
	"runtime"
)

func handleErr(err error) {
	log.Println(err)
	pcs := make([]uintptr, 10)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		fmt.Printf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
}
