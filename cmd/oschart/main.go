package main

import (
	"os"
	"path/filepath"

	"k8s.io/klog/v2"

	"github.com/sjenning/oschart/pkg/cmd"
	"github.com/sjenning/oschart/pkg/cmd/oschart"
)

func main() {
	defer klog.Flush()

	baseName := filepath.Base(os.Args[0])

	err := oschart.NewCommand(baseName).Execute()
	cmd.CheckError(err)
}
