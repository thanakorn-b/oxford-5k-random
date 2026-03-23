package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"
)

func runChallenge(sc *bufio.Scanner, entries []entry, client *http.Client, email string, cache *translationCache, s styles) {
	var known, unknown int
	var streak, bestStreak int
	round := 1
	current := entries[rand.IntN(len(entries))]

	b, r := s.bold, s.reset
	if b == "" {
		b, r = "", ""
	}
	fmt.Printf("%s┌─ CHALLENGE MODE ─┐%s\n", b, r)
	fmt.Println("  You only see the English headword.")
	fmt.Println("  know, k      → I know it (next word)")
	fmt.Println("  unknown, u   → I don't (unknow works too)")
	fmt.Println("  translate, t → Thai + part of speech (hint; same word until you answer)")
	fmt.Println("  q, quit      → end run & see score")
	fmt.Printf("%s└──────────────────┘%s\n", b, r)
	fmt.Println()
	showChallengeWord(s, round, streak, current.Word)

	for {
		fmt.Fprint(os.Stderr, "challenge> ")
		if !sc.Scan() {
			printChallengeSummary(s, known, unknown, bestStreak)
			return
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		cmd := strings.ToLower(line)
		switch cmd {
		case "q", "quit", "done":
			printChallengeSummary(s, known, unknown, bestStreak)
			return
		case "know", "k", "yes", "y":
			known++
			streak++
			if streak > bestStreak {
				bestStreak = streak
			}
			round++
			current = entries[rand.IntN(len(entries))]
			showChallengeWord(s, round, streak, current.Word)
		case "unknown", "unknow", "u", "no":
			unknown++
			streak = 0
			round++
			current = entries[rand.IntN(len(entries))]
			showChallengeWord(s, round, streak, current.Word)
		case "translate", "t", "th":
			ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
			th, err := translateWithCache(ctx, client, email, cache, current.Word, current.Index)
			cancel()
			if err != nil {
				fmt.Printf("translate: %v\n", err)
				continue
			}
			printDimLine(s, th)
			if senses := formatSensesLine(current); senses != "" {
				printDimLine(s, senses)
			}
		case "h", "help", "?":
			fmt.Println("know · unknown · translate · q")
		default:
			fmt.Println("Try: know, unknown, translate, or q to finish.")
		}
	}
}
