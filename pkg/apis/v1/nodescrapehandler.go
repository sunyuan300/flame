package v1

import (
	"flame/pkg/factory/scrape"
	"flame/pkg/utils/k8s"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

// addNodeScrapeHandler 创建node_scrape
// @Summary 创建node scrape
// @Description 创建exporter_type是node的scrape
// @Tags prom
// @Accept application/json
// @Param node_scrape body scrape.NodeScrape ture "scrape的必填参数"
// @Success 200 {object} _ResponseAddNodeScrape "返回值"
// @Router /node_scrape [POST]
func addNodeScrapeHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ns scrape.NodeScrape
		if err := c.ShouldBindJSON(&ns); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"res": c.GetString("resId"),
				"msg": "add node scrape failed: parameter err.",
			})
			return
		}

		if f.PromController.Instance.ExistsJobName(ns.JobName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"res": c.GetString("resId"),
				"msg": "job_name existed.",
			})
			return
		}

		newScrapeConfig := ns.Marshal()
		f.PromController.Instance.Config.ScrapeConfigs = append(f.PromController.Instance.Config.ScrapeConfigs, newScrapeConfig)
		data := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}
		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"res": c.GetString("resId"),
				"msg": "add node scrape failed: update failed.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"res": c.GetString("resId"),
			"msg": "add node scrape success.",
		})
	}
}

// updateNodeScrapeHandler 更新指定scrape名称的静态监控目标
// @Summary 更新scrape的监控目标
// @Description 更新指定scrape名称的静态监控目标
// @Tags prom
// @Accept application/json
// @Param job_name path string ture "scrape的必须参数"
// @Param node_scrape body _RequestUpdateNodeScrape false "一个或多个scrape的更新参数"
// @Success 200 {object} _ResponseUpdateNodeScrape "返回值"
// @Router /node_scrape/{job_name} [POST]
func updateNodeScrapeHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ns scrape.NodeScrape
		if err := c.ShouldBindJSON(&ns); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"res": c.GetString("resId"),
				"msg": "update node scrape failed: parameter err.",
			})
			return
		}

		if !f.PromController.Instance.ExistsJobName(c.Param("job_name")) {
			c.JSON(http.StatusBadRequest, gin.H{
				"res": c.GetString("resId"),
				"msg": "job not found.",
			})
			return
		}

		newScrapeConfig := ns.Marshal()
		i := f.PromController.Instance.ScrapeMap[c.Param("job_name")]
		if newScrapeConfig.ScrapeInterval != 0 {
			f.PromController.Instance.Config.ScrapeConfigs[i].ScrapeInterval = newScrapeConfig.ScrapeInterval
		}
		if newScrapeConfig.ScrapeTimeout != 0 {
			f.PromController.Instance.Config.ScrapeConfigs[i].ScrapeTimeout = newScrapeConfig.ScrapeTimeout
		}
		if newScrapeConfig.MetricsPath != "" {
			f.PromController.Instance.Config.ScrapeConfigs[i].MetricsPath = newScrapeConfig.MetricsPath
		}
		if len(newScrapeConfig.RelabelConfigs) != 0 {
			f.PromController.Instance.Config.ScrapeConfigs[i].RelabelConfigs = newScrapeConfig.RelabelConfigs
		}
		data := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}
		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"res": c.GetString("resId"),
				"msg": "update node scrape failed: update failed.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"res": c.GetString("resId"),
			"msg": "update node scrape success.",
		})
	}
}
