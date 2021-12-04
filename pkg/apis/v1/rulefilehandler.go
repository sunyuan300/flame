package v1

import (
	"flame/pkg/factory/rules"
	"flame/pkg/utils/k8s"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"net/http"
)

func listRuleFileHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		var fileName []string
		for k := range f.RulesController.Instance.AllRulesGroups {
			fileName = append(fileName, k)
		}
		c.JSON(http.StatusOK, gin.H{
			"req": id,
			"data": gin.H{
				"items": fileName,
				"count": len(fileName),
			},
			"msg": "get rule file list success.",
		})
	}
}

func addRuleFileHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		var fileName *rules.RuleFile
		if err := c.ShouldBindJSON(&fileName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "add rule file failed: parameter err.",
			})
			return
		}

		f.RulesController.Instance.Lock.Lock()
		if f.RulesController.Instance.ExistsRuleFileName(fileName.FileName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "rule file_name existed.",
			})
			return
		}

		f.RulesController.Instance.AllRulesGroups[fileName.FileName] = &rulefmt.RuleGroups{}
		ruleData := make(map[string]string)
		for k, v := range f.RulesController.Instance.AllRulesGroups {
			groups, err := yaml.Marshal(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"req": id,
					"msg": "add rule file failed: marshal err.",
				})
				return
			}
			ruleData[k] = string(groups)
		}

		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("rules-configmap"), ruleData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "add rule file failed: update rule config failed.",
			})
			return
		}

		f.PromController.Instance.Lock.Lock()
		f.PromController.Instance.Config.RuleFiles = append(f.PromController.Instance.Config.RuleFiles, viper.GetString("rule-dir")+fileName.FileName)
		promData := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}

		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), promData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "add rule file failed: update prom config failed.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"req": id,
			"msg": "add rule file success.",
		})
	}
}

func removeRuleFileHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		fileName := c.Param("file_name")

		f.RulesController.Instance.Lock.Lock()
		if !f.RulesController.Instance.ExistsRuleFileName(fileName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "file_name not exist.",
			})
			return
		}

		delete(f.RulesController.Instance.AllRulesGroups, fileName)
		data := make(map[string]string)
		for k, v := range f.RulesController.Instance.AllRulesGroups {
			groups, err := yaml.Marshal(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"req": id,
					"msg": "add rule file failed: marshal err.",
				})
				return
			}
			data[k] = string(groups)
		}

		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("rules-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "add rule file failed: update failed.",
			})
			return
		}

		f.PromController.Instance.Lock.Lock()
		for i, v := range f.PromController.Instance.Config.RuleFiles {
			if v == (viper.GetString("rule-dir") + fileName) {
				f.PromController.Instance.Config.RuleFiles = append(f.PromController.Instance.Config.RuleFiles[:i],
					f.PromController.Instance.Config.RuleFiles[i+1:]...)
			}
		}
		promData := map[string]string{
			viper.GetString("prometheus.yml"): f.PromController.Instance.Config.String(),
		}

		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("prometheus-configmap"), promData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "remove rule file failed: update prom config failed.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"req": id,
			"msg": "add rule file success.",
		})
	}
}
