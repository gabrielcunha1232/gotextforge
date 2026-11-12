package fuzzy

import (
	"sort"
	"strings"
	"unicode"
)

type Match struct {
	Text      string
	Score     int
	Positions []int
}

func isSeparator(r rune) bool {
	return r == '/' || r == '\\' || r == '_' || r == '-' || r == '.' || r == ' '
}

func Score(text string, query string) (int, []int, bool) {
	if query == "" {
		return 0, nil, true
	}
	t := []rune(text)
	q := []rune(strings.ToLower(query))
	lower := []rune(strings.ToLower(text))

	nameStart := 0
	for i := len(t) - 1; i >= 0; i-- {
		if t[i] == '/' || t[i] == '\\' {
			nameStart = i + 1
			break
		}
	}

	best, bestPos, found := 0, []int(nil), false
	for _, from := range []int{nameStart, 0} {
		s, pos, ok := scoreFrom(t, lower, q, from, nameStart)
		if ok && (!found || s > best) {
			best, bestPos, found = s, pos, true
		}
		if nameStart == 0 {
			break
		}
	}
	if !found {
		return 0, nil, false
	}
	if strings.HasPrefix(strings.ToLower(string(t[nameStart:])), string(q)) {
		best += 20
	}
	return best - len(t)/8, bestPos, true
}

func scoreFrom(t, lower, q []rune, from, nameStart int) (int, []int, bool) {
	positions := make([]int, 0, len(q))
	score := 0
	qi := 0
	prev := -2
	for i := from; i < len(t) && qi < len(q); i++ {
		if lower[i] != q[qi] {
			continue
		}
		bonus := 1
		switch {
		case i == prev+1:
			bonus += 8
		case i == 0 || isSeparator(t[i-1]):
			bonus += 10
		case unicode.IsUpper(t[i]) && i > 0 && unicode.IsLower(t[i-1]):
			bonus += 6
		}
		if i >= nameStart {
			bonus += 4
		}
		if prev >= 0 {
			bonus -= min(i-prev-1, 5)
		}
		score += bonus
		positions = append(positions, i)
		prev = i
		qi++
	}
	return score, positions, qi == len(q)
}

func Filter(items []string, query string, limit int) []Match {
	var out []Match
	for _, it := range items {
		if s, pos, ok := Score(it, query); ok {
			out = append(out, Match{Text: it, Score: s, Positions: pos})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return len(out[i].Text) < len(out[j].Text)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}
