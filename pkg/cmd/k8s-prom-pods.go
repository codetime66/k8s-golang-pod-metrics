package cmd

import (
    "k8s.io/client-go/tools/clientcmd"
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

func StartUp(kubeconfig string, addr string, interval string) {

    // use the current context in kubeconfig
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
	panic(err.Error())
    }
    // creates the clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

    go func() {

     for {

       fmt.Printf("Performing an api http request - Current Unix Time: %v\n", time.Now().Unix())

       var pods PodMetricsList
       err = getMetrics(clientset, &pods)
       if err != nil {
          panic(err.Error())
       }

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

             memUsage.WithLabelValues(m.Metadata.Namespace, m.Metadata.Name, c.Name, "Ki").Set(n_mem_in_kb)
	     cpuUsage.WithLabelValues(m.Metadata.Namespace, m.Metadata.Name, c.Name, "n").Set(n_cpu_in_n)
          }
       }

       n_interval, err := strconv.Atoi(interval)
       if err != nil {
           n_interval=15
       }
       time.Sleep( time.Duration(n_interval) * time.Second)

     }

    }()

    fmt.Print("Server listening to ")
    fmt.Print(addr)
    fmt.Println(", metrics exposed on /metrics endpoint")
    http.Handle("/metrics", promhttp.Handler())
    log.Fatal(http.ListenAndServe(addr, nil))
}
