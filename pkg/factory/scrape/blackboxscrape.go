package scrape

import (
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/pkg/relabel"
	"net/url"
)

type BlackboxScrape struct {
	JobName string `from:"jobName"`
	// http_2xx or tcp_connect or icmp
	Module         string         `from:"params"`
	ScrapeInterval model.Duration `from:"interval"`
	ScrapeTimeout  model.Duration `from:"timeout"`
	// default /probe
	MetricsPath    string `from:"metricsPath"`
	BlackboxTarget string `form:"target"`
	// 它至少应该包含psa和exporter类型标注
	Labels map[string]string `from:"labels"`
}

func (bs *BlackboxScrape) Marshal() *config.ScrapeConfig {
	relabelConfigs := []*relabel.Config{
		{
			SourceLabels: model.LabelNames{"__address__"},
			TargetLabel:  "__param_target",
		}, {
			SourceLabels: model.LabelNames{"__param_target"},
			TargetLabel:  "instance",
		}, {
			TargetLabel: "__address__",
			Replacement: bs.BlackboxTarget,
		},
	}
	for k, v := range bs.Labels {
		relabelConfigs = append(relabelConfigs, &relabel.Config{
			TargetLabel: k,
			Replacement: v,
		})
	}
	return &config.ScrapeConfig{
		JobName:        bs.JobName,
		Params:         url.Values{"module": []string{bs.Module}},
		ScrapeInterval: bs.ScrapeInterval,
		ScrapeTimeout:  bs.ScrapeTimeout,
		MetricsPath:    bs.MetricsPath,
		//ServiceDiscoveryConfigs: discovery.Configs{
		//	discovery.StaticConfig{
		//		{
		//			Targets: []model.LabelSet{
		//				{model.AddressLabel: ss.Target},
		//			},
		//			Labels: ss.Labels,
		//		},
		//	},
		//},
		RelabelConfigs: relabelConfigs,
	}
}
