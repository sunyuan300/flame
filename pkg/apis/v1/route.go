package v1

import (
	_ "flame/docs"

	"flame/pkg/middle"
	gs "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func Group(f *Flame) {
	f.Web.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
	flameGroup := f.Web.Group("/api", middle.ReqId())

	// scrape apis
	ScrapeGroup := flameGroup.Group("/scrape")
	{
		ScrapeGroup.DELETE("/:job_name", removeScrapeHandler(f))
		// ?labels[a]=b&labels[b]=c
		ScrapeGroup.GET("", listScrapeHandler(f))
		ScrapeGroup.GET("/:job_name", getScrapeHandler(f))
	}
	NodeScrapeGroup := flameGroup.Group("/node_scrape")
	{
		NodeScrapeGroup.POST("", addNodeScrapeHandler(f))
		NodeScrapeGroup.POST("/:job_name", updateNodeScrapeHandler(f))
	}
	TargetGroup := flameGroup.Group("/scrape/:job_name/static_target")
	{
		//add & update & remove
		TargetGroup.POST("", UpdateTargetHandler(f))
	}

	// rule apis
	RulesFile := flameGroup.Group("/rule_files")
	{
		RulesFile.GET("", listRuleFileHandler(f))
		// rule file ends with .yaml or .yml
		RulesFile.POST("", addRuleFileHandler(f))
		RulesFile.DELETE("/:file_name", removeRuleFileHandler(f))

		RulesGroup := RulesFile.Group("/:file_name/rule_groups")
		{
			RulesGroup.GET("", listRuleGroupHandler(f))
			RulesGroup.GET("/:group_name", getRuleGroupHandler(f))
			RulesGroup.POST("", addRuleGroupHandler(f))
			RulesGroup.DELETE("/:group_name", removeRuleGroupHandler(f))
			rules := RulesGroup.Group("/:group_name/rules")
			{
				rules.POST("", updateRulesHandler(f))
			}
		}
	}
}
