package v1

import (
	"flame/pkg/factory/rules"
	"flame/pkg/utils/k8s"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"net/http"
)

func updateRulesHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		fileName := c.Param("file_name")
		group := &rules.RuleGroup{
			GroupName: c.Param("group_name"),
		}
		if err := c.ShouldBindJSON(&group.Rules); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "add rule group failed: parameter err.",
			})
			return
		}

		ruleNodes, err := group.Marshal()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "add rules failed: marshal err.",
			})
			return
		}
		for i, v := range f.RulesController.Instance.AllRulesGroups[fileName].Groups {
			if v.Name == group.GroupName {
				f.RulesController.Instance.AllRulesGroups[fileName].Groups[i].Rules = ruleNodes

				data := make(map[string]string)
				for k, v := range f.RulesController.Instance.AllRulesGroups {
					groups, err := yaml.Marshal(v)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"req": id,
							"msg": "add rule group failed: marshal err.",
						})
						return
					}
					data[k] = string(groups)
				}
				f.RulesController.Instance.Lock.Lock()
				if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("rules-configmap"), data); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"req": id,
						"msg": "add rule group failed: update rule config failed.",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"req": id,
					"msg": "update rules success.",
				})
				return
			}
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"req": id,
			"msg": "update rules failed.",
		})
	}
}
