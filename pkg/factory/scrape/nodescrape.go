package scrape

import (
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/pkg/relabel"
	"time"
)

type NodeScrape struct {
	JobName        string `json:"job_name"`
	ScrapeInterval int    `json:"scrape_interval"`
	ScrapeTimeout  int    `json:"scrape_timeout"`
	// default /metrics
	MetricsPath string `json:"metrics_path"`
	// 它至少应该包含psa和exporter类型标注
	Labels map[string]string `json:"labels"`
}

func (ns *NodeScrape) Marshal() *config.ScrapeConfig {
	var relabelConfigs []*relabel.Config
	for k, v := range ns.Labels {
		relabelConfigs = append(relabelConfigs, &relabel.Config{
			TargetLabel: k,
			Replacement: v,
		})
	}
	return &config.ScrapeConfig{
		JobName:        ns.JobName,
		ScrapeInterval: model.Duration(time.Duration(ns.ScrapeInterval) * time.Second),
		ScrapeTimeout:  model.Duration(time.Duration(ns.ScrapeTimeout) * time.Second),
		MetricsPath:    ns.MetricsPath,
		RelabelConfigs: relabelConfigs,
	}
}
