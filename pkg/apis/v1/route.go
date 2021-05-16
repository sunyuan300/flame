package v1

import (
	"flame/pkg/middle"
)

func Group(f *Flame) {
	flameGroup := f.Web.Group("/api", middle.ResId())
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
}
