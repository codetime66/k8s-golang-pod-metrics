package main

import (
    "flag"
    "path/filepath"
    "k8s.io/client-go/tools/clientcmd"
    "os"
    "encoding/json"
    "fmt"
    "time"

    "k8s.io/client-go/kubernetes"

    "strconv"
    "strings"
    "log"
    "net/http"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// PodMetricsList : PodMetricsList
type PodMetricsList struct {
    Kind       string `json:"kind"`
    APIVersion string `json:"apiVersion"`
    Metadata   struct {
        SelfLink string `json:"selfLink"`
    } `json:"metadata"`
    Items []struct {
        Metadata struct {
            Name              string    `json:"name"`
            Namespace         string    `json:"namespace"`
            SelfLink          string    `json:"selfLink"`
            CreationTimestamp time.Time `json:"creationTimestamp"`
        } `json:"metadata"`
        Timestamp  time.Time `json:"timestamp"`
        Window     string    `json:"window"`
        Containers []struct {
            Name  string `json:"name"`
            Usage struct {
                CPU    string `json:"cpu"`
                Memory string `json:"memory"`
            } `json:"usage"`
        } `json:"containers"`
    } `json:"items"`
}

var (
	memUsage = prometheus.NewGaugeVec(
             prometheus.GaugeOpts{
                  Namespace: "k8s_metrics",
                  Name:      "container_memory_usage",
                  Help:      "container memory usage",
             },
             []string{"namespace", "pod", "container", "unit"},
        )

        cpuUsage = prometheus.NewGaugeVec(
             prometheus.GaugeOpts{
                  Namespace: "k8s_metrics",
                  Name:      "container_cpu_usage",
                  Help:      "container cpu usage",
             },
             []string{"namespace", "pod", "container", "unit"},
        )
)

func init() {
	prometheus.MustRegister(memUsage)
	prometheus.MustRegister(cpuUsage)
        prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}

func getMetrics(clientset *kubernetes.Clientset, pods *PodMetricsList) error {
    data, err := clientset.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/pods").DoRaw()
    if err != nil {
        return err
    }
    err = json.Unmarshal(data, &pods)
    return err
}

func main() {
    var kubeconfig *string
    if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
    } else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
    }
    var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
    flag.Parse()

    // use the current context in kubeconfig
    config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
    if err != nil {
	panic(err.Error())
    }
    // creates the clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }
    var pods PodMetricsList
    err = getMetrics(clientset, &pods)
    if err != nil {
        panic(err.Error())
    }

    go func() {
       for _, m := range pods.Items {

          for _, c := range m.Containers {

             s_mem_in_kb := strings.TrimSuffix(c.Usage.Memory, "Ki")
	     s_cpu_in_n := strings.TrimSuffix(c.Usage.CPU, "n")

             n_mem_in_kb, err := strconv.ParseFloat(s_mem_in_kb, 64)
             if err != nil {
                n_mem_in_kb=0
             }

             n_cpu_in_n, err := strconv.ParseFloat(s_cpu_in_n, 64)
             if err != nil {
                n_cpu_in_n=0
             }

             memUsage.WithLabelValues(m.Metadata.Namespace, m.Metadata.Name, c.Name, "Ki").Add(n_mem_in_kb)
	     cpuUsage.WithLabelValues(m.Metadata.Namespace, m.Metadata.Name, c.Name, "n").Add(n_cpu_in_n)
          }
       }
    }()

    fmt.Print("Server listening to ")
    fmt.Print(*addr)
    fmt.Println(", metrics exposed on /metrics endpoint")
    http.Handle("/metrics", promhttp.Handler())
    log.Fatal(http.ListenAndServe(*addr, nil))
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
