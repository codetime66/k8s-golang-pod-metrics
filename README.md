# k8s-golang-pod-metrics
custom prometheus k8s metrics

go mod init github.com/codetime66/k8s-golang-pod-metrics

./bin/k8s-prom-pods --help
Usage of ./bin/k8s-prom-pods:
  -kubeconfig string
    	(optional) absolute path to the kubeconfig file (default "/home/carlosfe/.kube/config")
  -listen-address string
    	The address to listen on for HTTP requests. (default ":8080")


./bin/k8s-prom-pods --kubeconfig ~/projects/kubeland/zubernetes/.kube/conf

---------
kubectl --kubeconfig ~/projects/kubeland/hzubernetes/.kube/conf -n infra exec -it mytool-fdf57d5bb-hzb7p -- curl -v -k https://kubernetes.default.svc:443/apis/metrics.k8s.io/v1beta1/pods
