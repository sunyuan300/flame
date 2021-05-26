package v1

import (
	"flame/pkg/utils/fshare"
	"flame/pkg/utils/k8s"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

// getScrapeHandler 获取scrape详细信息
// @Summary 查询scrape详情接口
// @Description 可获取整个scrape的
// @Tags prom
// @Accept application/json
// @Param job_name path string ture "精确的scrape名称"
// @Success 200 {object} _ResponseScrapeInfo "返回值"
// @Router /scrape/{job_name} [get]
func getScrapeHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		i, ok := f.PromController.Instance.ScrapeMap[c.Param("job_name")]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": c.Param("job_name") + " not found.",
			})
			return
		}
		scrapeConfig := f.PromController.Instance.Config.ScrapeConfigs[i]
		c.JSON(http.StatusOK, gin.H{
			"req":  id,
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
		id, _ := c.Get("reqId")
		queryMap := c.QueryMap("labels")
		var req []string
		if len(queryMap) == 0 {
			for k := range f.PromController.Instance.ScrapeMap {
				req = append(req, k)
			}
			c.JSON(http.StatusOK, gin.H{
				"req": id,
				"data": gin.H{
					"items": req,
					"count": len(req),
				},
				"msg": "get scrape list success.",
			})
			return
		} else {
			for k, v := range queryMap {
				jobs := f.PromController.Instance.LabelsMap[k][v]
				req = fshare.Intersect(req, jobs)
			}
			req = fshare.SliceDeduplication(req)
			c.JSON(http.StatusOK, gin.H{
				"req": id,
				"data": gin.H{
					"items": req,
					"count": len(req),
				},
				"msg": "get scrape list success.",
			})
		}
	}
}

// removeScrapeHandler 删除指定名称的scrape
// @Summary 删除scrape
// @Description 删除指定名称的scrape
// @Tags prom
// @Accept application/json
// @Param job_name path string ture "精确的scrape名称"
// @Success 200 {object} _ResponseRemoveScrape "返回值"
// @Router /scrape/{job_name} [delete]
func removeScrapeHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		jobName := c.Param("job_name")

		i := f.PromController.Instance.ScrapeMap[jobName]
		f.PromController.Instance.Config.ScrapeConfigs = append(f.PromController.Instance.Config.ScrapeConfigs[:i],
			f.PromController.Instance.Config.ScrapeConfigs[i+1:]...)
		data := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}
		f.PromController.Instance.Lock.Lock()

		if !f.PromController.Instance.ExistsJobName(jobName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "job_name not exist.",
			})
			return
		}

		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "remove node scrape failed: update failed.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"req": id,
			"msg": "remove node scrape success.",
		})
	}
}
