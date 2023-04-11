package lint

import "fmt"

// inserts and adds all panels Id for a given panel and all of its subpanels.
// recursive, could be faster
func map_panel_ids(ids *map[int]int, p *Panel) {
	if _, ok := (*ids)[p.Id]; !ok {
		(*ids)[p.Id] = 1
	} else {
		(*ids)[p.Id] += 1
	}
	// iterate over sub panels within panel
	for _, p2 := range p.Panels {
		map_panel_ids(ids, &p2)
	}
}

// linting rule which enumerates all panels on the dashboard and checks whether
// the input panel id appears more than once in the dashboard.
// assumes panels don't have circular references.
func NewPanelUniqueIdFunc() *PanelRuleFunc {

	return &PanelRuleFunc{
		name:        "panel-unique-id",
		description: "Checks that each panel has a unique ID",
		fn: func(d Dashboard, p Panel) Result {
			ids := make(map[int]int)
			// iterate over base panels
			for _, inner := range d.Panels {
				map_panel_ids(&ids, &inner)
			}
			// iterate panels in each row
			for _, row := range d.Rows {
				for _, inner := range row.Panels {
					map_panel_ids(&ids, &inner)
				}
			}
			// see if num of panels with id > 1
			if val, ok := ids[p.Id]; ok {
				if val > 1 {
					// calculate suggested assignment
					// get find max key in map
					// iterate from 1..max
					// for each val, check if
					max := -1
					for k, _ := range ids {
						if k > max {
							max = k
						}
					}
					candidates := make([]int, 1)
					for i := 1; i < max; i++ {
						if _, ok := ids[i]; !ok {
							candidates = append(candidates, i)
						}
					}
					return Result{
						Severity: Error,
						Message:  fmt.Sprintf("Dashboard '%s', panel with id '%v' has panels with %d duplicate id(s). Candidate ids: %v (max: %d)", d.Title, p.Id, val, candidates, max),
					}
				}
			} else {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s', panel with id '%v' has panels with missing Id?? Please contact the lint rule developer", d.Title, p.Id),
				}
			}
			return ResultSuccess
		},
	}
}
