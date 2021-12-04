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

// UpdateTargetHandler 更新指定node_scrape名称的参数
// @Summary 更新node_scrape参数
// @Description 更新指定node_scrape名称的参数
// @Tags prom
// @Accept application/json
// @Param job_name path string ture "scrape的名称"
// @Param targets body target.StaticTarget true "target的数组"
// @Success 200 {object} _ResponseUpdateNodeScrape "返回值"
// @Router /scrape/{job_name}/static_target [POST]
func UpdateTargetHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		var st target.StaticTarget
		jobName := c.Param("job_name")

		if err := c.ShouldBindJSON(&st); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"res": id,
				"msg": "update target failed: parameter err.",
			})
			return
		}
		var targets []model.LabelSet
		for _, v := range st.Targets {
			targets = append(targets, model.LabelSet{model.AddressLabel: model.LabelValue(v)})
		}

		f.PromController.Instance.Lock.Lock()
		if !f.PromController.Instance.ExistsJobName(jobName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"res": id,
				"msg": "job_name not exist.",
			})
			return
		}
		i := f.PromController.Instance.ScrapeMap[jobName]
		f.PromController.Instance.Config.ScrapeConfigs[i].ServiceDiscoveryConfigs = discovery.Configs{discovery.StaticConfig{{Targets: targets}}}
		data := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}

		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"res": id,
				"msg": "add target failed: update failed.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"res": id,
			"msg": "add target success.",
		})
	}
}
