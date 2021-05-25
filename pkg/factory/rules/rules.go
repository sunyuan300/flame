package rules

import (
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	"gopkg.in/yaml.v3"
)

type Rule struct {
	//Record      string            `json:"record,omitempty"`
	Alert           string            `json:"alert,omitempty"`
	Expr            string            `json:"expr"`
	For             string            `json:"for,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Summary         string            `json:"summary,omitempty"`
	Description     string            `json:"description,omitempty"`
	UserDescription string            `json:"user_description,omitempty"`
}

type RuleFile struct {
	FileName string `json:"file_name"`
}

type RuleGroup struct {
	GroupName string `json:"group_name,omitempty"`
	Rules     []Rule `json:"rules,omitempty"`
}

func (rg *RuleGroup) Marshal() ([]rulefmt.RuleNode, error) {
	var ruleNodes []rulefmt.RuleNode
	for _, v := range rg.Rules {
		alert := yaml.Node{}
		expr := yaml.Node{}
		annotations := make(map[string]string, 2)
		alert.SetString(v.Alert)
		expr.SetString(v.Expr)
		duration, err := model.ParseDuration(v.For)
		if err != nil {
			return nil, err
		}
		annotations["summary"] = v.Summary
		annotations["description"] = v.Description
		annotations["user_description"] = v.UserDescription
		ruleNodes = append(ruleNodes, rulefmt.RuleNode{
			Alert:       alert,
			Expr:        expr,
			For:         duration,
			Labels:      v.Labels,
			Annotations: annotations,
		})

	}
	return ruleNodes, nil
}

func (r *Rule) Marshal() (*rulefmt.RuleNode, error) {
	alert := yaml.Node{}
	expr := yaml.Node{}
	annotations := make(map[string]string, 2)
	alert.SetString(r.Alert)
	expr.SetString(r.Expr)
	//var duration *model.Duration
	//err := duration.Set(r.For)
	duration, err := model.ParseDuration(r.For)
	if err != nil {
		return nil, err
	}
	annotations["summary"] = r.Summary
	annotations["description"] = r.Description
	annotations["user_description"] = r.UserDescription
	return &rulefmt.RuleNode{
		Alert:       alert,
		Expr:        expr,
		For:         duration,
		Labels:      r.Labels,
		Annotations: annotations,
	}, nil
}

func UnMarshal(ruleNodes []rulefmt.RuleNode) []Rule {
	var rs []Rule
	for _, v := range ruleNodes {
		rs = append(rs, Rule{
			Alert:           v.Alert.Value,
			Expr:            v.Expr.Value,
			For:             v.For.String(),
			Labels:          v.Labels,
			Summary:         v.Annotations["summary"],
			Description:     v.Annotations["description"],
			UserDescription: v.Annotations["user_description"],
		})
	}
	return rs
}
