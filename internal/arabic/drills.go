package arabic

import (
	"fmt"
	"math/rand"
)

// DrillType identifies which category of drill is active.
type DrillType int

const (
	DrillPositionalForms DrillType = iota
	DrillLetterSounds
	DrillConnectors
	DrillSunMoon
	DrillConfusables
	DrillStage
)

// DrillQuestion represents a single question in an Arabic drill session.
type DrillQuestion struct {
	Prompt       string
	SubPrompt    string
	Glyph        string   // Main character / fragment shown prominently
	Options      []string // Answer choices
	CorrectIndex int      // 0-based index of correct option
	Explanation  string   // Pedagogical explanation shown upon answering
	LetterID     int      // Associated letter ID (if applicable)
}

// GeneratePositionalFormQuestions creates questions testing positional shapes (isolated, initial, medial, final).
func GeneratePositionalFormQuestions(count int) []DrillQuestion {
	var questions []DrillQuestion
	positions := []string{"isolated", "initial", "medial", "final"}
	posLabels := map[string]string{
		"isolated": "Isolated (منفصل)",
		"initial":  "Initial (بداية)",
		"medial":   "Medial (وسط)",
		"final":    "Final (نهاية)",
	}

	indices := rand.Perm(len(Letters))
	for i := 0; i < count; i++ {
		letter := Letters[indices[i%len(Letters)]]
		qType := rand.Intn(3)

		switch qType {
		case 0:
			// "Which is the [position] form of [letter]?"
			targetPos := positions[rand.Intn(len(positions))]
			correctForm := letter.Forms.FormAt(targetPos)

			// Collect options from other positions of same letter or other letters
			var opts []string
			seen := make(map[string]bool)
			opts = append(opts, correctForm)
			seen[correctForm] = true

			// Other positions of same letter
			for _, p := range positions {
				f := letter.Forms.FormAt(p)
				if !seen[f] {
					opts = append(opts, f)
					seen[f] = true
				}
			}
			// If still under 4 (e.g. non-connectors where initial == isolated), add from visual family or random
			for _, cid := range letter.Confusable {
				if len(opts) >= 4 {
					break
				}
				cf := GetLetterByID(cid).Forms.FormAt(targetPos)
				if !seen[cf] {
					opts = append(opts, cf)
					seen[cf] = true
				}
			}
			for len(opts) < 4 {
				rl := Letters[rand.Intn(len(Letters))]
				rf := rl.Forms.FormAt(targetPos)
				if !seen[rf] {
					opts = append(opts, rf)
					seen[rf] = true
				}
			}

			// Shuffle options
			rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
			cIdx := 0
			for idx, opt := range opts {
				if opt == correctForm {
					cIdx = idx
					break
				}
			}

			questions = append(questions, DrillQuestion{
				Prompt:       fmt.Sprintf("Which is the %s form of %s (%s)?", posLabels[targetPos], letter.ArabicName, letter.EnglishName),
				SubPrompt:    "Select the matching cursive form:",
				Glyph:        letter.Forms.Isolated,
				Options:      opts,
				CorrectIndex: cIdx,
				Explanation:  fmt.Sprintf("%s (%s): Isolated [%s], Initial [%s], Medial [%s], Final [%s]. %s", letter.ArabicName, letter.EnglishName, letter.Forms.Isolated, letter.Forms.Initial, letter.Forms.Medial, letter.Forms.Final, letter.Tips),
				LetterID:     letter.ID,
			})

		case 1:
			// "What position does this form represent for [letter]?"
			targetPos := positions[rand.Intn(len(positions))]
			formGlyph := letter.Forms.FormAt(targetPos)

			opts := []string{
				"Isolated (منفصل)",
				"Initial (بداية)",
				"Medial (وسط)",
				"Final (نهاية)",
			}
			cIdx := 0
			for idx, p := range positions {
				if p == targetPos {
					cIdx = idx
					break
				}
			}

			questions = append(questions, DrillQuestion{
				Prompt:       fmt.Sprintf("What position does this form represent for %s (%s)?", letter.ArabicName, letter.EnglishName),
				SubPrompt:    "Identify the cursive placement:",
				Glyph:        formGlyph,
				Options:      opts,
				CorrectIndex: cIdx,
				Explanation:  fmt.Sprintf("[%s] is the %s position for %s (%s). All forms: %s / %s / %s / %s.", formGlyph, posLabels[targetPos], letter.ArabicName, letter.EnglishName, letter.Forms.Isolated, letter.Forms.Initial, letter.Forms.Medial, letter.Forms.Final),
				LetterID:     letter.ID,
			})

		case 2:
			// "Which letter is this cursive fragment: [form]?"
			targetPos := positions[rand.Intn(len(positions))]
			formGlyph := letter.Forms.FormAt(targetPos)
			correctName := fmt.Sprintf("%s (%s)", letter.EnglishName, letter.ArabicName)

			var opts []string
			seen := make(map[string]bool)
			opts = append(opts, correctName)
			seen[correctName] = true

			// Distractors preferentially from confusables
			for _, cid := range letter.Confusable {
				cl := GetLetterByID(cid)
				cName := fmt.Sprintf("%s (%s)", cl.EnglishName, cl.ArabicName)
				if !seen[cName] {
					opts = append(opts, cName)
					seen[cName] = true
				}
			}
			for len(opts) < 4 {
				rl := Letters[rand.Intn(len(Letters))]
				rName := fmt.Sprintf("%s (%s)", rl.EnglishName, rl.ArabicName)
				if !seen[rName] {
					opts = append(opts, rName)
					seen[rName] = true
				}
			}

			rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
			cIdx := 0
			for idx, opt := range opts {
				if opt == correctName {
					cIdx = idx
					break
				}
			}

			questions = append(questions, DrillQuestion{
				Prompt:       fmt.Sprintf("Which letter does this %s shape belong to?", posLabels[targetPos]),
				SubPrompt:    "Identify the letter from its cursive fragment:",
				Glyph:        formGlyph,
				Options:      opts,
				CorrectIndex: cIdx,
				Explanation:  fmt.Sprintf("[%s] is the %s form of %s (%s). %s", formGlyph, posLabels[targetPos], letter.EnglishName, letter.ArabicName, letter.Tips),
				LetterID:     letter.ID,
			})
		}
	}

	return questions
}

