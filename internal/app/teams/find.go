package teams

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/masonhuemmer/m365/internal/app/auth"
	"github.com/masonhuemmer/m365/internal/domain"
)

const (
	IntentPerson = "person"
	IntentGroup  = "group"
	ScanPageSize = 50
	ScanMaxPages = 10
	// SelfChatID is Teams Notes (chat with yourself). Graph list omits it.
	SelfChatID = "48:notes"
)

type FindQuery struct {
	Raw    string
	Intent string
	Needle string
}

type FindResult struct {
	Query      string        `json:"query"`
	Intent     string        `json:"intent"`
	Limit      int           `json:"limit"`
	Count      int           `json:"count"`
	Incomplete bool          `json:"incomplete"`
	Items      []domain.Chat `json:"items"`
}

var groupWithRe = regexp.MustCompile(`(?i)^(?:the )?group with (.+)$`)

func ParseQuery(raw string, groupFlag bool) (FindQuery, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return FindQuery{}, domain.Usage("query is required")
	}
	q := FindQuery{Raw: raw, Needle: raw, Intent: IntentPerson}
	if groupFlag {
		q.Intent = IntentGroup
		return q, nil
	}
	if m := groupWithRe.FindStringSubmatch(raw); m != nil {
		q.Intent = IntentGroup
		q.Needle = strings.TrimSpace(m[1])
	}
	return q, nil
}

func Find(ctx context.Context, st Store, sess domain.Session, q FindQuery, top int) (FindResult, error) {
	out := FindResult{Query: q.Raw, Intent: q.Intent, Items: []domain.Chat{}}
	if err := auth.Require(sess, false, true); err != nil {
		return out, err
	}
	n, err := domain.NormalizeFindTop(top)
	if err != nil {
		return out, err
	}
	out.Limit = n
	type scored struct {
		c domain.Chat
		s int
	}
	var hits []scored
	if c, s, ok := selfChat(sess, q); ok {
		hits = append(hits, scored{c, s})
		if s >= 2 {
			out.Items = []domain.Chat{c}
			out.Count = 1
			return out, nil
		}
	}
	page := ""
	incomplete := false
	for i := 0; i < ScanMaxPages; i++ {
		p, err := st.ListChats(ctx, ScanPageSize, page)
		if err != nil {
			return out, err
		}
		for _, c := range p.Items {
			if s, ok := scoreChat(c, q, sess.Account); ok {
				hits = append(hits, scored{c, s})
			}
		}
		if len(hits) >= n && n > 1 {
			break
		}
		if p.NextPage == nil || *p.NextPage == "" {
			break
		}
		page = *p.NextPage
		if i == ScanMaxPages-1 {
			incomplete = true
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].s > hits[j].s })
	if len(hits) > n {
		hits = hits[:n]
	}
	for _, h := range hits {
		out.Items = append(out.Items, h.c)
	}
	out.Count = len(out.Items)
	out.Incomplete = incomplete && out.Count <= 1
	return out, nil
}

func scoreChat(c domain.Chat, q FindQuery, self string) (int, bool) {
	needle := strings.ToLower(strings.TrimSpace(q.Needle))
	is1 := strings.EqualFold(c.Type, "oneOnOne")
	isG := strings.EqualFold(c.Type, "group")
	if q.Intent == IntentPerson {
		if !is1 {
			return 0, false
		}
		best := 0
		for _, p := range c.Members {
			if self != "" && strings.EqualFold(p.Address, self) {
				continue
			}
			if s := memberScore(p, needle); s > best {
				best = s
			}
		}
		return best, best > 0
	}
	if !isG {
		return 0, false
	}
	best := 0
	if strings.Contains(strings.ToLower(c.Topic), needle) {
		best = 3
	}
	for _, p := range c.Members {
		if s := memberScore(p, needle); s > best {
			best = s
		}
	}
	return best, best > 0
}

func selfChat(sess domain.Session, q FindQuery) (domain.Chat, int, bool) {
	if q.Intent != IntentPerson || strings.TrimSpace(sess.Account) == "" {
		return domain.Chat{}, 0, false
	}
	local, _, _ := strings.Cut(sess.Account, "@")
	p := domain.Person{
		Name:    strings.ReplaceAll(strings.ReplaceAll(local, ".", " "), "_", " "),
		Address: sess.Account,
	}
	s := memberScore(p, strings.ToLower(strings.TrimSpace(q.Needle)))
	if s == 0 {
		return domain.Chat{}, 0, false
	}
	return domain.Chat{
		ID:      SelfChatID,
		Type:    "oneOnOne",
		Members: []domain.Person{{Name: p.Name, Address: sess.Account}},
	}, s, true
}

func memberScore(p domain.Person, q string) int {
	name := strings.ToLower(strings.TrimSpace(p.Name))
	addr := strings.ToLower(strings.TrimSpace(p.Address))
	if name == q || addr == q {
		return 3
	}
	if strings.HasPrefix(name, q) || strings.HasPrefix(addr, q) {
		return 2
	}
	if strings.Contains(name, q) || strings.Contains(addr, q) {
		return 1
	}
	return 0
}
