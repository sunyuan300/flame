package v1

import (
	"flame/pkg/factory/target"
	"flame/pkg/utils/k8s"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery"
	"github.com/spf13/viper"
	"net/http"
)

func UpdateTargetHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		var st target.StaticTarget
		jobName := c.Param("job_name")
		if !f.PromController.Instance.ExistsJobName(jobName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"res": c.GetString("resId"),
				"msg": "job_name not exist.",
			})
			return
		}

		if err := c.ShouldBindJSON(&st); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"res": c.GetString("resId"),
				"msg": "update target failed: parameter err.",
			})
			return
		}
		var targets []model.LabelSet
		for _, v := range st.Targets {
			targets = append(targets, model.LabelSet{model.AddressLabel: model.LabelValue(v)})
		}
		i := f.PromController.Instance.ScrapeMap[jobName]
		f.PromController.Instance.Config.ScrapeConfigs[i].ServiceDiscoveryConfigs = discovery.Configs{discovery.StaticConfig{{Targets: targets}}}
		data := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}
		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"res": c.GetString("resId"),
				"msg": "add target failed: update failed.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"res": c.GetString("resId"),
			"msg": "add target success.",
		})
	}
}