// GenerateLetterSoundQuestions creates questions testing letter names, phonetics, and transliterations.
func GenerateLetterSoundQuestions(count int) []DrillQuestion {
	var questions []DrillQuestion
	indices := rand.Perm(len(Letters))

	for i := 0; i < count; i++ {
		letter := Letters[indices[i%len(Letters)]]
		mode := rand.Intn(2)

		if mode == 0 {
			// Show glyph, guess name & sound
			correctOpt := fmt.Sprintf("%s /%s/ (%s)", letter.EnglishName, letter.Translit, letter.ArabicName)

			var opts []string
			seen := make(map[string]bool)
			opts = append(opts, correctOpt)
			seen[correctOpt] = true

			for _, cid := range letter.Confusable {
				cl := GetLetterByID(cid)
				cOpt := fmt.Sprintf("%s /%s/ (%s)", cl.EnglishName, cl.Translit, cl.ArabicName)
				if !seen[cOpt] {
					opts = append(opts, cOpt)
					seen[cOpt] = true
				}
			}
			for len(opts) < 4 {
				rl := Letters[rand.Intn(len(Letters))]
				rOpt := fmt.Sprintf("%s /%s/ (%s)", rl.EnglishName, rl.Translit, rl.ArabicName)
				if !seen[rOpt] {
					opts = append(opts, rOpt)
					seen[rOpt] = true
				}
			}

			rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
			cIdx := 0
			for idx, opt := range opts {
				if opt == correctOpt {
					cIdx = idx
					break
				}
			}

			questions = append(questions, DrillQuestion{
				Prompt:       "What is the name and sound of this Arabic letter?",
				SubPrompt:    letter.SoundDesc,
				Glyph:        letter.Forms.Isolated,
				Options:      opts,
				CorrectIndex: cIdx,
				Explanation:  fmt.Sprintf("%s (%s) [%s]: %s. %s", letter.EnglishName, letter.ArabicName, letter.Forms.Isolated, letter.SoundDesc, letter.Tips),
				LetterID:     letter.ID,
			})
		} else {
			// Show name & sound description, select the correct Arabic letter
			correctOpt := letter.Forms.Isolated

			var opts []string
			seen := make(map[string]bool)
			opts = append(opts, correctOpt)
			seen[correctOpt] = true

			for _, cid := range letter.Confusable {
				cf := GetLetterByID(cid).Forms.Isolated
				if !seen[cf] {
					opts = append(opts, cf)
					seen[cf] = true
				}
			}
			for len(opts) < 4 {
				rf := Letters[rand.Intn(len(Letters))].Forms.Isolated
				if !seen[rf] {
					opts = append(opts, rf)
					seen[rf] = true
				}
			}

			rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
			cIdx := 0
			for idx, opt := range opts {
				if opt == correctOpt {
					cIdx = idx
					break
				}
			}

			questions = append(questions, DrillQuestion{
				Prompt:       fmt.Sprintf("Which character is %s (%s)?", letter.EnglishName, letter.ArabicName),
				SubPrompt:    fmt.Sprintf("Sound: %s (IPA: %s)", letter.SoundDesc, letter.IPA),
				Glyph:        letter.ArabicName,
				Options:      opts,
				CorrectIndex: cIdx,
				Explanation:  fmt.Sprintf("%s (%s) is written as [%s]. %s", letter.EnglishName, letter.ArabicName, letter.Forms.Isolated, letter.Tips),
				LetterID:     letter.ID,
			})
		}
	}

	return questions
}

