package prometheus_test

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/prometheus/pkg/rulefmt"
	"github.com/stretchr/testify/assert"

	"github.com/slok/sloth/internal/alert"
	"github.com/slok/sloth/internal/prometheus"
)

func getAlertGroup() alert.MWMBAlertGroup {
	return alert.MWMBAlertGroup{
		PageQuick: alert.MWMBAlert{
			ShortWindow: 5 * time.Minute,
			LongWindow:  1 * time.Hour,
		},
		PageSlow: alert.MWMBAlert{
			ShortWindow: 30 * time.Minute,
			LongWindow:  6 * time.Hour,
		},
		TicketQuick: alert.MWMBAlert{
			ShortWindow: 2 * time.Hour,
			LongWindow:  1 * 24 * time.Hour,
		},
		TicketSlow: alert.MWMBAlert{
			ShortWindow: 6 * time.Hour,
			LongWindow:  3 * 24 * time.Hour,
		},
	}
}

func TestGenerateSLIRecordingRules(t *testing.T) {
	tests := map[string]struct {
		slo        prometheus.SLO
		alertGroup alert.MWMBAlertGroup
		expRules   []rulefmt.Rule
		expErr     bool
	}{
		"Having and SLO with invalid expression should fail.": {
			slo: prometheus.SLO{
				ID:         "test",
				Name:       "test-name",
				Service:    "test-svc",
				TimeWindow: 30 * 24 * time.Hour,
				SLI: prometheus.CustomSLI{
					ErrorQuery: `rate(my_metric[{{}.window}}]{error="true"})`,
					TotalQuery: `rate(my_metric[{{.window}}])`,
				},
				Labels: map[string]string{
					"kind": "test",
				},
			},
			alertGroup: getAlertGroup(),
			expErr:     true,
		},

		"Having and wrong variablein the expression should fail.": {
			slo: prometheus.SLO{
				ID:         "test",
				Name:       "test-name",
				Service:    "test-svc",
				TimeWindow: 30 * 24 * time.Hour,
				SLI: prometheus.CustomSLI{
					ErrorQuery: `rate(my_metric[{{.Window}}]{error="true"})`,
					TotalQuery: `rate(my_metric[{{.window}}])`,
				},
				Labels: map[string]string{
					"kind": "test",
				},
			},
			alertGroup: getAlertGroup(),
			expErr:     true,
		},

		"Having and SLO an its mwmb alerts should create the recording rules.": {
			slo: prometheus.SLO{
				ID:         "test",
				Name:       "test-name",
				Service:    "test-svc",
				TimeWindow: 30 * 24 * time.Hour,
				SLI: prometheus.CustomSLI{
					ErrorQuery: `rate(my_metric[{{.window}}]{error="true"})`,
					TotalQuery: `rate(my_metric[{{.window}}])`,
				},
				Labels: map[string]string{
					"kind": "test",
				},
			},
			alertGroup: getAlertGroup(),
			expRules: []rulefmt.Rule{
				{
					Record: "slo:sli_error:ratio_rate5m",
					Expr:   "(rate(my_metric[5m]{error=\"true\"}))\n/\n(rate(my_metric[5m]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "5m",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate30m",
					Expr:   "(rate(my_metric[30m]{error=\"true\"}))\n/\n(rate(my_metric[30m]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "30m",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate1h",
					Expr:   "(rate(my_metric[1h]{error=\"true\"}))\n/\n(rate(my_metric[1h]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "1h",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate2h",
					Expr:   "(rate(my_metric[2h]{error=\"true\"}))\n/\n(rate(my_metric[2h]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "2h",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate6h",
					Expr:   "(rate(my_metric[6h]{error=\"true\"}))\n/\n(rate(my_metric[6h]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "6h",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate1d",
					Expr:   "(rate(my_metric[1d]{error=\"true\"}))\n/\n(rate(my_metric[1d]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "1d",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate3d",
					Expr:   "(rate(my_metric[3d]{error=\"true\"}))\n/\n(rate(my_metric[3d]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "3d",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate30d",
					Expr:   "max(avg_over_time(slo:sli_error:ratio_rate5m{sloth_id=\"test\", sloth_service=\"test-svc\", sloth_slo=\"test-name\"}[30d])) without(sloth_window)",
					Labels: map[string]string{
						"sloth_window": "30d",
					},
				},
			},
		},

		"An SLO alert with duplicated time windows should appear once and sorted.": {
			slo: prometheus.SLO{
				ID:         "test",
				Name:       "test-name",
				Service:    "test-svc",
				TimeWindow: 30 * 24 * time.Hour,
				SLI: prometheus.CustomSLI{
					ErrorQuery: `rate(my_metric[{{.window}}]{error="true"})`,
					TotalQuery: `rate(my_metric[{{.window}}])`,
				},
				Labels: map[string]string{
					"kind": "test",
				},
			},
			alertGroup: alert.MWMBAlertGroup{
				PageQuick:   alert.MWMBAlert{ShortWindow: 3 * time.Hour, LongWindow: 2 * time.Hour},
				PageSlow:    alert.MWMBAlert{ShortWindow: 3 * time.Hour, LongWindow: 1 * time.Hour},
				TicketQuick: alert.MWMBAlert{ShortWindow: 1 * time.Hour, LongWindow: 2 * time.Hour},
				TicketSlow:  alert.MWMBAlert{ShortWindow: 2 * time.Hour, LongWindow: 1 * time.Hour},
			},
			expRules: []rulefmt.Rule{
				{
					Record: "slo:sli_error:ratio_rate1h",
					Expr:   "(rate(my_metric[1h]{error=\"true\"}))\n/\n(rate(my_metric[1h]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "1h",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate2h",
					Expr:   "(rate(my_metric[2h]{error=\"true\"}))\n/\n(rate(my_metric[2h]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "2h",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate3h",
					Expr:   "(rate(my_metric[3h]{error=\"true\"}))\n/\n(rate(my_metric[3h]))\n",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
						"sloth_window":  "3h",
					},
				},
				{
					Record: "slo:sli_error:ratio_rate30d",
					Expr:   "max(avg_over_time(slo:sli_error:ratio_rate3h{sloth_id=\"test\", sloth_service=\"test-svc\", sloth_slo=\"test-name\"}[30d])) without(sloth_window)",
					Labels: map[string]string{
						"sloth_window": "30d",
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotRules, err := prometheus.SLIRecordingRulesGenerator.GenerateSLIRecordingRules(context.TODO(), test.slo, test.alertGroup)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expRules, gotRules)
			}
		})
	}
}

