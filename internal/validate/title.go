package validate

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Observation-title house style (VALIDATOR.md §14-17): sentence case,
// concise, factual, no trailing period, detail in the body.
//
// Hard boundaries: nothing here rewrites anything. Suggestions are
// display-only and must differ from the original by capitalization or
// trailing punctuation only — never by adding, removing, or reordering
// words. All-caps titles get no suggestion at all, because a
// mechanical lowercase cannot distinguish acronyms from shouting.

const maxTitleLen = 60

func titleStyle(r *report, title string) {
	if title == "" {
		return
	}
	where := fmt.Sprintf("output: title %q", clip(title, 50))

	base := strings.TrimRight(title, ".!")
	if base != title && base != "" {
		r.warnSuggest(where,
			"titles are headlines and carry no trailing period or exclamation mark",
			base)
	}
	if base == "" {
		base = title
	}

	if isAllCaps(title) {
		r.warnf(where, "title is all-caps — the voice is calm; use sentence case (no suggestion offered: acronyms make a mechanical lowercase unsafe)")
	} else if looksTitleCased(base) {
		r.warnSuggest(where,
			"title looks Title-Cased — house style is sentence case",
			sentenceCase(base)+"   (verify proper nouns)")
	}

	if utf8.RuneCountInString(title) > maxTitleLen {
		r.warnf(where, "title is over %d characters — move the specifics into the observation body", maxTitleLen)
	}
}

// isAllCaps: four or more letters, every one of them uppercase.
func isAllCaps(s string) bool {
	letters := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters++
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return letters >= 4
}

// looksTitleCased: three or more words, and most non-leading words of
// four or more letters begin with a capital. Fully-uppercase tokens
// (RAID, GB, S02E04) are treated as acronyms and don't count.
func looksTitleCased(s string) bool {
	words := strings.Fields(s)
	if len(words) < 3 {
		return false
	}
	eligible, capped := 0, 0
	for _, w := range words[1:] {
		first, _ := utf8.DecodeRuneInString(w)
		if !unicode.IsLetter(first) || utf8.RuneCountInString(w) < 4 || isAcronymToken(w) {
			continue
		}
		eligible++
		if unicode.IsUpper(first) {
			capped++
		}
	}
	return eligible >= 2 && capped*2 > eligible
}

// sentenceCase lowers the first rune of each capitalized non-leading
// word, leaving acronym tokens untouched. Capitalization-only: word
// count and order are preserved by construction.
func sentenceCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words[1:] {
		if isAcronymToken(w) {
			continue
		}
		first, size := utf8.DecodeRuneInString(w)
		if unicode.IsUpper(first) {
			words[i+1] = string(unicode.ToLower(first)) + w[size:]
		}
	}
	return strings.Join(words, " ")
}

// isAcronymToken: two or more letters, all uppercase (RAID, GB, S02E04).
func isAcronymToken(w string) bool {
	letters := 0
	for _, r := range w {
		if unicode.IsLetter(r) {
			letters++
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return letters >= 2
}
