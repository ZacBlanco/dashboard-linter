package lint

import (
	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

type Rule interface {
	Description() string
	Name() string
}

type IntegrationRule interface {
	Rule
	LintIntegration(*integrations.Integration) Result
}

type DashboardRule interface {
	Rule
	LintDashboard(*integrations.Integration, Dashboard) Result
}

type PanelRule interface {
	Rule
	LintPanel(*integrations.Integration, Dashboard, Panel) Result
}

type TargetRule interface {
	Rule
	LintTarget(*integrations.Integration, Dashboard, Panel, Target) Result
}

// RuleSet contains a list of linting rules, a list of integrations, and list of configurations
// which dictate the ultimate application of a rule to an integration based on the configs.
type RuleSet struct {
	integrationRules []IntegrationRule
	dashboardRules   []DashboardRule
	panelRules       []PanelRule
	targetRules      []TargetRule
	integrations     []*integrations.Integration
}

func (s *RuleSet) Integrations() []*integrations.Integration {
	return s.integrations
}

func (s *RuleSet) Lint() (*ResultSet, []error) {
	var errs []error
	resSet := &ResultSet{config: &Configuration{}}
	for _, i := range s.integrations {
		for _, ir := range s.integrationRules {
			resSet.AddResult(ResultContext{
				Result:      ir.LintIntegration(i),
				Integration: i,
				Rule:        ir,
			})
		}
		// Dashboards
		for _, d := range i.Dashboards {
			dash, err := NewDashboardFromGrafanaDashboard(d)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			for _, dr := range s.dashboardRules {
				resSet.AddResult(ResultContext{
					Result:      dr.LintDashboard(i, dash),
					Integration: i,
					Rule:        dr,
					Dashboard:   &dash,
				})
			}
			panels := dash.GetPanels()
			for p := range panels {
				for _, pr := range s.panelRules {
					resSet.AddResult(ResultContext{
						Result:      pr.LintPanel(i, dash, panels[p]),
						Integration: i,
						Rule:        pr,
						Dashboard:   &dash,
						Panel:       &panels[p],
					})
				}
				targets := panels[p].Targets
				for t := range targets {
					for _, tr := range s.targetRules {
						resSet.AddResult(ResultContext{
							Result:      tr.LintTarget(i, dash, panels[p], targets[t]),
							Integration: i,
							Rule:        tr,
							Dashboard:   &dash,
							Panel:       &panels[p],
							Target:      &targets[t],
						})
					}
				}
			}
		}
		// One day rules, and alerts
	}
	return resSet, errs
}

func (s *RuleSet) AddIntegration(i *integrations.Integration) {
	s.integrations = append(s.integrations, i)
}

func NewRuleSet() RuleSet {
	return RuleSet{
		integrationRules: NewIntegrationRules(),
	}
}

func NewIntegrationRules() []IntegrationRule {
	return []IntegrationRule{
		NewMetaLogoURLRule(),
		NewMetaSupportedPlatformsRule(),
	}
}