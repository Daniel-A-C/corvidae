# Corvidae

A fast, lightweight terminal memory workstation and language learning environment built in Go. Corvidae combines the SuperMemo-2 (SM-2) Spaced Repetition algorithm, multiple-choice testing, Diglot Weave reading immersion, word-by-word text recitation, and a specialized Arabic Alphabet Academy—keeping your hands on the home row for rapid, friction-free study sessions.

Designed to be extended instantly using human-readable YAML and plain-text files.

```
       _____                      _     _            
      / ____|                    (_)   | |           
     | |     ___  _ ____   ___  _  __| | __ _  ___ 
     | |    / _ \| '__\ \ / / |/ _` |/ _` |/ _ \
     | |___| (_) | |   \ V /| | (_| | (_| |  __/
      \_____\___/|_|    \_/ |_|\__,_|\__,_|\___|
```

---

## Features

- **Terminal-Native:** High-performance, responsive TUI built with Charm's `bubbletea` and `lipgloss`.
- **Spaced Repetition (SM-2):** Intelligent scheduling that optimizes recall intervals based on historical performance.
- **Multiple Choice Quiz Mode:** Active multiple-choice quizzes with context-matched distractors from your target language.
- **Arabic Alphabet Academy:** Comprehensive curriculum for learning the 28 Arabic letters across 4 cursive forms, phonetics, articulation points, Harakat (short vowels), auxiliary marks, and 6 interactive drills.
- **Learn by Reading (Diglot Weave):** Read literature or scripture while vocabulary is progressively substituted with target language words, accompanied by deep grammar discussions.
- **Memorize by Options:** Memorize long texts verse-by-verse with continuous portion selection, streak counters, and instant distractor choices.
- **Marked / Starred Cards:** Press `[m]` anywhere (review, quiz, reading) to star difficult words into a dedicated `marked.yaml` deck for focused drill sessions.
- **Hummingbird 30-Key Home-Row Selection:** Direct, zero-scroll access to any menu item, deck, or choice via ergonomic home-row letter mapping.
- **Persistent Deck Selection:** Active deck selections are saved to disk in your deck files (`selected: true`) and persist across sessions.
- **Human-Readable YAML & Text Files:** Simple, plain-text schemas make generating decks by hand or with AI agents effortless.

---

## Installation

