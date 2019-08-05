package main

import (
    "flag"
    "path/filepath"
    "os"

    "github.com/codetime66/k8s-golang-pod-metrics/pkg/cmd"
)

func main() {
    var kubeconfig *string
    if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
    } else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
    }
    var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
    var interval = flag.String("api-get-interval", "15", "The interval for api requests.")

    flag.Parse()

    cmd.StartUp(*kubeconfig, *addr, *interval)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