// GenerateConnectorQuestions tests the 6 non-connecting letters versus dual-connectors.
func GenerateConnectorQuestions(count int) []DrillQuestion {
	var questions []DrillQuestion
	nonConnectors := NonConnectingLetters()
	indices := rand.Perm(len(Letters))

	for i := 0; i < count; i++ {
		letter := Letters[indices[i%len(Letters)]]
		qType := rand.Intn(2)

		if qType == 0 {
			// "Can [letter] connect to the letter that comes AFTER it (to the left)?"
			opts := []string{
				"Yes — Connects on BOTH sides (Dual-connector)",
				"No — Never connects to the left (Non-connector)",
			}
			cIdx := 0
			if !letter.IsConnector {
				cIdx = 1
			}

			ruleExplanation := "This is one of the 6 non-connecting letters (ا, د, ذ, ر, ز, و). It never connects forward to the left!"
			if letter.IsConnector {
				ruleExplanation = "This is a regular dual-connecting letter. It connects both backwards (to the right) and forwards (to the left)."
			}

			questions = append(questions, DrillQuestion{
				Prompt:       fmt.Sprintf("Can %s (%s) connect to the letter that follows it (to the left)?", letter.ArabicName, letter.EnglishName),
				SubPrompt:    "Consider Arabic cursive connection rules:",
				Glyph:        letter.Forms.Isolated,
				Options:      opts,
				CorrectIndex: cIdx,
				Explanation:  fmt.Sprintf("%s (%s): %s", letter.ArabicName, letter.EnglishName, ruleExplanation),
				LetterID:     letter.ID,
			})
		} else {
			// "Which of these letters is a Non-Connecting letter?"
			nc := nonConnectors[rand.Intn(len(nonConnectors))]
			correctOpt := fmt.Sprintf("%s (%s)", nc.Forms.Isolated, nc.EnglishName)

			var opts []string
			seen := make(map[string]bool)
			opts = append(opts, correctOpt)
			seen[correctOpt] = true

			for len(opts) < 4 {
				rl := Letters[rand.Intn(len(Letters))]
				if rl.IsConnector {
					rOpt := fmt.Sprintf("%s (%s)", rl.Forms.Isolated, rl.EnglishName)
					if !seen[rOpt] {
						opts = append(opts, rOpt)
						seen[rOpt] = true
					}
				}
			}

			rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
			cIdx := 0
			for idx, opt := range opts {
				if opt == correctOpt {
					cIdx = idx
					break
				}
			}

			questions = append(questions, DrillQuestion{
				Prompt:       "Which of these letters is a Non-Connecting (stubborn) letter?",
				SubPrompt:    "Non-connectors never connect to the letter following them (to their left):",
				Glyph:        "ا د ذ ر ز و",
				Options:      opts,
				CorrectIndex: cIdx,
				Explanation:  fmt.Sprintf("[%s] is one of the 6 non-connecting letters: ا, د, ذ, ر, ز, و. Any letter following them must begin in its isolated or initial form.", correctOpt),
				LetterID:     nc.ID,
			})
		}
	}

	return questions
}

