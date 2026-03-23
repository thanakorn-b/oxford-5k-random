package main

import (
	"fmt"
	"os"
)

// styles holds ANSI SGR codes; empty strings mean plain output.
type styles struct {
	bold, dim, green, gold, cyan, magenta, reset string
}

func isTerminal(f *os.File) bool {
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}

func newStyles(color bool) styles {
	if !color {
		return styles{}
	}
	return styles{
		bold:    "\033[1m",
		dim:     "\033[2m",
		green:   "\033[32m",
		gold:    "\033[33m",
		cyan:    "\033[36m",
		magenta: "\033[35m",
		reset:   "\033[0m",
	}
}

func printDimLine(s styles, line string) {
	if s.dim != "" {
		fmt.Printf("%s%s%s\n", s.dim, line, s.reset)
		return
	}
	fmt.Println(line)
}

func printGameBanner(s styles) {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║     OXFORD 5K · CLI VOCAB GAME       ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	sub := "  Study · Thai hints · Self-score challenge"
	if s.dim != "" {
		fmt.Printf("%s%s%s\n\n", s.dim, sub, s.reset)
	} else {
		fmt.Println(sub)
		fmt.Println()
	}
}

func printMainMenu(s styles) {
	b, r, d := s.bold, s.reset, s.dim
	if b == "" {
		b, r, d = "", "", ""
	}
	fmt.Printf("%s── MAIN MENU ──%s\n", b, r)
	fmt.Println("  n, next      — new word + part of speech & CEFR level")
	fmt.Println("  t, translate — Thai for last word (online; saved in -cache file)")
	fmt.Printf("  %sc%s, challenge, play — vocabulary challenge (score at the end)\n", b, r)
	fmt.Println("  h, help      — this list")
	fmt.Println("  menu         — show menu again")
	fmt.Println("  q, quit      — exit game")
	if d != "" {
		fmt.Printf("%s  (in challenge: know · unknown · translate · q)%s\n", d, r)
	}
	fmt.Println()
}

func tierMessage(pct float64) string {
	switch {
	case pct >= 90:
		return "Outstanding — you know your advanced core vocabulary."
	case pct >= 70:
		return "Great run — solid coverage."
	case pct >= 50:
		return "Good effort — keep drilling the gaps."
	default:
		return "Keep going — unknown words are where the growth is."
	}
}

func printChallengeSummary(s styles, known, unknown, bestStreak int) {
	total := known + unknown
	fmt.Println()
	if total == 0 {
		if s.dim != "" {
			fmt.Printf("%sNo rounds played — see you next time.%s\n", s.dim, s.reset)
		} else {
			fmt.Println("No rounds played — see you next time.")
		}
		fmt.Println()
		return
	}
	pct := float64(known) / float64(total) * 100
	b, g, r := s.bold, s.green, s.reset
	if b == "" {
		b, g, r = "", "", ""
	}
	fmt.Printf("%s╔══ CHALLENGE RESULT ══╗%s\n", b, r)
	fmt.Printf("  Known:    %s%d%s / %d  (%.1f%%)\n", g, known, r, total, pct)
	fmt.Printf("  Unknown:  %d\n", unknown)
	if bestStreak >= 2 {
		fmt.Printf("  Longest \"know\" streak: %d\n", bestStreak)
	}
	fmt.Printf("%s╚══════════════════════╝%s\n", b, r)
	fmt.Println()
	if s.gold != "" {
		fmt.Printf("%s%s%s\n", s.gold, tierMessage(pct), s.reset)
	} else {
		fmt.Println(tierMessage(pct))
	}
	fmt.Println()
}

func showChallengeWord(s styles, round, streak int, word string) {
	d, c, m, r := s.dim, s.cyan, s.magenta, s.reset
	if d == "" {
		fmt.Printf("Round %d", round)
		if streak > 0 {
			fmt.Printf(" · streak %d", streak)
		}
		fmt.Println()
		fmt.Println(word)
		return
	}
	fmt.Printf("%s── Round %d", d, round)
	if streak > 0 {
		fmt.Printf("%s · %sstreak %d%s", d, c, streak, d)
	}
	fmt.Printf(" ──%s\n", r)
	if m != "" {
		fmt.Printf("%s%s%s\n", m, word, r)
	} else {
		fmt.Println(word)
	}
}
