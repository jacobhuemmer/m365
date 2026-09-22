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
}

type ChangeStore interface {
	Delta(context.Context, DeltaQuery) (domain.MailDeltaPage, error)
}

type WatchStateStore interface {
	Load(account, folder string) (domain.MailWatchState, error)
	Save(account, folder string, state domain.MailWatchState) error
}

type EventSink interface {
	Emit(context.Context, domain.MailWatchEvent) error
}
