package mailclassifier

import (
	"context"

	"github.com/masonhuemmer/m365/internal/domain"
)

type Fake struct {
	Result domain.ResponseClassification
	Err    error
}

func (f *Fake) Classify(_ context.Context, _ domain.ClassificationInput) (domain.ResponseClassification, error) {
	if f != nil && f.Err != nil {
		return domain.ResponseClassification{}, f.Err
	}
	if f != nil && f.Result.Status != "" {
		return f.Result, nil
	}
	confidence := 0.85
	return domain.ResponseClassification{
		Status:            domain.ResponseWaitingOnTarget,
		TargetProbability: 0.9,
		Probabilities: map[domain.ResponseStatus]float64{
			domain.ResponseWaitingOnTarget:    0.9,
			domain.ResponseWaitingOnOther:     0.04,
			domain.ResponseNoResponseExpected: 0.03,
			domain.ResponseUnclear:            0.03,
		},
		Confidence: &confidence,
		Model:      "fake-response-classifier-v1",
	}, nil
}
