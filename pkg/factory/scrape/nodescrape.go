package scrape

import (
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/pkg/relabel"
)

type NodeScrape struct {
	JobName        string `json:"job_name"`
	ScrapeInterval string `json:"scrape_interval"`
	ScrapeTimeout  string `json:"scrape_timeout"`
	// default /metrics
	MetricsPath string `json:"metrics_path"`
	// 它至少应该包含psa和exporter类型标注
	Labels map[string]string `json:"labels"`
}

func (ns *NodeScrape) Marshal() (*config.ScrapeConfig, error) {
	var interval model.Duration
	var timeout model.Duration
	var err error

	if len(ns.ScrapeInterval) != 0 {
		interval, err = model.ParseDuration(ns.ScrapeInterval)
		if err != nil {
			return nil, err
		}
	}

	if len(ns.ScrapeTimeout) != 0 {
		timeout, err = model.ParseDuration(ns.ScrapeTimeout)
		if err != nil {
			return nil, err
		}
	}

	var relabelConfigs []*relabel.Config
	for k, v := range ns.Labels {
		relabelConfigs = append(relabelConfigs, &relabel.Config{
			TargetLabel: k,
			Replacement: v,
		})
	}

	return &config.ScrapeConfig{
		JobName:        ns.JobName,
		ScrapeInterval: interval,
		ScrapeTimeout:  timeout,
		MetricsPath:    ns.MetricsPath,
		RelabelConfigs: relabelConfigs,
	}, nil
}
