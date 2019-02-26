package main

import (
	"os"
	"path/filepath"

	"github.com/golang/glog"

	"github.com/sjenning/oschart/pkg/cmd"
	"github.com/sjenning/oschart/pkg/cmd/oschart"
)

func main() {
	defer glog.Flush()

	baseName := filepath.Base(os.Args[0])

	err := oschart.NewCommand(baseName).Execute()
	cmd.CheckError(err)
}
