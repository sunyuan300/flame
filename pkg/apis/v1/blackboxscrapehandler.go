package v1

import (
	"flame/pkg/factory/scrape"
	"flame/pkg/utils/k8s"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

func addBlackboxScrapeHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		var bs scrape.BlackboxScrape
		if err := c.ShouldBindJSON(&bs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "add blackbox scrape failed: parameter err.",
			})
			return
		}

		if f.PromController.Instance.ExistsJobName(bs.JobName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "job_name existed.",
			})
			return
		}

		newScrapeConfig, err := bs.Marshal()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "add blackbox scrape failed: marshal failed.",
			})
			return
		}

		f.PromController.Instance.Config.ScrapeConfigs = append(f.PromController.Instance.Config.ScrapeConfigs, newScrapeConfig)
		data := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}
		f.PromController.Instance.Lock.Lock()
		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "add node scrape failed: update failed.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"req": id,
			"msg": "add node scrape success.",
		})
	}
}

func updateBlackboxScrapeHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		var bs scrape.BlackboxScrape
		if err := c.ShouldBindJSON(&bs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "update blackbox scrape failed: parameter err.",
			})
			return
		}

		if !f.PromController.Instance.ExistsJobName(c.Param("job_name")) {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "job not found.",
			})
			return
		}

		newScrapeConfig, err := bs.Marshal()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "update blackbox scrape failed: marshal failed.",
			})
			return
		}
		i := f.PromController.Instance.ScrapeMap[c.Param("job_name")]
		if newScrapeConfig.ScrapeInterval != 0 {
			f.PromController.Instance.Config.ScrapeConfigs[i].ScrapeInterval = newScrapeConfig.ScrapeInterval
		}
		if newScrapeConfig.ScrapeTimeout != 0 {
			f.PromController.Instance.Config.ScrapeConfigs[i].ScrapeTimeout = newScrapeConfig.ScrapeTimeout
		}
		if len(newScrapeConfig.Params.Get("module")) != 0 {
			f.PromController.Instance.Config.ScrapeConfigs[i].Params = newScrapeConfig.Params
		}
		if newScrapeConfig.MetricsPath != "" {
			f.PromController.Instance.Config.ScrapeConfigs[i].MetricsPath = newScrapeConfig.MetricsPath
		}
		if bs.BlackboxTarget != "" && len(bs.Labels) != 0 {
			f.PromController.Instance.Config.ScrapeConfigs[i].RelabelConfigs = newScrapeConfig.RelabelConfigs
		} else if (bs.BlackboxTarget == "" && len(bs.Labels) != 0) || (bs.BlackboxTarget != "" && len(bs.Labels) == 0) {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "target and labels must exist together or not exist together.",
			})
			return
		}
		data := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}
		f.PromController.Instance.Lock.Lock()
		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "update blackbox scrape failed: update failed.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"req": id,
			"msg": "update blackbox scrape success.",
		})
	}
}