// GenerateSunMoonQuestions creates questions testing Sun vs Moon letter assimilation.
func GenerateSunMoonQuestions(count int) []DrillQuestion {
	var questions []DrillQuestion
	indices := rand.Perm(len(Letters))

	for i := 0; i < count; i++ {
		letter := Letters[indices[i%len(Letters)]]

		opts := []string{
			"☀️ Sun Letter (حرف شمسي) — Assimilates with 'Al-' (e.g. ash-shams)",
			"🌙 Moon Letter (حرف قمري) — Clear 'l' sound (e.g. al-qamar)",
		}
		cIdx := 1
		if letter.IsSunLetter {
			cIdx = 0
		}

		explanation := fmt.Sprintf("%s (%s) is a Moon Letter (حرف قمري). When attaching 'الـ' (al-), the 'l' is clearly pronounced as 'al-'.", letter.EnglishName, letter.ArabicName)
		if letter.IsSunLetter {
			explanation = fmt.Sprintf("%s (%s) is a Sun Letter (حرف شمسي). When attaching 'الـ' (al-), the 'l' assimilates and doubles the letter with Shaddah (e.g. 'ash-', 'at-').", letter.EnglishName, letter.ArabicName)
		}

		questions = append(questions, DrillQuestion{
			Prompt:       fmt.Sprintf("Is %s (%s) a Sun Letter or a Moon Letter?", letter.ArabicName, letter.EnglishName),
			SubPrompt:    "Determine if the definite article 'الـ' assimilates:",
			Glyph:        letter.Forms.Isolated,
			Options:      opts,
			CorrectIndex: cIdx,
			Explanation:  explanation,
			LetterID:     letter.ID,
		})
	}

	return questions
}

// GenerateConfusableQuestions tests subtle visual and phonetic distinctions between confusable pairs.
func GenerateConfusableQuestions(count int) []DrillQuestion {
	var questions []DrillQuestion

	// Distinctive minimal pairs (Emphatic vs Plain, Throats, Shapes)
	pairs := [][]int{
		{3, 16},  // Taa (ت) vs Taa' (ط)
		{8, 15},  // Daal (د) vs Daad (ض)
		{12, 14}, // Seen (س) vs Saad (ص)
		{9, 17},  // Dhaal (ذ) vs Dhaa' (ظ)
		{6, 26},  // Pharyngeal Haa (ح) vs Glottal Haa (ه)
		{6, 7},   // Haa (ح) vs Khaa (خ)
		{18, 19}, // 'Ayn (ع) vs Ghayn (غ)
		{20, 21}, // Faa (ف) vs Qaaf (ق)
		{21, 22}, // Qaaf (ق) vs Kaaf (ك)
		{1, 23},  // Alif (ا) vs Laam (ل)
		{2, 25},  // Baa (ب) vs Noon (ن)
	}

	for i := 0; i < count; i++ {
		pair := pairs[i%len(pairs)]
		l1 := GetLetterByID(pair[0])
		l2 := GetLetterByID(pair[1])

		// Randomize which one is target
		target := l1
		distractor := l2
		if rand.Intn(2) == 1 {
			target = l2
			distractor = l1
		}

		opts := []string{
			fmt.Sprintf("%s (%s) — %s", target.EnglishName, target.ArabicName, target.SoundDesc),
			fmt.Sprintf("%s (%s) — %s", distractor.EnglishName, distractor.ArabicName, distractor.SoundDesc),
		}
		rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
		cIdx := 0
		targetPrefix := fmt.Sprintf("%s (%s)", target.EnglishName, target.ArabicName)
		for idx, opt := range opts {
			if len(opt) >= len(targetPrefix) && opt[:len(targetPrefix)] == targetPrefix {
				cIdx = idx
				break
			}
		}

		questions = append(questions, DrillQuestion{
			Prompt:       "Minimal Pair Challenge: Distinguish between these close Arabic letters",
			SubPrompt:    fmt.Sprintf("Identify the letter displayed below (%s vs %s):", l1.EnglishName, l2.EnglishName),
			Glyph:        target.Forms.Isolated,
			Options:      opts,
			CorrectIndex: cIdx,
			Explanation:  fmt.Sprintf("Correct! [%s] is %s (%s). Contrast: %s [%s] is %s, whereas %s [%s] is %s.", target.Forms.Isolated, target.EnglishName, target.ArabicName, target.EnglishName, target.Forms.Isolated, target.SoundDesc, distractor.EnglishName, distractor.Forms.Isolated, distractor.SoundDesc),
			LetterID:     target.ID,
		})
	}

	return questions
}

