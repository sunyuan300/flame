package v1

import (
	"flame/pkg/factory/scrape"
	"flame/pkg/utils/fshare"
	"flame/pkg/utils/k8s"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

func listNodeScrapeHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		queryMap := c.QueryMap("labels")
		var res []string
		if len(queryMap) == 0 {
			for k := range f.PromController.Instance.ScrapeMap {
				res = append(res, k)
			}
			c.JSON(http.StatusOK, gin.H{
				"res":  c.GetString("resId"),
				"data": res,
				"msg":  "get scrape list success.",
			})
			return
		} else {
			for k, v := range queryMap {
				jobs := f.PromController.Instance.LabelsMap[k][v]
				res = append(res, jobs...)
			}
			res = fshare.SliceDeduplication(res)
			c.JSON(http.StatusOK, gin.H{
				"res":  c.GetString("resId"),
				"data": res,
				"msg":  "get scrape list success.",
			})
		}
	}
}

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

func removeNodeScrapeHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobName := c.Param("job_name")
		if !f.PromController.Instance.ExistsJobName(jobName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"res": c.GetString("resId"),
				"msg": "job_name not exist.",
			})
			return
		}

		i := f.PromController.Instance.ScrapeMap[jobName]
		f.PromController.Instance.Config.ScrapeConfigs = append(f.PromController.Instance.Config.ScrapeConfigs[:i],
			f.PromController.Instance.Config.ScrapeConfigs[i+1:]...)
		data := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}
		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"res": c.GetString("resId"),
				"msg": "remove node scrape failed: update failed.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"res": c.GetString("resId"),
			"msg": "remove node scrape success.",
		})
	}
}
