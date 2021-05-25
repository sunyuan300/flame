package scrape

import (
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/pkg/relabel"
	"net/url"
)

type BlackboxScrape struct {
	JobName string `from:"job_ame"`
	// http_2xx or tcp_connect or icmp
	Module         string `from:"params"`
	ScrapeInterval string `from:"interval"`
	ScrapeTimeout  string `from:"timeout"`
	// default /probe
	MetricsPath    string `from:"metrics_ath"`
	BlackboxTarget string `form:"target"`
	// 它至少应该包含psa和exporter类型标注
	Labels map[string]string `from:"labels"`
}

func (bs *BlackboxScrape) Marshal() (*config.ScrapeConfig, error) {
	interval, err := model.ParseDuration(bs.ScrapeInterval)
	if err != nil {
		return nil, err
	}
	timeout, err := model.ParseDuration(bs.ScrapeTimeout)
	if err != nil {
		return nil, err
	}
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
		ScrapeInterval: interval,
		ScrapeTimeout:  timeout,
		MetricsPath:    bs.MetricsPath,
		RelabelConfigs: relabelConfigs,
	}, nil
}
