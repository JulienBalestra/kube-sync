package kubesync

import (
	"fmt"
	"github.com/JulienBalestra/kube-sync/pkg/utils/kubeclient"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	prometheusExporterPath = "/metrics"
	kubeSyncAnnotationKey  = "kube-sync/source"
)

// Config contains static configuration
type Config struct {
	SyncInterval                  time.Duration
	SourceConfigmapName           string
	SourceConfigmapNamespace      string
	PrometheusExporterBindAddress string
}

// KubeSync contains the state
type KubeSync struct {
	Conf *Config

	kubeClient *kubeclient.KubeClient

	promSyncLatency    prometheus.Histogram
	promSuccessCounter prometheus.Counter
	promErrorCounter   prometheus.Counter
}

// RegisterPrometheusMetrics is a convenient function to create and register prometheus metrics
func RegisterPrometheusMetrics(s *KubeSync) error {
	labels := prometheus.Labels{
		"ns": s.Conf.SourceConfigmapNamespace,
		"cm": s.Conf.SourceConfigmapName,
	}
	s.promSyncLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        "kubernetes_cm_update_latency_seconds",
		Help:        "Latency of configmap update",
		ConstLabels: labels,
	})
	s.promSuccessCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "kubernetes_cm_updates",
		Help:        "Total number of Kubernetes configmap successfully updated",
		ConstLabels: labels,
	})
	s.promErrorCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name:        "kubernetes_cm_update_errors",
		Help:        "Total number of Kubernetes configmap updated errors",
		ConstLabels: labels,
	})
	err := prometheus.Register(s.promSyncLatency)
	if err != nil {
		return err
	}
	err = prometheus.Register(s.promSuccessCounter)
	if err != nil {
		return err
	}
	err = prometheus.Register(s.promErrorCounter)
	if err != nil {
		return err
	}
	return nil
}

// NewKubeSync creates a new KubeSync
func NewKubeSync(kubeConfigPath string, conf *Config) (*KubeSync, error) {
	if conf.SyncInterval == 0 {
		err := fmt.Errorf("invalid value for SyncInterval: %s", conf.SyncInterval.String())
		glog.Errorf("Cannot use the provided config: %v", err)
		return nil, err
	}
	k, err := kubeclient.NewKubeClient(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	s := &KubeSync{
		kubeClient: k,
		Conf:       conf,
	}
	err = RegisterPrometheusMetrics(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *KubeSync) processSync() error {
	start := time.Now()
	sourceCM, err := s.kubeClient.GetKubernetesClient().CoreV1().ConfigMaps(s.Conf.SourceConfigmapNamespace).Get(s.Conf.SourceConfigmapName, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Cannot get cm/%s in ns %s: %v", s.Conf.SourceConfigmapName, s.Conf.SourceConfigmapNamespace, err)
		return err
	}
	allNamespaces, err := s.kubeClient.GetKubernetesClient().CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		glog.Errorf("Cannot list all namespaces: %v", err)
		return err
	}

	// copy and reset some fields
	newCM := sourceCM.DeepCopy()
	newCM.ResourceVersion = ""
	newCM.Namespace = ""
	newCM.UID = ""
	newCM.GenerateName = ""
	newCM.SelfLink = ""
	newCM.CreationTimestamp.Reset()
	if newCM.Annotations == nil {
		newCM.Annotations = make(map[string]string)
	}
	kubeSyncAnnotationValue := fmt.Sprintf(`{"namespace":%q,"name":%q,"uid":%q,"resourceVersion":%q,"last-update":%d}`, sourceCM.Namespace, sourceCM.Name, sourceCM.UID, sourceCM.ResourceVersion, start.Unix())
	newCM.Annotations[kubeSyncAnnotationKey] = kubeSyncAnnotationValue
	glog.V(0).Infof("Configmap create/update with the following annotation %s: %s", kubeSyncAnnotationKey, kubeSyncAnnotationValue)

	glog.V(1).Infof("Configmap to sync across all namespaces is: %s", newCM.String())

	var errs []string
	for _, ns := range allNamespaces.Items {
		// do not override the source
		if ns.Name == sourceCM.Namespace {
			glog.V(1).Infof("Skipping sync over the namespace %s: namespace of the source configmap", ns.Name)
			continue
		}
		newCM.Namespace = ns.Name
		_, err = s.kubeClient.GetKubernetesClient().CoreV1().ConfigMaps(ns.Name).Update(newCM)
		if err != nil && errors.IsNotFound(err) {
			glog.V(0).Infof("Creating cm/%s in ns %s", newCM.Name, ns.Name)
			_, err = s.kubeClient.GetKubernetesClient().CoreV1().ConfigMaps(ns.Name).Create(newCM)
		}
		if err != nil {
			glog.Errorf("Unexpected error while applying cm/%s in ns %s: %v", newCM.Name, ns.Name, err)
			s.promErrorCounter.Inc()
			errs = append(errs, err.Error())
			continue
		}
		glog.V(0).Infof("Successfully sync cm/%s in ns %s with the ns %s", sourceCM.Name, sourceCM.Namespace, ns.Name)
		s.promSuccessCounter.Inc()
	}
	if errs == nil {
		s.promSyncLatency.Observe(time.Now().Sub(start).Seconds())
		return nil
	}
	return fmt.Errorf("%s", strings.Join(errs, "; "))
}

func (s *KubeSync) registerListeners() {
	if s.Conf.PrometheusExporterBindAddress != "" {
		promRouter := mux.NewRouter()
		promRouter.Path(prometheusExporterPath).Methods("GET").Handler(promhttp.Handler())
		promServer := &http.Server{
			Handler:      promRouter,
			Addr:         s.Conf.PrometheusExporterBindAddress,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		glog.V(0).Infof("Starting prometheus exporter on %s%s", s.Conf.PrometheusExporterBindAddress, prometheusExporterPath)
		go promServer.ListenAndServe()
	}
	// Known issue with Mux and the registering of pprof:
	// https://stackoverflow.com/questions/19591065/profiling-go-web-application-built-with-gorillas-mux-with-net-http-pprof
	const pprofBind = "127.0.0.1:6060"
	pprofRouter := mux.NewRouter()
	pprofRouter.HandleFunc("/debug/pprof/", pprof.Index)
	pprofRouter.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofRouter.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofRouter.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofRouter.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	pprofServer := &http.Server{
		Handler:      pprofRouter,
		Addr:         pprofBind,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	glog.V(0).Infof("Starting pprof on %s/debug/pprof", pprofBind)
	go pprofServer.ListenAndServe()
}

// Sync start the loop
func (s *KubeSync) Sync() error {
	sigCh := make(chan os.Signal)
	defer close(sigCh)

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	// Sync once and fail fast to crash the Pod in case of error
	glog.V(0).Infof("Starting to sync ...")
	err := s.processSync()
	if err != nil {
		return err
	}
	glog.V(0).Infof("Successfully sync, now syncing every %s", s.Conf.SyncInterval.String())

	s.registerListeners()

	ticker := time.NewTicker(s.Conf.SyncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = s.processSync()

		case sig := <-sigCh:
			glog.V(0).Infof("Signal %s received, stopping", sig.String())
			return nil
		}
	}
}
