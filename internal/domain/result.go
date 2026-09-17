package domain

const (
	DefaultMailTop      = 10
	DefaultTeamsTop     = 20
	DefaultCalendarTop  = 10
	DefaultCalendarsTop = 20
	DefaultFilesTop     = 20
	DefaultFindTop      = 10
	MaxFindTop          = 20
	MaxTop              = 50
)

type CommandResult struct {
	Stdout any
	Err    error
}

type ResultLimit struct {
	Limit    int
	Count    int
	NextPage string
}

func NormalizeTop(raw int, def int) (int, error) {
	if raw == 0 {
		return def, nil
	}
	if raw < 1 || raw > MaxTop {
		return 0, Usagef("top must be 1..%d, got %d", MaxTop, raw)
	}
	return raw, nil
}

func NormalizeFindTop(raw int) (int, error) {
	if raw == 0 {
		return DefaultFindTop, nil
	}
	if raw < 1 || raw > MaxFindTop {
		return 0, Usagef("top must be 1..%d, got %d", MaxFindTop, raw)
	}
	return raw, nil
}
