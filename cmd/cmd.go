package cmd

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/JulienBalestra/kube-sync/pkg/kubesync"
)

const programName = "kube-sync"

var viperConfig = viper.New()

// NewCommand creates a new command and return a return code
func NewCommand() (*cobra.Command, *int) {
	var verbose int
	var exitCode int

	rootCommand := &cobra.Command{
		Use:   fmt.Sprintf("%s command line", programName),
		Short: "Use this command to do sync a configmap between namespaces",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			flag.Lookup("alsologtostderr").Value.Set("true")
			flag.Lookup("v").Value.Set(strconv.Itoa(verbose))
		},
		Args: cobra.ExactArgs(2),
		Example: fmt.Sprintf(`
%s [namespace] [configmap name] --sync-interval 1m
`, programName),
		Run: func(cmd *cobra.Command, args []string) {
			s, err := newSync(args[0], args[1])
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 1
				return
			}
			err = s.Sync()
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 2
				return
			}
		},
	}

	rootCommand.PersistentFlags().IntVarP(&verbose, "verbose", "v", 0, "verbose level")
	viperConfig.SetDefault("kubeconfig-path", "")
	rootCommand.PersistentFlags().String("kubeconfig-path", viperConfig.GetString("kubeconfig-path"), "kubernetes config path, leave empty for inCluster config")
	viperConfig.BindPFlag("kubeconfig-path", rootCommand.PersistentFlags().Lookup("kubeconfig-path"))

	viperConfig.SetDefault("sync-interval", 1*time.Minute)
	rootCommand.PersistentFlags().String("sync-interval", viperConfig.GetString("sync-interval"), "interval for each sync")
	viperConfig.BindPFlag("sync-interval", rootCommand.PersistentFlags().Lookup("sync-interval"))

	viperConfig.SetDefault("disable-prometheus-exporter", false)
	rootCommand.PersistentFlags().Bool("disable-prometheus-exporter", viperConfig.GetBool("disable-prometheus-exporter"), "disable prometheus exporter /metrics")
	viperConfig.BindPFlag("disable-prometheus-exporter", rootCommand.PersistentFlags().Lookup("disable-prometheus-exporter"))

	viperConfig.SetDefault("prometheus-exporter-bind", "0.0.0.0:8484")
	rootCommand.PersistentFlags().String("prometheus-exporter-bind", viperConfig.GetString("prometheus-exporter-bind"), "prometheus exporter bind address")
	viperConfig.BindPFlag("prometheus-exporter-bind", rootCommand.PersistentFlags().Lookup("prometheus-exporter-bind"))

	return rootCommand, &exitCode
}

func newSync(namespace, configmapName string) (*kubesync.KubeSync, error) {
	conf := &kubesync.Config{
		SourceConfigmapNamespace: namespace,
		SourceConfigmapName:      configmapName,
		SyncInterval:             viperConfig.GetDuration("sync-interval"),
	}
	if !viperConfig.GetBool("disable-prometheus-exporter") {
		conf.PrometheusExporterBindAddress = viperConfig.GetString("prometheus-exporter-bind")
	}
	return kubesync.NewKubeSync(viperConfig.GetString("kubeconfig-path"), conf)
}
