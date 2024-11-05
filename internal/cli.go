package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3/log"
)

// Creates the instance. Checks if the deno command exists: exits with 1 if it doesn't.
// Uses default system's output.
func ExecDeno(arg ...string) *exec.Cmd {
	_, err := exec.LookPath("deno")
	if err != nil {
		fmt.Println("Deno is not installed or not available in PATH.")
		os.Exit(1)
	}

	cmd := exec.Command("deno", arg...)

	return cmd
}

func WaitForBundleWatch() {

	optionWatch := "--watch"
	optionBuild := "--build"

	isBuild := slices.Contains(os.Args, optionBuild)
	isWatch := slices.Contains(os.Args, optionWatch)

	if !isBuild && !isWatch {
		log.Info("You can use --build or --watch option to bundle js, css and assets before running server.")
		return
	}

	if !isBuild && !isWatch {
		log.Fatal("Use --build or --watch, if you want to bundle while running the server. You have used both.")
		return
	}

	var deno *exec.Cmd
	if isWatch {
		log.Info("Creating file listeners for bundling js, css and assets...")
		deno = ExecDeno("task", "build", "--watch")
	} else {
		log.Info("Bundling js, css and assets...")
		deno = ExecDeno("task", "build")
	}

	reader, err := deno.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	readerErr, err := deno.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(reader)
	scannerErr := bufio.NewScanner(readerErr)
	err = deno.Start()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for scannerErr.Scan() {
			line := scannerErr.Text()

			fmt.Println(line)

			// see ./web/build.ts file
			if strings.Contains(line, "Error") {
				go func() {
					deno.Wait()
					os.Exit(1)
				}()
				continue
			}
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()

		fmt.Println(line)

		// see ./web/build.ts file
		if strings.Contains(line, "Done:") {
			log.Info("Now starting the server...")
			break
		}
	}

}
