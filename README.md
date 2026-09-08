# Corvidae

A fast, lightweight terminal flashcard application built in Go. Corvidae uses the SuperMemo-2 (SM-2) Spaced Repetition algorithm to optimize memory retention, keeping your hands on the home row for rapid, friction-free study sessions.

Designed to be extended instantly using human-readable YAML files.

## Features

- **Terminal-Native:** Built with `bubbletea` and `lipgloss` for a fast, responsive UI.
- **Home Row Grading:** Never move your hands. Grade your recall using `d, f, g, h, j, k`.
- **Spaced Repetition (SM-2):** Automatically schedules cards based on your past performance.
- **YAML Decks:** Create, edit, and share decks easily using plain text.
- **Focus Mode:** Toggle card explanations on and off with a single keystroke.
- **Force Review:** Cram ahead of schedule with the "review all" mode.

## Installation

Ensure you have [Go installed](https://go.dev/doc/install) on your system.

1. Clone the repository:

   ```bash
   git clone https://github.com/Daniel-A-C/corvidae.git
   cd corvidae
   ```

2. Download dependencies and run:

   ```bash
   go mod tidy
   go run main.go
   ```

3. (Optional) Build an executable binary to run globally:

   ```bash
   go build -o corvidae main.go
   sudo mv corvidae /usr/local/bin/
   ```

   You can point Corvidae to your decks folder from anywhere:

   ```bash
   corvidae -d /path/to/decks
   ```

## Creating Decks

Corvidae organizes flashcard decks into language subdirectories under the `decks/` folder (for example, `decks/Mandarin/` or `decks/Spanish/`). Any `.yaml` deck placed inside a language directory will be automatically discovered. The structure is simple, allowing you or an AI agent to generate new decks instantly.

Example `decks/Mandarin/basics.yaml`:

```yaml
cards:
  - character: 你好
    pinyin: nǐ hǎo
    meaning: Hello
    explanation: "你 (nǐ) means 'you' and 好 (hǎo) means 'good' or 'well'. Literally: 'You good'."
  - character: 电脑
    pinyin: diànnǎo
    meaning: Computer
    explanation: "电 (diàn) means 'electric' and 脑 (nǎo) means 'brain'. Literally: 'Electric brain'."
```

> **Note:** The app will automatically inject tracking fields (`ease`, `interval`, `reps`, `next_review`) into the YAML file as you study to persist your progress.

## Usage & Controls

When you launch Corvidae, you will be guided through a selection workflow:

1. **Practice Mode**: Choose between SM-2 Spaced Repetition or Multiple Choice Quiz.
2. **Directory / Language Selection**: Select which language category you want to study (e.g. `Mandarin`, `Spanish`, `French`, `Polish`, `Arabic`).
3. **Deck Selection**: Pick one or more decks within that language directory.

- `j` / `k` or Arrow Keys: Navigate menus.
- `Spacebar`: Toggle deck selections on/off.
- `Enter`: Confirm selection or enter directory.
- `Esc` or `b`: Go back to the previous menu.
- `q`: Quit the app at any time.

During a review session, the controls map to a 6-point grading scale:

- `Spacebar`: Reveal the answer.
- `e`: Toggle the card explanation (if one exists).
- `d`: Blackout (Forgot completely)
- `f`: Wrong (Remembered incorrectly)
- `g`: Hard (Remembered, but with immense effort)
- `h`: Good (Remembered with moderate effort)
- `j`: Easy (Remembered instantly)
- `k`: Perfect (Effortless recall)
- `q`: Quit the app at any time.

When a deck is complete for the day, press `r` to force a full review of all cards in the deck, regardless of their due date.

## The Spaced Repetition Algorithm

Corvidae implements a modified version of the SuperMemo-2 (SM-2) algorithm.

When you grade a card, the algorithm calculates an "Ease" factor and determines the optimal number of days before you should see that card again.

- Cards you struggle with (`d`, `f`, `g`) will show up more frequently.
- Cards you know well (`j`, `k`) will have their intervals expanded aggressively, ensuring you spend your time strictly on what you are close to forgetting.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
