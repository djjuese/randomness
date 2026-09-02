package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/Trisia/randomness"
	"github.com/Trisia/randomness/detect"
)

// ReportFormatter 报告格式化接口
type ReportFormatter interface {
	// FormatTestReport 格式化检测报告
	// - results: 检测结果
	// - w: 输出流
	// - 返回值: 错误信息
	FormatTestReport(results []*R, w io.Writer) error
	// FormatAnalysisReport 格式化分析报告
	// - results: 检测结果
	// - w: 输出流
	// - 返回值: 错误信息
	FormatAnalysisReport(results []AnalysisResult, w io.Writer) error
}

// ReportCollector 统一数据收集器
type ReportCollector struct {
	results      []*R       // 检测结果
	mu           sync.Mutex // 互斥锁
	format       string     // 输出格式 (csv/json)
	reportPath   string     // 报告输出路径
	analysisPath string     // 分析报告输出路径
}

// NewReportCollector 创建新的数据收集器
// - format: 输出格式 (csv/json/xml)
// - reportPath: 报告输出路径
// - analysisPath: 分析报告输出路径
func NewReportCollector(format, reportPath, analysisPath string) *ReportCollector {
	return &ReportCollector{
		results:      make([]*R, 0),
		format:       format,
		reportPath:   reportPath,
		analysisPath: analysisPath,
	}
}

// AddResult 添加检测结果
func (c *ReportCollector) AddResult(result *R) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, result)
}

// GetResults 获取所有检测结果
func (c *ReportCollector) GetResults() []*R {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.results
}

// GenerateReports 生成检测报告和分析报告
func (c *ReportCollector) GenerateReports() error {
	results := c.GetResults()

	// 生成检测报告
	if c.reportPath != "" && len(results) > 0 {
		if err := c.generateTestReport(results, c.reportPath); err != nil {
			return fmt.Errorf("生成检测报告失败: %v", err)
		}
	}

	// 生成分析报告
	if c.analysisPath != "" && len(results) > 0 {
		if err := c.generateAnalysisReport(results, c.analysisPath); err != nil {
			return fmt.Errorf("生成分析报告失败: %v", err)
		}
	}

	return nil
}

// generateTestReport 生成检测报告
func (c *ReportCollector) generateTestReport(results []*R, outputPath string) error {
	// 创建输出文件
	_ = os.MkdirAll(filepath.Dir(outputPath), os.FileMode(0600))
	file, err := os.OpenFile(outputPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.FileMode(0600))
	if err != nil {
		return err
	}
	defer file.Close()

	formatter := getFormatter(c.format)
	return formatter.FormatTestReport(results, file)
}

// generateAnalysisReport 生成分析报告
func (c *ReportCollector) generateAnalysisReport(results []*R, outputPath string) error {
	analysisResults, err := analyzeResults(results)
	if err != nil {
		return err
	}

	// 创建输出文件
	_ = os.MkdirAll(filepath.Dir(outputPath), os.FileMode(0600))
	file, err := os.OpenFile(outputPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.FileMode(0600))
	if err != nil {
		return err
	}
	defer file.Close()

	formatter := getFormatter(c.format)
	return formatter.FormatAnalysisReport(analysisResults, file)
}

// analyzeResults applies GM/T 0005-2021 clauses 6.1 through 6.4. P_value is
// used only for the sample pass count. The complete set of Q_values is divided
// into ten equal intervals and tested for uniformity; alpha_T applies to the
// resulting P_T, never to an individual sample's Q_value.
func analyzeResults(results []*R) ([]AnalysisResult, error) {
	if len(results) == 0 {
		return nil, nil
	}

	total := len(results)
	required := detect.Threshold(total)
	sampleCountPassed := total == 1000
	items := make([]AnalysisResult, 0, len(results[0].TestItems))

	for itemIndex, first := range results[0].TestItems {
		passCount := 0
		qValues := make([]float64, 0, total)
		bins := [10]int{}

		for _, result := range results {
			if len(result.TestItems) != len(results[0].TestItems) ||
				result.TestItems[itemIndex].TestName != first.TestName {
				return nil, fmt.Errorf("inconsistent test item layout in %s", result.Name)
			}
			item := result.TestItems[itemIndex]
			if math.IsNaN(item.PValue) || math.IsInf(item.PValue, 0) || item.PValue < 0 || item.PValue > 1 ||
				math.IsNaN(item.QValue) || math.IsInf(item.QValue, 0) || item.QValue < 0 || item.QValue > 1 {
				return nil, fmt.Errorf("invalid P/Q value for %s in %s", first.TestName, result.Name)
			}
			if item.PValue >= randomness.Alpha {
				passCount++
			}
			qValues = append(qValues, item.QValue)
			bin := int(item.QValue * 10)
			if bin == 10 { // Q=1 belongs to the closed final interval [0.9, 1].
				bin = 9
			}
			bins[bin]++
		}

		expected := float64(total) / 10.0
		chiSquare := 0.0
		for _, count := range bins {
			delta := float64(count) - expected
			chiSquare += delta * delta / expected
		}
		uniformityP := detect.ThresholdQ(qValues)
		passCountPassed := passCount >= required
		uniformityPassed := uniformityP >= randomness.AlphaT

		items = append(items, AnalysisResult{
			TestName:          first.TestName,
			PassCount:         passCount,
			TotalCount:        total,
			RequiredPassCount: required,
			PassRate:          float64(passCount) / float64(total),
			PassCountPassed:   passCountPassed,
			QBins:             bins,
			QChiSquare:        chiSquare,
			QUniformityP:      uniformityP,
			QUniformityPassed: uniformityPassed,
			SampleCountPassed: sampleCountPassed,
			IsPassed:          sampleCountPassed && passCountPassed && uniformityPassed,
		})
	}
	return items, nil
}

