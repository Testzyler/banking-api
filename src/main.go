package main

import (
	"runtime"

	"github.com/Testzyler/banking-api/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
