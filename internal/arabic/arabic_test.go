package arabic

import (
	"testing"
)

func TestLettersIntegrity(t *testing.T) {
	if len(Letters) != 28 {
		t.Fatalf("expected 28 letters, got %d", len(Letters))
	}

	for i, l := range Letters {
		if l.ID != i+1 {
			t.Errorf("expected letter ID %d, got %d", i+1, l.ID)
		}
		if l.ArabicName == "" || l.EnglishName == "" {
			t.Errorf("letter %d missing name", l.ID)
		}
		if l.Forms.Isolated == "" || l.Forms.Initial == "" || l.Forms.Medial == "" || l.Forms.Final == "" {
			t.Errorf("letter %d (%s) missing one or more positional forms", l.ID, l.EnglishName)
		}
		if len(l.Examples) == 0 {
			t.Errorf("letter %d (%s) has no example words", l.ID, l.EnglishName)
		}
		for _, ex := range l.Examples {
			if ex.Arabic == "" || ex.Meaning == "" {
				t.Errorf("letter %d (%s) has incomplete example: %+v", l.ID, l.EnglishName, ex)
			}
		}
	}
}

func TestNonConnectingLetters(t *testing.T) {
	nc := NonConnectingLetters()
	if len(nc) != 6 {
		t.Fatalf("expected exactly 6 non-connecting letters, got %d", len(nc))
	}
	expectedMap := map[string]bool{
		"ا": true, // Alif
		"د": true, // Daal
		"ذ": true, // Dhaal
		"ر": true, // Raa
		"ز": true, // Zaay
		"و": true, // Waaw
	}
	for _, l := range nc {
		if !expectedMap[l.Forms.Isolated] {
			t.Errorf("unexpected non-connecting letter: %s (%s)", l.Forms.Isolated, l.EnglishName)
		}
	}
}

func TestSunAndMoonLetters(t *testing.T) {
	sun := SunLetters()
	moon := MoonLetters()

	if len(sun) != 14 {
		t.Errorf("expected 14 sun letters, got %d", len(sun))
	}
	if len(moon) != 14 {
		t.Errorf("expected 14 moon letters, got %d", len(moon))
	}

	if len(sun)+len(moon) != len(Letters) {
		t.Errorf("sum of sun (%d) and moon (%d) must equal total letters (%d)", len(sun), len(moon), len(Letters))
	}
}

func TestAuxiliaryItems(t *testing.T) {
	if len(AuxiliaryItems) == 0 {
		t.Fatal("expected non-empty auxiliary items")
	}
	for _, a := range AuxiliaryItems {
		if a.Arabic == "" || a.Name == "" || a.Description == "" {
			t.Errorf("incomplete auxiliary item: %+v", a)
		}
	}
}

func TestStages(t *testing.T) {
	if len(Stages) != 8 {
		t.Fatalf("expected 8 stages, got %d", len(Stages))
	}
	letterCount := 0
	for _, st := range Stages[:7] {
		letterCount += len(st.LetterIDs)
	}
	if letterCount != 28 {
		t.Errorf("stages 1-7 should cover all 28 letters, got %d", letterCount)
	}
}

func TestDrillGenerators(t *testing.T) {
	t.Run("PositionalForms", func(t *testing.T) {
		qs := GeneratePositionalFormQuestions(10)
		if len(qs) != 10 {
			t.Fatalf("expected 10 questions, got %d", len(qs))
		}
		for i, q := range qs {
			if q.Prompt == "" || q.Glyph == "" {
				t.Errorf("question %d missing prompt or glyph", i)
			}
			if len(q.Options) < 2 {
				t.Errorf("question %d has insufficient options: %d", i, len(q.Options))
			}
			if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
				t.Errorf("question %d correct index out of bounds: %d", i, q.CorrectIndex)
			}
		}
	})

	t.Run("LetterSounds", func(t *testing.T) {
		qs := GenerateLetterSoundQuestions(10)
		if len(qs) != 10 {
			t.Fatalf("expected 10 questions, got %d", len(qs))
		}
		for i, q := range qs {
			if q.Prompt == "" || len(q.Options) < 2 {
				t.Errorf("invalid letter sound question %d: %+v", i, q)
			}
			if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
				t.Errorf("question %d correct index out of bounds", i)
			}
		}
	})

	t.Run("Connectors", func(t *testing.T) {
		qs := GenerateConnectorQuestions(10)
		if len(qs) != 10 {
			t.Fatalf("expected 10 questions, got %d", len(qs))
		}
		for i, q := range qs {
			if q.Prompt == "" || len(q.Options) < 2 {
				t.Errorf("invalid connector question %d: %+v", i, q)
			}
			if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
				t.Errorf("question %d correct index out of bounds", i)
			}
		}
	})

	t.Run("SunMoon", func(t *testing.T) {
		qs := GenerateSunMoonQuestions(10)
		if len(qs) != 10 {
			t.Fatalf("expected 10 questions, got %d", len(qs))
		}
		for i, q := range qs {
			if q.Prompt == "" || len(q.Options) != 2 {
				t.Errorf("invalid sun moon question %d", i)
			}
			if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
				t.Errorf("question %d correct index out of bounds", i)
			}
		}
	})

	t.Run("Confusables", func(t *testing.T) {
		qs := GenerateConfusableQuestions(10)
		if len(qs) != 10 {
			t.Fatalf("expected 10 questions, got %d", len(qs))
		}
		for i, q := range qs {
			if q.Prompt == "" || len(q.Options) != 2 {
				t.Errorf("invalid confusable question %d", i)
			}
			if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
				t.Errorf("question %d correct index out of bounds", i)
			}
		}
	})

	t.Run("StageQuestions", func(t *testing.T) {
		for _, stage := range Stages {
			qs := GenerateStageQuestions(stage, 5)
			if len(qs) != 5 {
				t.Errorf("stage %d expected 5 questions, got %d", stage.Number, len(qs))
			}
			for i, q := range qs {
				if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
					t.Errorf("stage %d question %d correct index out of bounds", stage.Number, i)
				}
			}
		}
	})
}

func TestLettersConfusablesValid(t *testing.T) {
	for _, l := range Letters {
		for _, cid := range l.Confusable {
			cl := GetLetterByID(cid)
			if cl == nil {
				t.Errorf("letter %d (%s) references invalid confusable ID %d", l.ID, l.EnglishName, cid)
			}
		}
	}
}

func TestNilSafety(t *testing.T) {
	if l := GetLetterByID(0); l != nil {
		t.Errorf("expected GetLetterByID(0) to be nil, got %+v", l)
	}
	if l := GetLetterByID(99); l != nil {
		t.Errorf("expected GetLetterByID(99) to be nil, got %+v", l)
	}

	emptyStage := Stage{Number: 99, Name: "Empty", LetterIDs: []int{}}
	qs := GenerateStageQuestions(emptyStage, 5)
	if qs != nil {
		t.Errorf("expected nil questions for empty stage, got %v", qs)
	}

	invalidStage := Stage{Number: 100, Name: "Invalid", LetterIDs: []int{999}}
	qsInvalid := GenerateStageQuestions(invalidStage, 5)
	if qsInvalid != nil {
		t.Errorf("expected nil questions for invalid stage, got %v", qsInvalid)
	}
}
