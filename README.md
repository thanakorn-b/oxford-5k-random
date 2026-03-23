# Oxford 5K — CLI vocabulary game

Terminal practice for the **Oxford 5000** (American English) extended list: random words, optional **English → Thai** hints via [MyMemory](https://mymemory.translated.net/), and a **self-scored challenge** mode. Translations you fetch are saved locally so repeat lookups stay offline.

## Requirements

- **Go** 1.24+ (see `go.mod`)
- **Network** when translating (first time per headword, unless it is already in the cache)
- Word list: **`american_oxford_5000.json`** (included). To rebuild from the official PDF, use **`extract_oxford_pdf.py`** (needs Python + `pypdf`; see [requirements.txt](requirements.txt))

## Quick start

```bash
cd oxford-5k-random
go run .
```

Build a binary:

```bash
go build -o oxford5k .
./oxford5k
```

## Main commands

| Input | Action |
|--------|--------|
| `n`, `next` | Random entry: headword, part(s) of speech, CEFR level (B2 / C1) |
| `t`, `translate` | Thai for the **last** word shown by `n` (uses cache when possible) |
| `c`, `challenge`, `play` | Challenge: **word only**; you mark know / unknown / ask for translate |
| `h`, `help`, `menu`, `m` | Show the menu |
| `q`, `quit`, `exit` | Quit |

### Challenge mode

- **`know`** / **`k`** (also `yes`, `y`) — I know it → next word, streak +1  
- **`unknown`** / **`u`** / **`unknow`** / **`no`** — I don’t → next word, streak resets  
- **`translate`** / **`t`** — Thai + part-of-speech line for the **current** word (hint; same word until you answer)  
- **`q`** / **`quit`** — End run: **known / total**, **%**, longest **know** streak (if ≥ 2), short tier message  

## Flags

| Flag | Default | Meaning |
|------|---------|---------|
| `-json` | `american_oxford_5000.json` | Path to the lexicon JSON |
| `-cache` | `translation_cache.json` | Read/write Thai cache (JSON) |
| `-email` | *(empty)* | Optional contact email for MyMemory (higher daily quota) |
| `-no-color` | off | Disable ANSI colors (e.g. for logs or pipes) |

Example:

```bash
go run . -json ./data/words.json -cache ./data/th.json -email you@example.com
```

## Data files

### Lexicon (`american_oxford_5000.json`)

- `entries[]`: each item has `index` (0-based list position), `word`, `senses[]` with `part_of_speech` and `level`.
- Regenerate from PDF:  
  `python3 extract_oxford_pdf.py /path/to/American_Oxford_5000.pdf -o american_oxford_5000.json`

### Translation cache (`translation_cache.json`)

Created automatically. Shape:

```json
{
  "version": 1,
  "by_word": {
    "example": { "index": 0, "th": "…" }
  }
}
```

Keys are **exact headword strings** as in the lexicon. If saving fails, you still see the translation; a **warning** is printed to **stderr**.

## Project layout

| File | Role |
|------|------|
| `main.go` | Flags, stdin loop, wiring |
| `lexicon.go` | Types, load lexicon, format lines |
| `translate.go` | MyMemory client + JSON cache |
| `ui.go` | Banner, menu, colors, challenge summary |
| `challenge.go` | Challenge REPL |
| `extract_oxford_pdf.py` | PDF → lexicon JSON |

## Notes

- **MyMemory** is a free tier service; quality and limits vary. For production or heavy use, consider a paid translation API.
- The **Oxford** word list is subject to **Oxford University Press** copyright; this repo only provides tooling and a **sample/export workflow** — ensure your use complies with their terms.