Ensure you have [Go 1.20+](https://go.dev/doc/install) installed.

### 1. Clone the repository:

```bash
git clone https://github.com/Daniel-A-C/corvidae.git
cd corvidae
```

### 2. Download dependencies and run:

```bash
go mod tidy
go run main.go
```

### 3. Build an executable binary:

```bash
go build -o corvidae main.go
sudo mv corvidae /usr/local/bin/
```

### CLI Options

Corvidae can be pointed to custom directories from anywhere:

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--decks` | `-d` | `decks` | Path to flashcard decks directory |
| `--texts` | `-t` | `memorizationTexts` | Path to memorization texts directory |
| `--reading` | `-r` | `readingTranslationTexts` | Path to reading translation texts directory |
| `--help` | `-h` | — | Show command-line help and usage |

Example:
```bash
corvidae -d ~/Documents/decks -r ~/Documents/readingTexts
```

---

## Modes & Workflows

When you launch Corvidae, you are presented with 5 core modes plus instant access to your Marked Cards deck:

```
Select Mode:
  [a] > Spaced Repetition (SM-2)
  [s]   Multiple Choice Quiz
  [d]   Arabic Alphabet Academy
  [f]   Memorize by Options
  [g]   Learn by Reading

  [m] ★ Practice Marked Cards (14 cards)
```

Use home-row shortcuts `[a, s, d, f, g]` or `[j/k / ↑↓]` to select a mode. Press `[m]` to jump immediately into your starred cards.

---

### 1. Spaced Repetition (SM-2)

Study flashcards scheduled for review today. Corvidae implements the SuperMemo-2 spaced repetition algorithm, calculating optimal intervals so you review cards right before you forget them.

#### Workflow:
1. Choose **Spaced Repetition (SM-2)** (`[a]`).
2. Navigate language folders and subdirectories (e.g. `Mandarin/Balatro`, `Spanish`).
3. Select one or more decks using `[Space]` or direct Hummingbird key. Selections persist automatically.
4. Press `[Enter]` or `[Tab]` to begin review.

#### Review Controls:
- `Spacebar`: Reveal the card's answer (pinyin/pronunciation and meaning).
- `e`: Toggle detailed etymological or grammatical explanation.
- `m`: Star/mark the card (adds to `decks/marked.yaml`).
- `u`: Unmark card.
- `d`: **Blackout** (Grade 0: Forgot completely)
- `f`: **Wrong** (Grade 1: Remembered incorrectly)
- `g`: **Hard** (Grade 2: Remembered with immense effort)
- `h`: **Good** (Grade 3: Remembered with moderate effort)
- `j`: **Easy** (Grade 4: Remembered easily)
- `k`: **Perfect** (Grade 5: Effortless, instant recall)
- `r`: Force review / cram all cards in the deck regardless of scheduled due date.
- `q`: Quit.

---

### 2. Multiple Choice Quiz

Test recognition and active recall against 5 multiple-choice options.

- Distractors are intelligently selected from cards in the same deck and language family.
- Immediate visual feedback highlights correct and incorrect responses.
- Press `[m]` during feedback to star any card you got wrong or want to review later.

#### Quiz Controls:
- `a, s, d, f, g` or `1, 2, 3, 4, 5`: Select option 1 through 5.
- `m`: Star/mark the current card.
- `u`: Unmark card.
- `Spacebar` / `Enter`: Advance to the next question.
- `r`: Retry the quiz after completion.
- `q`: Quit.

---

### 3. Arabic Alphabet Academy

A standalone learning environment for mastering the Arabic script from ground zero to fluency.

#### Core Modules:
1. **Interactive Alphabet Explorer (`[a]`):**
   - Browse all 28 core Arabic letters with complete cursive positional forms: **Isolated**, **Initial**, **Medial**, and **Final**.
   - Review pronunciation descriptions, transliterations, IPA transcriptions, memory mnemonics, and real-world example words.
   - Press `[Tab]` to switch to the **Auxiliary & Harakat** tab: learn short vowel marks (Fatha, Kasra, Damma, Sukun), Shaddah (gemination), Tanwin, Hamza seats, Ta Marbuta, and Alif Maqsura.
2. **8 Progressive Learning Stages (`[b]`):**
   - **Stage 1:** The Anchors & Boat Letters (Alif, Baa, Taa, Thaa)
   - **Stage 2:** The Crown Letters (Jeem, Haa, Khaa)
   - **Stage 3:** The Non-Connecting Sliders (Daal, Dhaal, Raa, Zaay)
   - **Stage 4:** The Whistling Teeth (Seen, Sheen, Saad, Daad)
   - **Stage 5:** The Emphatic Horns & Deep Throat (Taa, Dhaa, Ayn, Ghayn)
   - **Stage 6:** The Looped Tails & Hooks (Faa, Qaaf, Kaaf, Laam)
   - **Stage 7:** The Final Flow (Meem, Noon, Haa, Waaw, Yaa)
   - **Stage 8:** Auxiliary Characters & Diacritics (Harakat, Hamza, Ta Marbuta)
3. **6 Targeted Practice Drills (`[c - g]`):**
   - **Positional Forms Drill:** Identify isolated, initial, medial, and final shapes in context.
   - **Letter Sound & Phonetics Drill:** Match Arabic letters to phonetic descriptions and articulation points.
   - **Connector Drill:** Test recognition of the 6 non-connecting (right-joining only) letters (ا, د, ذ, ر, ز, و).
   - **Sun & Moon Letters Drill:** Master the 14 Sun letters (حروف شمسية) vs. 14 Moon letters (حروف قمرية) for the definite article (الـ).
   - **Confusable Pairs Drill:** Distinguish visually or phonetically similar pairs (e.g. ص vs س, ض vs د, ط vs ت, ق vs ك, ح vs هـ).
   - **Stage Mastery Drills:** Target questions exclusively from your active stage.

#### Academy Controls:
- In Explorer: `[Tab]` switches Letter / Auxiliary tabs; `[hl / ←→ / jk]` navigates; `[b / Esc]` goes back.
- In Stages: `[1-8]` selects stage; inside a stage: `[Enter]` or `[d]` launches the stage drill.
- In Drills: `[a, s, d, f]` or numbers `[1-4]` select answers; `[Space / Enter / n]` advances after feedback.

---

### 4. Memorize by Options

Memorize complete texts (scripture, poetry, speeches, prose) word-by-word through active distractor selection.

1. **Continuous Portion Selector:** Select any continuous range of verses/sections to practice:
   - Use `[Tab / ↑↓]` to switch focus between Start and End bounds.
   - Use `[←→ / hl]` to adjust the focused boundary.
   - Hotkeys: `[` / `]` adjust Start; `{` / `}` (or `-` / `+`) adjust End.
   - Quick Actions: `[x]` expands portion (+1 verse); `[n]` advances to the next continuous portion chunk.
2. **Word-by-Word Flow:** Each word presents 5 plausible options. Select using home row `[a, s, d, f, g]` or `[1-5]`.
3. **Streak & Mistake Feedback:** Correct answers advance instantly for fluid recitation; mistakes are highlighted with corrective feedback.
4. **Auto-Generation:** Place any raw `.txt` file in `memorizationTexts/`; Corvidae automatically parses verses, generates plausible distractors, and creates a companion `.yaml` file.

Example `memorizationTexts/filemon.yaml`:
```yaml
title: "Epístola a Filemón"
language: "Spanish"
words:
  - word: "Pablo,"
    prefix: "1 "
    options: ["Pablo,", "Pedro,", "Juan,", "Lucas,", "Santiago,"]
  - word: "prisionero"
    options: ["prisionero", "siervo", "apóstol", "cautivo", "enviado"]
```

---

### 5. Learn by Reading (Diglot Weave)

Read literature or scripture in English while vocabulary from your target language is progressively woven into the text.

1. **Distraction-Free Terminal Reader:** Beautiful, center-justified reading layout designed for reading long-form texts (e.g. *1 John* or *Harry Potter Chapter 1*).
2. **Real-Time Aggressiveness Tuning (`+` / `-` / `[` / `]`):**
   - **0% (Pure Reading):** Uninterrupted English reading.
   - **20% (Gentle):** Reviews known words and weaves ~1 new word per sentence.
   - **40% (Balanced):** Balanced vocabulary immersion.
   - **70% (Intensive):** High-density vocabulary immersion.
   - **100% (Full Immersion):** Every translatable word replaced.
3. **Word Inspection:** Unfamiliar words display shortcut badges (`[a]`, `[s]`, `[d]`, `[z]`, etc.). Press the key to view:
   - Target characters, Pinyin transliteration, pronunciation, and meaning.
   - Detailed grammatical explanation.
   - **Instant Flashcard Creation (`[m]`):** Press `[m]` to immediately add the word to your Marked Cards deck!
4. **Sentence Grammar & Discussion Box (`[h]` or automatically on sentence complete):**
   - Natural target-language sentence in proper word order (e.g. Chinese characters + Pinyin).
   - In-depth grammar analysis explaining syntactic differences (aspect markers, clause ordering, particles).
   - Substituted vocabulary recap table.
5. **Mastery Tracking:** Automatically tracks word encounters: `Unknown` ➔ `Learning` ➔ `Known`.

---

### 6. Marked Cards System

Corvidae features a unified **Marked Cards** workflow across all modes:

- **Starring Cards:** Press `[m]` while reviewing cards in SM-2 mode, answering multiple-choice quiz questions, or inspecting unfamiliar words in Learn by Reading.
- **Disk Persistence:** Marked cards are persisted to `decks/marked.yaml` with language and deck origin metadata.
- **Direct Practice:** From the main menu, press `[m]` to start an immediate practice session with your starred cards.
- **Language Distractor Matching:** When quizzing marked cards, Corvidae automatically generates distractors matching the card's native language.

---

## The Hummingbird 30-Key Input System

Corvidae implements the **Hummingbird** ergonomic input layout across all menus, directory selectors, deck lists, and quiz options:

```
[q] [w] [e] [r] [t]   [y] [u] [i] [o] [p]   <-- Top Row
[a] [s] [d] [f] [g]   [h] [j] [k] [l] [;]   <-- Home Row
[z] [x] [c] [v] [b]   [n] [m] [,] [.] [/]   <-- Bottom Row
```

- **Group 1 (Home Left):** `a, s, d, f, g` (Items 1–5)
- **Group 2 (Home Right):** `h, j, k, l, ;` (Items 6–10)
- **Group 3 (Bottom Left):** `z, x, c, v, b` (Items 11–15)
- **Group 4 (Bottom Right):** `n, m, ,, ., /` (Items 16–20)
- **Group 5 (Top Left):** `q, w, e, r, t` (Items 21–25)
- **Group 6 (Top Right):** `y, u, i, o, p` (Items 26–30)

Every listed item displays its shortcut badge (e.g. `[a]`, `[s]`, `[h]`). You can select any item instantly without scrolling.

---

## Directory Structure

```
corvidae/
├── decks/                             # Spaced repetition decks by language
│   ├── Arabic/
│   ├── French/
│   ├── Mandarin/
│   │   ├── Balatro/                   # balatro1.yaml - balatro8.yaml
│   │   ├── Basics/                    # basics.yaml, numbers.yaml
│   │   ├── Conversation/              # getting_acquainted_1.yaml - 3.yaml
│   │   ├── Disney/                    # movie vocabulary decks
│   │   ├── Geography/                 # cities.yaml, regions.yaml
│   │   ├── Misc/                      # misc.yaml - misc4.yaml
│   │   ├── Names/                     # chinese_names.yaml, transliterations.yaml
│   │   └── Topics/
│   ├── Polish/
│   ├── Spanish/
│   └── marked.yaml                    # Auto-managed marked/starred cards
│
├── memorizationTexts/                 # Texts for Memorize by Options
│   ├── filemon.txt
│   └── filemon.yaml
│
├── readingTranslationTexts/           # Texts for Learn by Reading
│   ├── 1John.txt / 1John.yaml
│   ├── HarryPotterCh1.txt / .yaml
│   └── reading_progress.yaml          # Auto-saved reader mastery & bookmarks
│
├── internal/
│   ├── arabic/                        # Alphabet data, cursive forms, drills
│   ├── deck/                          # Deck models, YAML I/O, persistence
│   ├── memorize/                      # Portion selector, text parsing, distractors
│   ├── quiz/                          # Distractor generator & quiz engine
│   ├── reading/                       # Diglot Weave substitution & progress
│   ├── sm2/                           # SuperMemo-2 algorithm implementation
│   └── ui/                            # Bubble Tea views, update loop, Hummingbird
│
├── main.go                            # Entry point & CLI flag configuration
└── README.md
```

### Creating Flashcard Decks

Flashcard decks are simple `.yaml` files placed in any subdirectory under `decks/`:

```yaml
cards:
  - character: 电脑
    pinyin: diànnǎo
    meaning: Computer
    explanation: "电 (diàn) means 'electric' and 脑 (nǎo) means 'brain'. Literally: 'Electric brain'."
  - character: مَرْحَبًا
    pronunciation: marḥaban
    meaning: Hello
    explanation: Standard friendly greeting in Modern Standard Arabic.
```

- `character`: The target word or phrase.
- `pinyin`: Phonetic reading for Mandarin.
- `pronunciation`: Transliteration for Arabic or other non-Latin scripts.
- `meaning`: English translation.
- `explanation`: Etymology, literal translation, or usage tips (toggled with `[e]`).
- Spaced repetition fields (`interval`, `ease`, `reps`, `next_review`) and `selected: true` are managed automatically by Corvidae.

---

## Keyboard Shortcuts Reference

| Context | Key | Action |
| :--- | :--- | :--- |
| **Global** | `q` | Quit application |
| **Global** | `ctrl+c` | Force quit |
| **Global** | `esc` / `b` | Go back / up one level |
| **Main Menu** | `a, s, d, f, g` | Select Mode (Review, Quiz, Arabic, Memorize, Reading) |
| **Main Menu** | `m` | Start Marked Cards practice session |
| **Deck Selection** | `Spacebar` | Toggle deck selection |
| **Deck Selection** | `u` / `ctrl+d` | Deselect all decks across all folders |
| **Deck Selection** | `Tab` | Start practice immediately with currently selected decks |
| **Deck Selection** | `Enter` | Enter folder / start practice |
| **Deck Selection** | `Hummingbird` | Jump directly to folder or toggle deck |
| **SM-2 Review** | `Spacebar` | Reveal card answer |
| **SM-2 Review** | `e` | Toggle explanation |
| **SM-2 Review** | `m` | Toggle star/mark on card |
| **SM-2 Review** | `u` | Unmark card |
| **SM-2 Review** | `d, f, g, h, j, k` | Grade card recall (0 = Blackout to 5 = Perfect) |
| **SM-2 Review** | `r` | Force cram / review all cards |
| **Quiz** | `a, s, d, f, g` / `1-5` | Select multiple-choice option |
| **Quiz** | `m` | Toggle star/mark on card |
| **Quiz** | `Spacebar` / `Enter` | Next question after feedback |
| **Quiz** | `r` | Retry quiz |
| **Arabic Explorer** | `Tab` | Switch Letters vs. Auxiliary/Harakat tabs |
| **Arabic Explorer** | `h, l` / `←, →` / `j, k` | Browse letters or auxiliary marks |
| **Arabic Stages** | `1-8` | Select stage |
| **Arabic Stages** | `Enter` / `d` | Start stage-specific drill |
| **Arabic Drills** | `a, s, d, f` / `1-4` | Select answer choice |
| **Arabic Drills** | `Spacebar` / `Enter` / `n` | Advance after feedback |
| **Memorize Portion** | `Tab` / `↑, ↓` | Switch boundary focus (Start ↔ End) |
| **Memorize Portion** | `h, l` / `←, →` | Adjust focused boundary |
| **Memorize Portion** | `[` / `]` | Adjust Start boundary |
| **Memorize Portion** | `{` / `}` (or `-` / `+`) | Adjust End boundary |
| **Memorize Portion** | `x` | Expand portion (+1 section) |
| **Memorize Portion** | `n` | Advance to next portion chunk |
| **Memorize Portion** | `Enter` | Begin practicing selected portion |
| **Memorize Flow** | `a, s, d, f, g` / `1-5` | Select word option |
| **Reading Reader** | `Spacebar` / `Enter` / `n` | Advance to next sentence |
| **Reading Reader** | `p` / `←` | Previous sentence |
| **Reading Reader** | `+` / `-` (or `]` / `[`) | Increase / decrease substitution aggressiveness |
| **Reading Reader** | `0` | Set 0% pure English reading mode |
| **Reading Reader** | `h` / `?` / `d` | Toggle sentence grammar & discussion box |
| **Reading Reader** | `a, s, d, z, x...` | Inspect unfamiliar substituted word |
| **Reading Word Popup** | `m` | Star word directly into Marked Cards deck |
| **Reading Word Popup** | `k` | Mark word as known |
| **Reading Word Popup** | `Esc` / `Enter` | Close popup |

---

## The SM-2 Algorithm

Corvidae uses a tuned implementation of the SuperMemo-2 (SM-2) spaced repetition algorithm:

1. **Ease Factor (EF):** Initialized at `2.5` (minimum `1.3`). Adjusted upon every review:
   $$\text{EF}' = \text{EF} + (0.1 - (5 - q) \cdot (0.08 + (5 - q) \cdot 0.02))$$
   where $q$ is the performance grade ($0 \le q \le 5$).
2. **Interval Calculation:**
   - Grade $< 3$ (Hard/Wrong/Blackout): Repetition streak resets to $0$, interval resets to $1$ day.
   - Grade $\ge 3$ (Good/Easy/Perfect):
     - Repetition $0$: Interval $= 1$ day
     - Repetition $1$: Interval $= 6$ days
     - Repetition $n \ge 2$: Interval $= \text{Round}(\text{Interval}_{n-1} \times \text{EF})$
3. **Next Review:** Calculated as $\text{Today} + \text{Interval}$ and saved to deck YAML files.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
