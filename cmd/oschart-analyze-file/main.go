package main

import (
	"os"
	"path/filepath"

	oschart_analyze "github.com/sjenning/oschart/pkg/cmd/oschart-analyze"

	"github.com/golang/glog"

	"github.com/sjenning/oschart/pkg/cmd"
)

func main() {
	defer glog.Flush()

	baseName := filepath.Base(os.Args[0])

	err := oschart_analyze.NewCommand(baseName).Execute()
	cmd.CheckError(err)
}