// GenerateStageQuestions creates a drill targeting only the letters within a specific learning stage.
func GenerateStageQuestions(stage Stage, count int) []DrillQuestion {
	if len(stage.LetterIDs) == 0 {
		// Stage 8: Auxiliary items
		var questions []DrillQuestion
		for i := 0; i < count; i++ {
			aux := AuxiliaryItems[i%len(AuxiliaryItems)]
			correctOpt := aux.Name

			var opts []string
			seen := make(map[string]bool)
			opts = append(opts, correctOpt)
			seen[correctOpt] = true

			for len(opts) < 4 {
				ra := AuxiliaryItems[rand.Intn(len(AuxiliaryItems))]
				if !seen[ra.Name] {
					opts = append(opts, ra.Name)
					seen[ra.Name] = true
				}
			}

			rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
			cIdx := 0
			for idx, opt := range opts {
				if opt == correctOpt {
					cIdx = idx
					break
				}
			}

			questions = append(questions, DrillQuestion{
				Prompt:       fmt.Sprintf("Stage 8 Mastery: %s", aux.Category),
				SubPrompt:    "Identify this special character or diacritic mark:",
				Glyph:        aux.Arabic,
				Options:      opts,
				CorrectIndex: cIdx,
				Explanation:  fmt.Sprintf("[%s] %s: %s (Example: %s)", aux.Arabic, aux.Name, aux.Description, aux.Example),
				LetterID:     0,
			})
		}
		return questions
	}

	var stageLetters []Letter
	for _, id := range stage.LetterIDs {
		l := GetLetterByID(id)
		if l != nil {
			stageLetters = append(stageLetters, *l)
		}
	}

	var questions []DrillQuestion
	for i := 0; i < count; i++ {
		letter := stageLetters[i%len(stageLetters)]
		correctOpt := fmt.Sprintf("%s (%s)", letter.EnglishName, letter.ArabicName)

		var opts []string
		seen := make(map[string]bool)
		opts = append(opts, correctOpt)
		seen[correctOpt] = true

		for _, sl := range stageLetters {
			sOpt := fmt.Sprintf("%s (%s)", sl.EnglishName, sl.ArabicName)
			if !seen[sOpt] {
				opts = append(opts, sOpt)
				seen[sOpt] = true
			}
		}
		for len(opts) < 4 {
			rl := Letters[rand.Intn(len(Letters))]
			rOpt := fmt.Sprintf("%s (%s)", rl.EnglishName, rl.ArabicName)
			if !seen[rOpt] {
				opts = append(opts, rOpt)
				seen[rOpt] = true
			}
		}

		rand.Shuffle(len(opts), func(a, b int) { opts[a], opts[b] = opts[b], opts[a] })
		cIdx := 0
		for idx, opt := range opts {
			if opt == correctOpt {
				cIdx = idx
				break
			}
		}

		questions = append(questions, DrillQuestion{
			Prompt:       fmt.Sprintf("Stage %d Mastery (%s)", stage.Number, stage.Name),
			SubPrompt:    fmt.Sprintf("Identify the letter: %s (IPA: %s)", letter.SoundDesc, letter.IPA),
			Glyph:        letter.Forms.Isolated,
			Options:      opts,
			CorrectIndex: cIdx,
			Explanation:  fmt.Sprintf("%s (%s): Forms [%s | %s | %s | %s]. %s", letter.EnglishName, letter.ArabicName, letter.Forms.Isolated, letter.Forms.Initial, letter.Forms.Medial, letter.Forms.Final, letter.Tips),
			LetterID:     letter.ID,
		})
	}

	return questions
}
