package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	jsonPath := flag.String("json", "american_oxford_5000.json", "path to Oxford 5000 JSON export")
	cachePath := flag.String("cache", "translation_cache.json", "path to Thai translation cache (read/write)")
	email := flag.String("email", "", "optional email for MyMemory API (higher daily quota)")
	noColor := flag.Bool("no-color", false, "disable ANSI colors (plain text)")
	flag.Parse()

	d, err := loadLexicon(*jsonPath)
	if err != nil {
		log.Fatalf("lexicon: %v", err)
	}

	tcache, err := loadTranslationCache(*cachePath)
	if err != nil {
		log.Fatalf("translation cache: %v", err)
	}

	httpClient := &http.Client{Timeout: 20 * time.Second}
	useColor := isTerminal(os.Stdout) && !*noColor
	st := newStyles(useColor)

	printGameBanner(st)
	printMainMenu(st)

	var last entry
	var hasLast bool
	sc := bufio.NewScanner(os.Stdin)

	for {
		fmt.Fprint(os.Stderr, "> ")
		if !sc.Scan() {
			break
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		cmd := strings.ToLower(line)

		switch cmd {
		case "q", "quit", "exit":
			if st.cyan != "" {
				fmt.Printf("%sThanks for playing. Goodbye!%s\n", st.cyan, st.reset)
			} else {
				fmt.Println("Thanks for playing. Goodbye!")
			}
			return

		case "h", "help", "?", "menu", "m":
			printMainMenu(st)

		case "n", "next":
			last = d.Entries[rand.IntN(len(d.Entries))]
			hasLast = true
			fmt.Println(formatEntry(last))

		case "t", "translate":
			if !hasLast {
				fmt.Println("Draw a card first: type n for a word.")
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
			th, err := translateWithCache(ctx, httpClient, *email, tcache, last.Word, last.Index)
			cancel()
			if err != nil {
				fmt.Printf("translate: %v\n", err)
				continue
			}
			printDimLine(st, th)

		case "c", "challenge", "ch", "play", "p":
			runChallenge(sc, d.Entries, httpClient, *email, tcache, st)

		default:
			fmt.Println("Unknown command. Try: n, t, c, menu, h, or q.")
		}
	}

	if err := sc.Err(); err != nil {
		log.Fatalf("stdin: %v", err)
	}
}
