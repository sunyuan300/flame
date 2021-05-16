package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func getNodeScrapeHandler(f *Flame) gin.HandlerFunc {
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
