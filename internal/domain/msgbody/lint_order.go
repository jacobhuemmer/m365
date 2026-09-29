package msgbody

import "sort"

// position orders problems: the token where a problem starts, then the
// byte offset within buffered text.
type position struct{ tok, off int }

// textSeg maps a byte offset in the pending text to its source token.
type textSeg struct{ off, tok int }

// ordered returns the problems sorted by the token where each starts;
// problems from the same token keep the order they were found in.
func (l *linter) ordered() Problems {
	idx := make([]int, len(l.out))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		pa, pb := l.seqs[idx[a]], l.seqs[idx[b]]
		return pa.tok < pb.tok || (pa.tok == pb.tok && pa.off < pb.off)
	})
	out := make(Problems, 0, len(l.out))
	for _, i := range idx {
		out = append(out, l.out[i])
	}
	return out
}
