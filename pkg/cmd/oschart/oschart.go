package oschart

import (
	"flag"
	"time"

	"github.com/spf13/cobra"

	"github.com/sjenning/oschart/pkg/client"
	"github.com/sjenning/oschart/pkg/cmd"
	"github.com/sjenning/oschart/pkg/controller"
	"github.com/sjenning/oschart/pkg/event"
	"github.com/sjenning/oschart/pkg/signals"
	"github.com/sjenning/oschart/pkg/ui"

	configinformers "github.com/openshift/client-go/config/informers/externalversions"
)

func NewCommand(name string) *cobra.Command {
	f := client.NewFactory(name)

	c := &cobra.Command{
		Use:   name,
		Short: "Monitor ClusterOperator phase transitions over time in a OpenShift cluster.",
		Run: func(c *cobra.Command, args []string) {
			cmd.CheckError(run(c, f))
		},
	}

	f.BindFlags(c.PersistentFlags())
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return c
}

func run(c *cobra.Command, f client.Factory) error {
	stopCh := signals.SetupSignalHandler()

	client, err := f.Client()
	if err != nil {
		return err
	}
	sharedInformerFactory := configinformers.NewSharedInformerFactory(client, time.Second*30)
	coInformer := sharedInformerFactory.Config().V1().ClusterOperators()
	eventStore := event.NewStore()
	controller := controller.New(client, coInformer, eventStore)
	if err != nil {
		return err
	}
	sharedInformerFactory.Start(stopCh)

	go ui.Run(eventStore, f.Port())
	if err = controller.Run(4, stopCh); err != nil {
		return err
	}

	return nil
}
