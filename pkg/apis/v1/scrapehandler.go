package v1

import (
	"flame/pkg/utils/fshare"
	"flame/pkg/utils/k8s"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

func getScrapeHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		i, ok := f.PromController.Instance.ScrapeMap[c.Param("job_name")]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"res": c.GetString("resId"),
				"msg": c.Param("job_name") + " not found.",
			})
			return
		}
		scrapeConfig := f.PromController.Instance.Config.ScrapeConfigs[i]
		c.JSON(http.StatusOK, gin.H{
			"res":  c.GetString("resId"),
			"data": scrapeConfig,
			"msg":  "get " + c.Param("job_name") + " success.",
		})
	}
}

// listScrapeHandler 查询scrape列表接口
// @Summary 查询scrape列表接口
// @Description 可按labels进行筛选查询
// @Tags prom
// @Accept application/json
// @Param labels[psa] query string false "可通过label筛选，example：?labels[k1]=v1"
// @Param labels[exporter_type] query string false "可通过多个label筛选获取交集，example：?labels[k1]=v1&labels[k2]=v2"
// @Success 200 {object} _ResponseScrapeList "返回值"
// @Router /scrape [get]
func listScrapeHandler(f *Flame) gin.HandlerFunc {
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
				res = fshare.Intersect(res, jobs)
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

func removeScrapeHandler(f *Flame) gin.HandlerFunc {
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