func TestGenerateMetaRecordingRules(t *testing.T) {
	tests := map[string]struct {
		slo        prometheus.SLO
		alertGroup alert.MWMBAlertGroup
		expRules   []rulefmt.Rule
		expErr     bool
	}{
		"Having and SLO an its mwmb alerts should create the metadata recording rules.": {
			slo: prometheus.SLO{
				ID:         "test",
				Name:       "test-name",
				Service:    "test-svc",
				Objective:  99.9,
				TimeWindow: 30 * 24 * time.Hour,
				Labels: map[string]string{
					"kind": "test",
				},
			},
			alertGroup: getAlertGroup(),
			expRules: []rulefmt.Rule{
				{
					Record: "slo:objective:ratio",
					Expr:   "vector(0.9990000000000001)",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
					},
				},
				{
					Record: "slo:error_budget:ratio",
					Expr:   "vector(1-0.9990000000000001)",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
					},
				},
				{
					Record: "slo:time_period:days",
					Expr:   "vector(30)",
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
					},
				},
				{
					Record: "slo:current_burn_rate:ratio",
					Expr: `slo:sli_error:ratio_rate5m{sloth_id="test", sloth_service="test-svc", sloth_slo="test-name"}
/ on(sloth_id, sloth_slo, sloth_service) group_left
slo:error_budget:ratio{sloth_id="test", sloth_service="test-svc", sloth_slo="test-name"}
`,
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
					},
				},
				{
					Record: "slo:period_burn_rate:ratio",
					Expr: `slo:sli_error:ratio_rate30d{sloth_id="test", sloth_service="test-svc", sloth_slo="test-name"}
/ on(sloth_id, sloth_slo, sloth_service) group_left
slo:error_budget:ratio{sloth_id="test", sloth_service="test-svc", sloth_slo="test-name"}
`,
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
					},
				},
				{
					Record: "slo:period_error_budget_remaining:ratio",
					Expr:   `1 - slo:period_burn_rate:ratio{sloth_id="test", sloth_service="test-svc", sloth_slo="test-name"}`,
					Labels: map[string]string{
						"kind":          "test",
						"sloth_service": "test-svc",
						"sloth_slo":     "test-name",
						"sloth_id":      "test",
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotRules, err := prometheus.MetadataRecordingRulesGenerator.GenerateMetadataRecordingRules(context.TODO(), test.slo, test.alertGroup)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expRules, gotRules)
			}
		})
	}
}
