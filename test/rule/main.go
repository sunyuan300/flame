package main

import (
	"flame/pkg/utils/k8s"
	"flame/pkg/watcher"
	"github.com/spf13/viper"
)

func main() {
	viper.Set("env", "dev")
	viper.Set("namespace", "prometheus")
	viper.Set("rules-configmap", "rules-config")
	clientSet := k8s.NewK8sClient()
	//promController := k8s.NewPrometheusController(clientSet)
	//promController.Start()
	c := watcher.NewRulesController(clientSet)
	go c.RunRulesController()
	select {}
}
