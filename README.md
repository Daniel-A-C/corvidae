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

Corvidae organizes flashcard decks into language directories and subdirectories under the `decks/` folder (for example, `decks/Mandarin/Basics/` or `decks/Spanish/`). Any `.yaml` deck placed inside a directory or subfolder will be automatically discovered. The structure is simple, allowing you or an AI agent to generate new decks instantly.

Example `decks/Mandarin/Basics/basics.yaml`:
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

For languages requiring phonetic transliterations (such as Arabic), you can use the `pronunciation` field. For Latin-alphabet languages (such as Spanish, French, or Polish), `character`, `meaning`, and `explanation` are sufficient.

> **Note:** The app will automatically inject tracking fields (`ease`, `interval`, `reps`, `next_review`) into the YAML file as you study to persist your progress.

## Usage & Controls

When you launch Corvidae, you will be guided through a selection workflow:

1. **Practice Mode**: Choose between SM-2 Spaced Repetition or Multiple Choice Quiz.
2. **Directory / Language Selection**: Navigate folders and subdirectories (e.g. `Mandarin/Balatro`, `Mandarin/Disney`, `Spanish`).
3. **Deck Selection**: Pick one or more decks. Decks selected across different folders and subfolders remain selected.

- `j` / `k` or Arrow Keys: Navigate menus.
- `Spacebar`: Toggle deck selections on/off.
- `Enter`: Confirm selection, enter folder, or start practice.
- `Tab`: Start practice immediately from directory navigation when decks are selected.
- `Esc` or `b`: Go up a level / back to the previous menu.
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

## Memorize by Options

Corvidae includes a dedicated **Memorize by Options** mode designed to help you memorize complete texts word-by-word through active recall:

1. **Text Discovery**: Place any text files in the `memorizationTexts/` directory (e.g., `memorizationTexts/filemon.yaml` or `filemon.txt`). Corvidae automatically scans and detects all available texts in this folder.
2. **Continuous Portion Selector**: After choosing a text, pick any continuous range of verses/sections to practice (e.g. Verses 15–20):
   - **Independent Boundary Shifting**: Shift the Start and End bounds independently using `[Tab / ↑↓]` to switch focus and `[←→ / hl]` to adjust, or use direct hotkeys `[` / `]` for Start and `{` / `}` (or `-` / `+`) for End.
   - **Expand & Advance**: Quickly expand `[x]` the portion (+1 verse) or advance `[n]` to the next continuous chunk.
   - **Looping Workflow**: Once you finish a portion, you are automatically returned to the Portion Selector with completion stats, ready to re-practice, expand, or advance.
3. **Word-by-Word Flow**: As you work through the text, each consecutive word presents 5 plausible options.
4. **Home Row or Numeric Selection**: Select options instantly using `a, s, d, f, g` or numbers `1-5`.
5. **Immediate Flow**: Correct selections advance to the next word immediately for fluid recitation and memorization.
6. **Mistake Feedback**: If an incorrect option is chosen, the interface highlights the error, displays the correct word, and allows you to retry or continue.
7. **Progress Persistence**: Progress is saved periodically, allowing you to resume long texts where you left off.

To supply custom texts, create a YAML file in `memorizationTexts/`:
```yaml
title: "Epístola a Filemón"
language: "Spanish"
words:
  - word: "Pablo,"
    prefix: "1 "
    options:
      - "Pablo,"
      - "Pedro,"
      - "Juan,"
      - "Lucas,"
      - "Santiago,"
  - word: "prisionero"
    options:
      - "prisionero"
      - "siervo"
      - "apóstol"
      - "cautivo"
      - "enviado"
```
If a plain `.txt` file is placed in `memorizationTexts/` without a YAML counterpart, Corvidae will automatically generate plausible distractors and create the companion `.yaml` file.

## Learn by Reading (Diglot Weave)

Corvidae features a **Learn by Reading** mode based on the *Diglot Weave* technique. Read through beloved scripture or literature while target vocabulary is progressively woven directly into the text.

1. **Center-Justified Terminal Reader**: Beautiful, distraction-free center-justified reader layout designed for reading the Bible (1 John) or literature (Harry Potter) in the terminal.
2. **Progressive Vocabulary Immersion**: Words in each sentence are replaced with target language equivalents (e.g. Mandarin Chinese characters).
3. **Active Recall & Word Lookup**:
   - If you know all words in the sentence, press `[Space]`, `[Enter]`, or `[n]` to proceed.
   - If you encounter an unfamiliar word, press its assigned shortcut key (e.g. `[a]`, `[s]`, `[d]`) to pop up its definition, Pinyin transliteration, and grammatical breakdown.
   - The system automatically updates its tracking of words you know vs. words you are learning.
4. **Instant Flashcard Integration**: Press `[m]` in the word detail popup to add any unfamiliar word directly to your **Marked Cards** deck for SM-2 Spaced Repetition review!
5. **Real-Time Aggressiveness Tuning**: Adjust substitution aggressiveness on the fly using `+` / `-` (or `[` / `]`):
   - **0%**: Pure English reading mode (uninterrupted Bible or book reading in the terminal).
   - **20% (Gentle)**: Reviews known words and introduces ~1 new word per sentence.
   - **40% (Balanced)**: Balanced vocabulary immersion.
   - **70% (Intensive)**: Accelerating vocabulary exposure.
   - **100% (Full Immersion)**: Every translatable word replaced.
6. **Sentence Grammar & Discussion Box**: Pops up after each sentence (or toggle with `[h]`), detailing:
   - Natural target language sentence in proper word order (e.g. full Mandarin sentence + Pinyin).
   - In-depth grammar notes explaining syntax differences (e.g., relative clauses, prepositional order, aspect particles like 的, 了, 着).
   - Substituted vocabulary recap table.
7. **Extensible Texts**: Place `.yaml` or `.txt` texts in `readingTranslationTexts/` (configurable with `-r /path/to/readingTranslationTexts`). Initial texts include *1 John* and *Harry Potter Chapter 1*.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
