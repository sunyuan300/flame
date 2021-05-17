package v1

import (
	_ "flame/docs"

	"flame/pkg/middle"
	gs "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func Group(f *Flame) {
	f.Web.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
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
