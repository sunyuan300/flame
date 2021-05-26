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

func listRuleGroupHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		fileName := c.Param("file_name")
		var ruleGroupItems []string
		for _, v := range f.RulesController.Instance.AllRulesGroups[fileName].Groups {
			ruleGroupItems = append(ruleGroupItems, v.Name)
		}
		c.JSON(http.StatusOK, gin.H{
			"req": id,
			"msg": "list rule group success.",
			"data": gin.H{
				"items": ruleGroupItems,
				"count": len(ruleGroupItems),
			},
		})
	}
}

func getRuleGroupHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		fileName := c.Param("file_name")
		groupName := c.Param("group_name")
		if !f.RulesController.Instance.ExistsRuleFileName(fileName) {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "file_name not exist.",
			})
			return
		}

		for _, v := range f.RulesController.Instance.AllRulesGroups[fileName].Groups {
			if v.Name == groupName {
				c.JSON(http.StatusOK, gin.H{
					"req":  id,
					"msg":  "get rules success.",
					"data": rules.UnMarshal(v.Rules),
				})
				return
			}
		}
	}
}

func addRuleGroupHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		fileName := c.Param("file_name")
		var groupName *rules.RuleGroup
		if err := c.ShouldBindJSON(&groupName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"req": id,
				"msg": "add rule group failed: parameter err.",
			})
			return
		}

		f.RulesController.Instance.Lock.Lock()
		for _, v := range f.RulesController.Instance.AllRulesGroups[fileName].Groups {
			if v.Name == groupName.GroupName {
				c.JSON(http.StatusBadRequest, gin.H{
					"req": id,
					"msg": "add rule group failed: group name exists.",
				})
				return
			}
		}

		var group rulefmt.RuleGroup
		group.Name = groupName.GroupName
		f.RulesController.Instance.AllRulesGroups[fileName].Groups = append(f.RulesController.Instance.AllRulesGroups[fileName].Groups, group)

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

		if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("rules-configmap"), data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"req": id,
				"msg": "add rule group failed: update rule config failed.",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"req": id,
			"msg": "add rule group success.",
		})
	}
}

func removeRuleGroupHandler(f *Flame) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := c.Get("reqId")
		fileName := c.Param("file_name")
		groupName := c.Param("group_name")

		f.RulesController.Instance.Lock.Lock()
		for i, v := range f.RulesController.Instance.AllRulesGroups[fileName].Groups {
			if v.Name == groupName {
				f.RulesController.Instance.AllRulesGroups[fileName].Groups = append(f.RulesController.Instance.AllRulesGroups[fileName].Groups[:i],
					f.RulesController.Instance.AllRulesGroups[fileName].Groups[i+1:]...)

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

				if err := k8s.ConfigMapUpdate(f.K8sClient, viper.GetString("rules-configmap"), data); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"req": id,
						"msg": "add rule group failed: update rule config failed.",
					})
					return
				}

				c.JSON(http.StatusBadRequest, gin.H{
					"req": id,
					"msg": "remove rule group success.",
				})
				return
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"req": id,
			"msg": "remove rule group failed: group name not exists.",
		})
	}
}
