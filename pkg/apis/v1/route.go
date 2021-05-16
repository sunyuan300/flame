package v1

import (
	"flame/pkg/middle"
)

func Group(f *Flame) {
	flameGroup := f.Web.Group("/api", middle.ResId())
	ScrapeGroup := flameGroup.Group("/scrape")
	{
		ScrapeGroup.GET("/:job_name", getNodeScrapeHandler(f))
	}
	NodeScrapeGroup := flameGroup.Group("/node_scrape")
	{
		NodeScrapeGroup.POST("", addNodeScrapeHandler(f))
		NodeScrapeGroup.DELETE("/:job_name", removeNodeScrapeHandler(f))
		// ?labels[a]=b&labels[b]=c
		NodeScrapeGroup.GET("", listNodeScrapeHandler(f))
	}
	TargetGroup := flameGroup.Group("/scrape/:job_name/static_target")
	{
		//add & update & remove
		TargetGroup.POST("", UpdateTargetHandler(f))
	}
}
