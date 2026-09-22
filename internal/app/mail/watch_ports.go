package mail

import (
	"context"

	"github.com/masonhuemmer/m365/internal/domain"
)

type DeltaQuery struct {
	Folder string
	Token  string
}

type WatchQuery struct {
	Folder          string
	IncludeExisting bool
	Classify        bool
	TargetAddresses []string
	TargetNames     []string
}

type WatchConfig struct {
	ClassificationEnabled bool
	ActionableThreshold   float64
}

type ChangeStore interface {
	Delta(context.Context, DeltaQuery) (domain.MailDeltaPage, error)
}

type ThreadStore interface {
	LatestThread(context.Context, string, int) (domain.MailThread, error)
}

type ResponseClassifier interface {
	Classify(context.Context, domain.ClassificationInput) (domain.ResponseClassification, error)
}

type WatchStateStore interface {
	Load(account, folder string) (domain.MailWatchState, error)
	Save(account, folder string, state domain.MailWatchState) error
}

type EventSink interface {
	Emit(context.Context, domain.MailWatchEvent) error
}
