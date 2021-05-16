package vipertest

import (
	"fmt"
	"github.com/spf13/viper"
)

func EchoFlag() {
	fmt.Print(viper.Get("prometheus-configmap"))
	fmt.Print(viper.Get("prometheus-configmap"))
	fmt.Print(viper.Get("prometheus-configmap"))
}
