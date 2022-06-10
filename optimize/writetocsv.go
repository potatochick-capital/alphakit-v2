package optimize

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/thecolngroup/alphakit/broker"
	"github.com/thecolngroup/alphakit/perf"
	"github.com/thecolngroup/util"
)

// WriteStudyResultToCSV writes the results of a study to CSV.
// Study results are flattened to separate files with foreign keys added to link rows.
// Study ID is prefixed to the result filenames.
//
// - phasereport.csv: summary performance of each paramset for each phase
//
// - trialreport.csv: detailed performance of each trial (backtest) in a phase
//
// - trades.csv: trades completed for each trial
//
// - curve.csv: equity curve for each trial
func WriteStudyResultToCSV(path string, study *Study) error {

	phaseReports, trialReports, trades, curves := prepareStudyForCSV(study)

	prefix := study.ID

	out := filepath.Join(path, fmt.Sprintf("%s-phasereports.csv", prefix))
	if err := saveDataToCSV(out, phaseReports); err != nil {
		return err
	}

	out = filepath.Join(path, fmt.Sprintf("%s-trialreports.csv", prefix))
	if err := saveDataToCSV(out, trialReports); err != nil {
		return err
	}

	out = filepath.Join(path, fmt.Sprintf("%s-trades.csv", prefix))
	if err := saveDataToCSV(out, trades); err != nil {
		return err
	}

	out = filepath.Join(path, fmt.Sprintf("%s-curves.csv", prefix))
	if err := saveDataToCSV(out, curves); err != nil {
		return err
	}

	return nil
}

// phaseReport is a wrapper on optimize.PhaseReport that adds a PK for saving to CSV.
type phaseReport struct {
	StudyID string      `csv:"study_id"`
	Report  PhaseReport `csv:"phasereport_,inline"`
}

// trialReport is a wrapper on perf.PerformanceReport that adds a compound key for saving to CSV.
type trialReport struct {
	StudyID       string                 `csv:"study_id"`
	PhaseReportID string                 `csv:"phasereport_id"`
	Backtest      perf.PerformanceReport `csv:"backtest_,inline"`
}

type tradeDetailRow struct {
	StudyID       string       `csv:"study_id"`
	PhaseReportID string       `csv:"phasereport_id"`
	BacktestID    string       `csv:"backtest_id"`
	Trade         broker.Trade `csv:"trade_,inline"`
}

type curveDetailRow struct {
	StudyID       string    `csv:"study_id"`
	PhaseReportID string    `csv:"phasereport_id"`
	BacktestID    string    `csv:"backtest_id"`
	Time          time.Time `csv:"time"`
	Amount        float64   `csv:"amount"`
}

// prepareStudyForCSV returns data that is ready for saving to CSV.
func prepareStudyForCSV(study *Study) ([]phaseReport, []trialReport, []tradeDetailRow, []curveDetailRow) {

	var phaseReports []phaseReport
	var trialReports []trialReport
	var tradeRows []tradeDetailRow
	var curveRows []curveDetailRow

	flattenResults := func(results map[ParamSetID]PhaseReport) {
		for k := range results {
			report := results[k]
			phaseReports = append(phaseReports, phaseReport{
				StudyID: study.ID,
				Report:  report,
			})
			for _, trial := range report.Trials {
				trialReports = append(trialReports, trialReport{
					StudyID:       study.ID,
					PhaseReportID: report.ID,
					Backtest:      trial,
				})

				if trial.PortfolioReport == nil || trial.TradeReport == nil {
					continue
				}

				for _, trade := range trial.TradeReport.Trades {
					tradeRows = append(tradeRows, tradeDetailRow{
						StudyID:       study.ID,
						PhaseReportID: report.ID,
						BacktestID:    trial.ID,
						Trade:         trade,
					})
				}
				curve := trial.PortfolioReport.EquityCurve
				sortedKeys := curve.SortKeys()
				for _, k := range sortedKeys {
					curveRows = append(curveRows, curveDetailRow{
						StudyID:       study.ID,
						PhaseReportID: report.ID,
						BacktestID:    trial.ID,
						Time:          k.Time(),
						Amount:        curve[k].InexactFloat64(),
					})
				}
			}
		}
	}

	flattenResults(study.TrainingResults)
	flattenResults(study.ValidationResults)

	return phaseReports, trialReports, tradeRows, curveRows
}

func saveDataToCSV(filename string, data interface{}) error {

	encMap := func(m map[string]any) ([]byte, error) {
		return []byte(fmt.Sprint(m)), nil
	}

	encParamMap := func(m ParamMap) ([]byte, error) {
		return []byte(fmt.Sprint(m)), nil
	}

	return util.WriteToCSV(filename, data, encMap, encParamMap)
}
