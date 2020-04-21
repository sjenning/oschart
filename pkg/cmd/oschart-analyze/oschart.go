package oschart_analyze

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	configv1 "github.com/openshift/api/config/v1"

	"github.com/spf13/cobra"

	"github.com/sjenning/oschart/pkg/cmd"
	"github.com/sjenning/oschart/pkg/event"
	"github.com/sjenning/oschart/pkg/ui"
)

var defaultHttpPort = 3001

type AnalyzeOptions struct {
	Files    []string
	HTTPPort uint16
}

func NewAnalyzeOptions() *AnalyzeOptions {
	return &AnalyzeOptions{}
}

func NewCommand(name string) *cobra.Command {
	o := NewAnalyzeOptions()

	c := &cobra.Command{
		Use:   name,
		Short: "Monitor ClusterOperator phase transitions over time in a OpenShift cluster.",
		Run: func(c *cobra.Command, args []string) {
			cmd.CheckError(o.Run())
		},
	}

	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	c.Flags().StringArrayVarP(&o.Files, "file", "f", o.Files, "files to load")
	c.Flags().Uint16Var(&o.HTTPPort, "http-port", uint16(defaultHttpPort), fmt.Sprintf("Port to serve charts on.", defaultHttpPort))

	return c
}

func (o *AnalyzeOptions) Run() error {
	eventStore := event.NewStore()

	// wire a way to go walk through a git repo from beginning to end.
	// For each commit modifying a clusteroperator, add that file to the event store
	// For each commit modifying a known {kube-apiserver,kube-controller-manager,kube-scheduler,etcd,authentication,openshift-apiserver}.operator.openshift.io, do similar

	// get from the commit
	commitTime := time.Now()
	// parse the yaml
	co := &configv1.ClusterOperator{}
	// co is added or updated
	data, _ := json.MarshalIndent(co, "", "  ")
	for _, condition := range co.Status.Conditions {
		if condition.Status == configv1.ConditionTrue {
			eventStore.AddAtTime(co.GetName(), string(condition.Type), string(condition.Type), string(data), commitTime)
		} else {
			eventStore.AddAtTime(co.GetName(), string(condition.Type), string(condition.Status), string(data), commitTime)
		}
	}

	// decompose this into the actual index.html.  We're clever people, this will work.
	go ui.Run(eventStore, o.HTTPPort)

	return nil
}
