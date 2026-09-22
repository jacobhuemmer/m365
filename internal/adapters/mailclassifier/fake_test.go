package mailclassifier

import (
	"context"
	"reflect"
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestFakeReturnsStableFourStateClassification(t *testing.T) {
	classifier := &Fake{}
	input := domain.ClassificationInput{Target: domain.ResponseTarget{Addresses: []string{"user@example.com"}}}

	first, err := classifier.Classify(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := classifier.Classify(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("fake result changed: first=%+v second=%+v", first, second)
	}
	if first.Status != domain.ResponseWaitingOnTarget || first.Model != "fake-response-classifier-v1" {
		t.Fatalf("fake result = %+v", first)
	}
	for _, status := range []domain.ResponseStatus{
		domain.ResponseWaitingOnTarget,
		domain.ResponseWaitingOnOther,
		domain.ResponseNoResponseExpected,
		domain.ResponseUnclear,
	} {
		if _, ok := first.Probabilities[status]; !ok {
			t.Fatalf("fake probabilities missing %q: %+v", status, first.Probabilities)
		}
	}
	if first.Confidence == nil || first.TargetProbability != first.Probabilities[domain.ResponseWaitingOnTarget] {
		t.Fatalf("fake confidence/probability = %+v", first)
	}
}

func TestFakeCanReturnConfiguredResultAndError(t *testing.T) {
	want := domain.ResponseClassification{Status: domain.ResponseUnclear, Model: "configured"}
	classifier := &Fake{Result: want}
	got, err := classifier.Classify(context.Background(), domain.ClassificationInput{})
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("configured fake = %+v, %v", got, err)
	}

	classifier.Err = context.DeadlineExceeded
	if _, err := classifier.Classify(context.Background(), domain.ClassificationInput{}); err == nil {
		t.Fatal("expected configured error")
	}
}
