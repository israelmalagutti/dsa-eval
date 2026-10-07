# eval

A Go CLI for pattern-recognition DSA exams. Run it from this directory:

- `go run . [-quick | -n N]`: start a new exam. Quick is 6 questions (~15 min); the default is 10 (~30 min).
- `go run . resume [ID]`: continue an unfinished exam. With no ID it picks the latest one.
- `go run . list`: list all exams (ID, start date, mode, status), newest first. Exam IDs are the first 6 hex digits of a SHA-256, shown as `EXAM-1A2F24`; `resume` accepts either form, in any case.
- `go run . report`: show tiers per category and per topic, plus the 5 weakest topics.
- `go run . refs`: browse the references map, a dependency tree of topics with your progress per node, CLRS sections and links. Enter opens a node; `o` opens the selected CLRS ref in the PDF (`CLRS_PDF` sets its path, `CLRS_VIEWER` the viewer). In review, `r` opens it on the current question's node. Data is in `references.json`.

The exam is a full-screen Bubble Tea UI, one question per screen. Shift+←/→ moves between questions. There are two phases:

1. Answering: answer in any order. Time counts only while a question is on screen and adds up across visits. Ctrl+S submits.
2. Review: answers are locked, and you self-grade each one against the reference (keys 0/1/2).

`submitted_at` marks the switch to review; `resume` reopens whichever phase you were in. Each session is saved to `~/.local/share/dsa-eval/snapshots/<id>.json` (`$XDG_DATA_HOME` is honored; `-dir` overrides) on every navigation, grade and quit. Questions are in `questions.json`. Entries hold one item per plan question; `self_grade` and `self_tier` are null until graded.

## Rubric

Each answer has four parts: technique, signal, complexity, key idea.

- Grade 0: wrong technique, or skipped.
- Grade 1: right technique and signal.
- Grade 2: grade 1, plus correct complexity and a sound key idea.
- Tier = grade, except that grade 2 with `duration_sec <= 180` becomes tier 3 (Fluent).

## Grading a snapshot (when asked to "grade" one)

For each entry in `~/.local/share/dsa-eval/snapshots/<id>.json`, compare `answer` with the reference for `question_id` in `questions.json`:

- Set `ai_grade` to 0, 1 or 2 using the rubric. Accept any valid alternative technique, and judge meaning, not wording.
- Set `ai_tier` to the tier computed from `ai_grade` and `duration_sec`.
- Set `ai_feedback` to 1–2 sentences: what was right, and what was missing or wrong.
- Leave every other field unchanged. `report` uses `ai_tier` instead of `self_tier` when it is set.

Then give the user a short summary: the score, and the topics to review.
