package factory

import (
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	"sync"
)

type PromConfigInstance struct {
	Config    *config.Config
	ScrapeMap map[string]int
	LabelsMap map[string]map[string][]string
	Lock      sync.Mutex
}

type RulesConfigInstance struct {
	RuleGroups     *rulefmt.RuleGroups
	AllRulesGroups map[string]*rulefmt.RuleGroups
	Lock           sync.Mutex
}

func (p *PromConfigInstance) UpdateScrapeCache() {
	scrapeMap := make(map[string]int, len(p.Config.ScrapeConfigs))
	labelsMap := make(map[string]map[string][]string, 1)
	p.ScrapeMap = make(map[string]int, len(p.Config.ScrapeConfigs))
	for _, s := range p.Config.ScrapeConfigs {
		for _, v := range s.RelabelConfigs {
			if v.Replacement != "" {
				labelsMap[v.TargetLabel] = make(map[string][]string, 1)
			}
		}
	}
	for index, s := range p.Config.ScrapeConfigs {
		scrapeMap[s.JobName] = index
		for _, v := range s.RelabelConfigs {
			if v.Replacement != "" {
				labelsMap[v.TargetLabel][v.Replacement] = append(labelsMap[v.TargetLabel][v.Replacement], s.JobName)
			}
		}
	}
	p.ScrapeMap = scrapeMap
	p.LabelsMap = labelsMap
	p.Lock.Unlock()

}

func (p *PromConfigInstance) ExistsJobName(jobName string) bool {
	_, ok := p.ScrapeMap[jobName]
	if ok {
		// exists same name job
		return true
	} else {
		// not exists same name job
		return false
	}
}

func (r *RulesConfigInstance) ExistsRuleFileName(fileName string) bool {
	_, ok := r.AllRulesGroups[fileName]
	if ok {
		// exists same name rule file
		return true
	} else {
		// not exists same name rule file
		return false
	}
}