// JSONFormatter JSON格式输出
type JSONFormatter struct{}

func (f *JSONFormatter) FormatTestReport(results []*R, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func (f *JSONFormatter) FormatAnalysisReport(results []AnalysisResult, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// CSVFormatter CSV格式输出
type CSVFormatter struct{}

func (f *CSVFormatter) FormatTestReport(results []*R, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 写入CSV表头
	headers := []string{"文件名"}
	if len(results) > 0 {
		for _, item := range results[0].TestItems {
			headers = append(headers, item.TestName+" P值", item.TestName+" Q值")
		}
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// 写入数据
	for _, result := range results {
		record := []string{result.Name}
		for _, item := range result.TestItems {
			record = append(record,
				fmt.Sprintf("%.6f", item.PValue),
				fmt.Sprintf("%.6f", item.QValue))
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func (f *CSVFormatter) FormatAnalysisReport(results []AnalysisResult, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 写入CSV表头
	headers := []string{
		"检测项目（含参数）", "P通过数", "样本数", "P通过数门限", "P通过率", "P通过率是否通过",
		"Q值十区间计数", "Q均匀性卡方值", "Q均匀性P_T", "Q均匀性门限", "Q均匀性是否通过",
		"样本数是否为1000", "最终判定",
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// 写入数据
	for _, result := range results {
		passRateStr := fmt.Sprintf("%.4f", result.PassRate)
		record := []string{
			result.TestName,
			strconv.Itoa(result.PassCount),
			strconv.Itoa(result.TotalCount),
			strconv.Itoa(result.RequiredPassCount),
			passRateStr,
			yesNo(result.PassCountPassed),
			fmt.Sprint(result.QBins),
			fmt.Sprintf("%.12g", result.QChiSquare),
			fmt.Sprintf("%.12g", result.QUniformityP),
			fmt.Sprintf("%.4g", randomness.AlphaT),
			yesNo(result.QUniformityPassed),
			yesNo(result.SampleCountPassed),
			yesNo(result.IsPassed),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func yesNo(value bool) string {
	if value {
		return "是"
	}
	return "否"
}

// XMLFormatter XML格式输出
type XMLFormatter struct{}

func (f *XMLFormatter) FormatTestReport(results []*R, w io.Writer) error {
	_, err := w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<RandomnessTestReport>\n"))
	if err != nil {
		return err
	}

	for _, result := range results {
		_, err = w.Write([]byte(fmt.Sprintf("  <File name=\"%s\">\n", result.Name)))
		if err != nil {
			return err
		}

		for _, item := range result.TestItems {
			_, err = w.Write([]byte(fmt.Sprintf("    <Test name=\"%s\" p=\"%.6f\" q=\"%.6f\"/>\n",
				item.TestName, item.PValue, item.QValue)))
			if err != nil {
				return err
			}
		}

		_, err = w.Write([]byte("  </File>\n"))
		if err != nil {
			return err
		}
	}

	_, err = w.Write([]byte("</RandomnessTestReport>\n"))
	return err
}

func (f *XMLFormatter) FormatAnalysisReport(results []AnalysisResult, w io.Writer) error {
	_, err := w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<AnalysisReport>\n"))
	if err != nil {
		return err
	}

	for _, result := range results {
		isPassedStr := "true"
		if !result.IsPassed {
			isPassedStr = "false"
		}

		_, err = w.Write([]byte(fmt.Sprintf("  <Test name=\"%s\" passCount=\"%d\" totalCount=\"%d\" requiredPassCount=\"%d\" passRate=\"%.4f\" qBins=\"%v\" qChiSquare=\"%.12g\" qUniformityPT=\"%.12g\" qUniformityThreshold=\"%.4g\" passCountPassed=\"%t\" qUniformityPassed=\"%t\" sampleCountPassed=\"%t\" isPassed=\"%s\"/>\n",
			result.TestName, result.PassCount, result.TotalCount, result.RequiredPassCount,
			result.PassRate, result.QBins, result.QChiSquare, result.QUniformityP,
			randomness.AlphaT, result.PassCountPassed, result.QUniformityPassed,
			result.SampleCountPassed, isPassedStr)))
		if err != nil {
			return err
		}
	}

	_, err = w.Write([]byte("</AnalysisReport>\n"))
	return err
}

// getFormatter 根据格式名称获取格式化器
func getFormatter(format string) ReportFormatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "xml":
		return &XMLFormatter{}
	case "csv":
		fallthrough
	default:
		return &CSVFormatter{}
	}
}
