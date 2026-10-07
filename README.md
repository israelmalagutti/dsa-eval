# dsa-eval

A terminal exam that measures how well you recognize DSA patterns. Each question is a short problem. You name the technique, the signal that points to it, its complexity, and the key idea. Then you grade yourself against a reference answer. Over time, `report` builds a per-topic map of your strengths and gaps.

## Install

```sh
go install github.com/israelmalagutti/dsa-eval@latest
```

## Usage

```sh
dsa-eval -quick       # 6 questions, ~15 min
dsa-eval              # 10 questions, ~30 min
dsa-eval -n 4         # custom length
dsa-eval resume [ID]  # continue an unfinished exam
dsa-eval list         # all exams
dsa-eval report       # tiers per category and topic
dsa-eval refs         # topic dependency map with study references
```

- Sessions are saved as JSON in `~/.local/share/dsa-eval/snapshots` (or under `$XDG_DATA_HOME`). Use `-dir` to change the location.
- In an exam, Shift+←/→ moves between questions and Ctrl+S submits. Type `:h` in a field for help.

### Scoring

| Tier | Meaning |
|---|---|
| 0 Unknown | Wrong technique |
| 1 Recognize | Right technique and signal |
| 2 Can implement | Also correct complexity and key idea |
| 3 Fluent | Tier 2, within 3 minutes |

## Question bank

The reference answers in `questions.json` are checked by executable tests in `verify/`. Each key idea is implemented exactly as written and compared against a brute-force solution on random inputs:

```sh
go test ./...
```

## References and CLRS

`dsa-eval refs` shows a dependency map of topics. Each topic has a summary, links, and section and page references to *Introduction to Algorithms*, 3rd edition (Cormen, Leiserson, Rivest, Stein).

The book is **not** included. The summaries are original, and the page references are citations. To open a reference at the right page, point `CLRS_PDF` at your own legally obtained copy:

```sh
export CLRS_PDF="$HOME/books/clrs-3e.pdf"
export CLRS_VIEWER=zathura   # optional; defaults to your system PDF handler
```

## Disclaimer

This project is not affiliated with or endorsed by LeetCode, MIT Press, or the authors of *Introduction to Algorithms*. Problems are original paraphrases of classic algorithm exercises. Links point to the original sources.

## License

MIT
