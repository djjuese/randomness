package main

import (
	"testing"
)

func makeAnalysisResults(passCount int, q func(int) float64) []*R {
	results := make([]*R, 1000)
	for i := range results {
		p := 0.5
		if i >= passCount {
			p = 0.001
		}
		results[i] = &R{
			Name: "sample",
			TestItems: []TestItem{{
				PValue: p, QValue: q(i), TestName: "test",
			}},
		}
	}
	return results
}

func TestAnalyzeResultsGMT0005Pass(t *testing.T) {
	results, err := analyzeResults(makeAnalysisResults(981, func(i int) float64 {
		return (float64(i%10) + 0.5) / 10.0
	}))
	if err != nil {
		t.Fatal(err)
	}
	got := results[0]
	if got.PassCount != 981 || got.RequiredPassCount != 981 || !got.PassCountPassed {
		t.Fatalf("unexpected P pass result: %+v", got)
	}
	if got.QUniformityP != 1 || !got.QUniformityPassed || !got.IsPassed {
		t.Fatalf("unexpected Q uniformity result: %+v", got)
	}
}

func TestAnalyzeResultsChecksQDistributionNotIndividualThreshold(t *testing.T) {
	results := makeAnalysisResults(1000, func(i int) float64 {
		if i == 0 {
			return 0.00001
		}
		return 0.5
	})
	got, err := analyzeResults(results)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].PassCount != 1000 {
		t.Fatalf("an individual Q value changed the P pass count: %+v", got[0])
	}
	if got[0].QUniformityPassed || got[0].IsPassed {
		t.Fatalf("non-uniform Q distribution passed: %+v", got[0])
	}
}

func TestAnalyzeResultsRejects980PassingSamples(t *testing.T) {
	got, err := analyzeResults(makeAnalysisResults(980, func(i int) float64 {
		return (float64(i%10) + 0.5) / 10.0
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got[0].PassCountPassed || got[0].IsPassed {
		t.Fatalf("980 passing samples were accepted: %+v", got[0])
	}
}

func TestAnalyzeResultsRequires1000Samples(t *testing.T) {
	input := makeAnalysisResults(1000, func(i int) float64 {
		return (float64(i%10) + 0.5) / 10.0
	})[:999]
	got, err := analyzeResults(input)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].SampleCountPassed || got[0].IsPassed {
		t.Fatalf("non-standard sample count was accepted: %+v", got[0])
	}
}
