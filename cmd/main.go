package main

// @title flame接口说明
// @version 1.0
// @description flame提供相应的接口，用于prometheus的监控注册、告警添加

// @contact.name solid

// @host localhost:8080
// @BasePath /api
import (
	"flame/pkg/apis/v1"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"time"
)

func init() {
	rootCmd.Flags().StringP("namespace", "n", "", "prometheus cluster namespace in k8s")
	rootCmd.Flags().StringP("prometheus-configmap", "", "prometheus-config", "configmap name for prometheus.yaml")
	rootCmd.Flags().StringP("prometheus.yml", "", "prometheus.yml", "prometheus.yml file name")
	rootCmd.Flags().StringP("rules-configmap", "", "rules-config", "configmap name for prometheus.yaml")
	rootCmd.Flags().StringP("rule-dir", "", "/etc/prometheus/rules/", "rule dir, ends with /")
	rootCmd.Flags().StringP("env", "", "dev", "run env")
	_ = viper.BindPFlag("namespace", rootCmd.Flags().Lookup("namespace"))
	_ = viper.BindPFlag("prometheus-configmap", rootCmd.Flags().Lookup("prometheus-configmap"))
	_ = viper.BindPFlag("prometheus.yml", rootCmd.Flags().Lookup("prometheus.yml"))
	_ = viper.BindPFlag("rules-configmap", rootCmd.Flags().Lookup("rules-configmap"))
	_ = viper.BindPFlag("rule-dir", rootCmd.Flags().Lookup("rule-dir"))
	_ = viper.BindPFlag("env", rootCmd.Flags().Lookup("env"))
}

var rootCmd = &cobra.Command{
	Use:   "flame",
	Short: "prometheus's targets、rules and alerts config manager",
	Long:  "pm is an apis interface used to manage prometheus's targets, rules and alerts.",
	Run: func(cmd *cobra.Command, args []string) {
		v1.NewAndRunFlame()
	},
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	rand.Seed(time.Now().UnixNano())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
