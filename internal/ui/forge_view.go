package ui

import (
	"fmt"
	"strings"

	"flashcards/internal/forge"

	"github.com/charmbracelet/lipgloss"
)

func suitStyled(s forge.RadicalSuit) string {
	txt := s.Icon() + " " + s.Label()
	switch s {
	case forge.SuitNature:
		return ForgeSuitNatureStyle.Render(txt)
	case forge.SuitHuman:
		return ForgeSuitHumanStyle.Render(txt)
	case forge.SuitCraft:
		return ForgeSuitCraftStyle.Render(txt)
	default:
		return ForgeSuitCoreStyle.Render(txt)
	}
}

func suitIconStyled(s forge.RadicalSuit) string {
	switch s {
	case forge.SuitNature:
		return ForgeSuitNatureStyle.Render(s.Icon())
	case forge.SuitHuman:
		return ForgeSuitHumanStyle.Render(s.Icon())
	case forge.SuitCraft:
		return ForgeSuitCraftStyle.Render(s.Icon())
	default:
		return ForgeSuitCoreStyle.Render(s.Icon())
	}
}

func (m Model) viewForgeMenu() string {
	frameWidth := m.frameWidth()
	var b strings.Builder

	syncLabel := "SM-2 Sync: OFF"
	syncPill := ModePillStyle.Render(syncLabel)
	if m.isForgeSM2SyncEnabled() {
		syncPill = SyncOnPillStyle.Render("SM-2 Sync: ON")
	}

	b.WriteString(renderSplitLine(
		ForgeTitleStyle.Render("⚒ MORPHEME FORGE — MANDARIN ROGUELIKE ⚒"),
		syncPill,
		frameWidth,
	) + "\n")
	b.WriteString(HintStyle.Render("Forge Hanzi compounds & Pivot Chains • Harry Potter & Gospel of John Lexicon") + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	if m.ForgeProgress != nil && m.ForgeLexicon != nil {
		mastered := 0
		for _, item := range m.ForgeProgress.Compounds {
			if item.Reps > 0 {
				mastered++
			}
		}
		statsLine := fmt.Sprintf("Lexicon: %d Hanzi • %d Compounds   |   Runs: %d (Wins: %d)   |   Best Hand: %d   |   Forged: %d",
			len(m.ForgeLexicon.Hanzi), len(m.ForgeLexicon.Compounds),
			m.ForgeProgress.RunsPlayed, m.ForgeProgress.RunsWon,
			m.ForgeProgress.BestScore, mastered)
		b.WriteString(DashboardPillStyle.Render(truncateToWidth(statsLine, frameWidth-4)) + "\n\n")
	}

	b.WriteString(SectionTitleStyle.Render("Select Starter Deck") + "\n\n")
	for i, arch := range forge.StarterArchetypes {
		key := IndexToKey(i)
		cursor := "  "
		title := arch.Name
		if m.ForgeMenuCursor == i {
			cursor = "> "
			title = CursorStyle.Render(arch.Name)
		}
		b.WriteString(fmt.Sprintf("%s%s  %s  %s\n",
			CursorStyle.Render(cursor),
			KeyStyle.Render("["+key+"]"),
			title,
			ForgeCoinStyle.Render("("+arch.Passive+")"),
		))
		b.WriteString(fmt.Sprintf("        %s\n\n", HintStyle.Render(arch.Desc)))
	}

	b.WriteString(renderDivider(frameWidth) + "\n")
	b.WriteString(HintStyle.Render("SM-2 Card Physics:  ⚡ Glass (New/Leech = xMult, shatters on miss)  •  ◆ Foil (Learning)  •  ★ Gold (Mature = +$)") + "\n")
	footer := fmt.Sprintf("[%s] Start Deck  •  [%s] Toggle Main-Deck SM-2 Sync  •  [%s] Main Menu",
		KeyStyle.Render("a-f / Enter"), KeyStyle.Render("S"), KeyStyle.Render("Esc"))
	b.WriteString(HintStyle.Render(footer))
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewForgePlay() string {
	run := m.ForgeRun
	if run == nil {
		return m.viewForgeMenu()
	}
	frameWidth := m.frameWidth()
	var b strings.Builder

	// Top HUD: Ante / Blind / Target / Coins
	anteTitle := fmt.Sprintf("ANTE %d/%d • %s", run.Ante, forge.WinAnte, run.BlindKind.Name())
	if run.EndlessMode {
		anteTitle = fmt.Sprintf("ENDLESS ANTE %d • %s", run.Ante, run.BlindKind.Name())
	}
	rightHUD := fmt.Sprintf("%s  %s  %s",
		ForgeChipsStyle.Render(fmt.Sprintf("Hands:%d", run.HandsLeft)),
		HintStyle.Render(fmt.Sprintf("Discards:%d", run.DiscardsLeft)),
		ForgeCoinStyle.Render(fmt.Sprintf("$%d", run.Money)),
	)
	b.WriteString(renderSplitLine(ForgeTitleStyle.Render(anteTitle), rightHUD, frameWidth) + "\n")

	// Boss Blind banner if active
	if run.ActiveBoss != forge.BossNone {
		if bd, ok := forge.GetBossBlind(run.ActiveBoss); ok {
			extra := bd.Desc
			if run.ActiveBoss == forge.BossSilencedSuit && run.SilencedSuit != "" {
				extra = fmt.Sprintf("Silenced Suit: %s (0 Chips)", run.SilencedSuit.Label())
			}
			b.WriteString(ForgeBossPillStyle.Render("BOSS: "+bd.Name) + " " + ErrorStyle.Render(extra) + "\n")
		}
	}

	// Score progress bar
	pct := 0
	if run.TargetScore > 0 {
		pct = (run.BlindScore * 100) / run.TargetScore
		if pct > 100 {
			pct = 100
		}
	}
	barWidth := 24
	filled := (pct * barWidth) / 100
	bar := CorrectStyle.Render(strings.Repeat("█", filled)) + HintStyle.Render(strings.Repeat("░", barWidth-filled))
	scoreLine := fmt.Sprintf("Score: %s / %s  [%s] %d%%   Deck: %d",
		ForgeChipsStyle.Render(fmt.Sprintf("%d", run.BlindScore)),
		BreadcrumbStyle.Render(fmt.Sprintf("%d", run.TargetScore)),
		bar, pct, len(run.DrawPile))
	b.WriteString(scoreLine + "\n")

	// Relics bar
	if len(run.Relics) > 0 {
		var pills []string
		for _, rid := range run.Relics {
			if def, ok := forge.GetRelic(rid); ok {
				pills = append(pills, ForgeRelicPillStyle.Render(def.Name))
			}
		}
		b.WriteString("Relics: " + strings.Join(pills, " ") + "\n")
	} else {
		b.WriteString(HintStyle.Render("Relics: (none yet — visit Shop after clearing Blind)") + "\n")
	}
	b.WriteString(renderDivider(frameWidth) + "\n")

	// Optional Reference Overlay
	if m.ForgeShowInfo {
		b.WriteString(m.renderForgeInfoPanel(frameWidth))
		b.WriteString("\n" + renderDivider(frameWidth) + "\n")
		b.WriteString(HintStyle.Render(fmt.Sprintf("[%s] Close Reference", KeyStyle.Render("v / Esc"))))
		return renderFrame(frameWidth, b.String())
	}

	// Scoring Breakdown Overlay after a Play completes
	if m.ForgeShowBreakdown && run.LastPlayResult != nil {
		b.WriteString(m.renderForgeBreakdown(run, frameWidth))
		return renderFrame(frameWidth, b.String())
	}

	// Staged Play Preview (no validity spoilers!)
	var stageParts []string
	for i, sw := range run.Staged {
		var w strings.Builder
		for _, ci := range sw.CardIndices {
			if ci >= 0 && ci < len(run.Hand) {
				w.WriteString(run.Hand[ci].Glyph)
			}
		}
		slotText := w.String()
		if slotText == "" {
			slotText = "_"
		}
		if i > 0 {
			if sw.IsPivot {
				stageParts = append(stageParts, HotStreakStyle.Render("⇄"))
			} else {
				stageParts = append(stageParts, HintStyle.Render("+"))
			}
		}
		stageParts = append(stageParts, CharStyle.Render("[ "+slotText+" ]"))
	}
	stageHeader := "STAGED PLAY:  " + strings.Join(stageParts, " ")
	b.WriteString("\n" + ForgeStageBoxStyle.Width(frameWidth-4).Render(stageHeader) + "\n\n")

	// 9-Card Hanzi Hand rendered in two tidy rows (5 + 4)
	renderCardTile := func(i int) string {
		card := run.Hand[i]
		key := IndexToKey(i)
		h := run.Lexicon.CharsByGlyph[card.Glyph]
		suitStr := ""
		toneStr := ""
		if h != nil {
			suitStr = suitIconStyled(h.Suit)
			toneStr = fmt.Sprintf("T%d", h.Tone)
		}
		mat := run.Progress.MaterialForChar(card.Glyph)
		matBadge := mat.Badge()
		box := ForgeCardGlassBox
		if mat == forge.MaterialFoil {
			box = ForgeCardFoilBox
		} else if mat == forge.MaterialGold {
			box = ForgeCardGoldBox
		}
		if run.ActiveBoss == forge.BossBlindMaterial {
			matBadge = "?"
			box = ForgeCardFoilBox
		}
		if run.IsCardStaged(i) {
			box = ForgeCardStagedBox
		}
		edTag := ""
		if card.Edition != forge.EditionNone {
			edTag = " " + ForgeCoinStyle.Render(card.Edition.Label())
		}
		line1 := fmt.Sprintf("%s %s %s", KeyStyle.Render("["+key+"]"), CharStyle.Render(card.Glyph), matBadge)
		line2 := fmt.Sprintf("%s %s%s", suitStr, HintStyle.Render(toneStr), edTag)
		return box.Width(13).Render(line1 + "\n" + line2)
	}

	var row1, row2 []string
	for i := 0; i < len(run.Hand); i++ {
		if i < 5 {
			row1 = append(row1, renderCardTile(i))
		} else {
			row2 = append(row2, renderCardTile(i))
		}
	}
	if len(row1) > 0 {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row1...) + "\n")
	}
	if len(row2) > 0 {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row2...) + "\n")
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	footer1 := fmt.Sprintf("[%s] Stage Hanzi  •  [%s] New Word  •  [%s] Pivot Chain ⇄  •  [%s] Undo",
		KeyStyle.Render("a-l"), KeyStyle.Render("Space"), KeyStyle.Render("Tab"), KeyStyle.Render("Bksp"))
	footer2 := fmt.Sprintf("[%s] Play & Prove  •  [%s] Discard Staged  •  [%s] Hand Levels/Relics  •  [%s] Clear/Menu",
		KeyStyle.Render("Enter"), KeyStyle.Render(";"), KeyStyle.Render("v"), KeyStyle.Render("Esc"))
	b.WriteString(HintStyle.Render(footer1) + "\n" + HintStyle.Render(footer2))
	return renderFrame(frameWidth, b.String())
}

func (m Model) renderForgeBreakdown(run *forge.Run, frameWidth int) string {
	res := run.LastPlayResult
	var b strings.Builder

	header := fmt.Sprintf("PLAY RESULT: %s (Lv.%d)", res.HandType.Name(), res.HandLevel)
	if res.HandType == forge.HandNone {
		header = "PLAY FIZZLED!"
	}
	b.WriteString("\n" + SectionTitleStyle.Render(header) + "\n\n")

	// Per-word outcomes
	for _, wo := range res.Words {
		if !wo.Valid {
			b.WriteString(fmt.Sprintf("  %s %s — %s\n", ErrorStyle.Render("✗"), CharStyle.Render(wo.Word), ErrorStyle.Render("Not in lexicon (Fizzled)")))
		} else if wo.Passed {
			extra := ""
			if wo.ShieldSaved {
				extra = " " + CorrectStyle.Render("[Forgiving Tutor!]")
			}
			b.WriteString(fmt.Sprintf("  %s %s (%s) — %s%s\n",
				CorrectStyle.Render("✓"),
				CharStyle.Render(wo.Word),
				PinyinStyle.Render(wo.Pinyin),
				MeaningStyle.Render(wo.Meaning),
				extra,
			))
		} else {
			shatter := ""
			if wo.GlassShattered {
				shatter = " " + ErrorStyle.Render("[GLASS SHATTERED!]")
			}
			b.WriteString(fmt.Sprintf("  %s %s (%s) — %s%s\n",
				ErrorStyle.Render("✗"),
				CharStyle.Render(wo.Word),
				PinyinStyle.Render(wo.Pinyin),
				MeaningStyle.Render(wo.Meaning),
				shatter,
			))
		}
	}

	b.WriteString("\n")
	for _, ev := range res.Events {
		b.WriteString(fmt.Sprintf("  • %-28s %s\n", ev.Label, HintStyle.Render(ev.Detail)))
	}

	b.WriteString("\n")
	totalLine := fmt.Sprintf("  %s  ×  %s  =  %s",
		ForgeChipsStyle.Render(fmt.Sprintf("%d Chips", res.Chips)),
		ForgeMultStyle.Render(fmt.Sprintf("%.1f Mult", res.Mult)),
		CorrectStyle.Render(fmt.Sprintf("+%d pts", res.TotalScore)),
	)
	if res.CoinsEarned > 0 {
		totalLine += "   " + ForgeCoinStyle.Render(fmt.Sprintf("(+$%d)", res.CoinsEarned))
	}
	b.WriteString(totalLine + "\n\n")

	if len(res.MissedHints) > 0 {
		b.WriteString(SectionTitleStyle.Render("💡 Also Available in That Hand:") + "\n")
		for _, hint := range res.MissedHints {
			if hint.IsPivot {
				b.WriteString(fmt.Sprintf("  %s\n    %s\n",
					HotStreakStyle.Render(hint.Label),
					HintStyle.Render(truncateToWidth(hint.Detail, frameWidth-8)),
				))
			} else {
				b.WriteString(fmt.Sprintf("  • %s — %s\n",
					PinyinStyle.Render(hint.Label),
					HintStyle.Render(truncateToWidth(hint.Detail, frameWidth-34)),
				))
			}
		}
		b.WriteString("\n")
	}

	if run.BlindCleared() {
		if run.WonRun && run.Ante == forge.WinAnte && run.BlindKind == forge.BlindBoss {
			b.WriteString(CorrectStyle.Render("★ ANTE 6 BOSS DEFEATED — YOU WON THE RUN! (Press Enter to enter Shop & Endless Mode) ★") + "\n")
		} else {
			b.WriteString(CorrectStyle.Render("★ BLIND CLEARED! Press Enter / Space to visit the Shop. ★") + "\n")
		}
	} else if run.GameOver {
		b.WriteString(ErrorStyle.Render("Out of Hands! Press Enter / Space for Run Summary.") + "\n")
	} else {
		b.WriteString(HintStyle.Render(fmt.Sprintf("[%s] Continue to Next Hand", KeyStyle.Render("Enter / Space"))))
	}
	return b.String()
}

func (m Model) renderForgeInfoPanel(frameWidth int) string {
	run := m.ForgeRun
	var b strings.Builder
	b.WriteString("\n" + SectionTitleStyle.Render("Hand Levels") + "\n")
	for _, ht := range forge.AllHandTypes {
		lvl := run.HandLevels[ht]
		c, mu := ht.StatsAtLevel(lvl)
		b.WriteString(fmt.Sprintf("  %-16s Lv.%d   %s × %s\n",
			ht.Name(), lvl,
			ForgeChipsStyle.Render(fmt.Sprintf("%d", c)),
			ForgeMultStyle.Render(fmt.Sprintf("%.0f", mu)),
		))
	}
	b.WriteString("\n" + SectionTitleStyle.Render("Held Relics") + "\n")
	if len(run.Relics) == 0 {
		b.WriteString("  " + HintStyle.Render("(none)") + "\n")
	}
	for _, rid := range run.Relics {
		if def, ok := forge.GetRelic(rid); ok {
			b.WriteString(fmt.Sprintf("  • %s — %s\n", ForgeCoinStyle.Render(def.Name), HintStyle.Render(def.Desc)))
		}
	}
	return b.String()
}

func (m Model) viewForgeProve() string {
	run := m.ForgeRun
	if run == nil {
		return m.viewForgeMenu()
	}
	frameWidth := m.frameWidth()
	var b strings.Builder

	total := len(run.Challenges)
	idx := run.ProveIdx
	if idx >= total {
		idx = total - 1
	}

	b.WriteString(renderSplitLine(
		ForgeTitleStyle.Render("⚡ PROVE YOUR FORGE ⚡"),
		ModePillStyle.Render(fmt.Sprintf("Word %d / %d", idx+1, total)),
		frameWidth,
	) + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	if m.ForgeMissPause {
		b.WriteString(ErrorStyle.Render("Recall Missed!") + "\n\n")
		b.WriteString("  " + BreadcrumbStyle.Render(run.LastProveMsg) + "\n\n")
		b.WriteString(HintStyle.Render("Take a second to lock it in, then press [Enter / Space] to continue."))
		return renderFrame(frameWidth, b.String())
	}

	if run.LastProveMsg != "" && !run.LastProveFailed {
		b.WriteString(CorrectStyle.Render(run.LastProveMsg) + "\n\n")
	}

	ch := run.Challenges[run.ProveIdx]
	badge := ""
	if ch.IsVolatileCompound {
		badge = "  " + ForgeTitleStyle.Render("⚡ VOLATILE DISCOVERY")
	}
	promptBox := QuizPromptBoxStyle.Width(frameWidth - 4).Render(
		fmt.Sprintf("%s%s\n%s",
			CharStyle.Render(ch.Word),
			badge,
			HintStyle.Render(ch.Mode.Label()),
		),
	)
	b.WriteString(promptBox + "\n\n")

	for i, opt := range ch.Options {
		key := IndexToKey(i)
		b.WriteString(fmt.Sprintf("  %s %s  %s\n",
			KeyStyle.Render("["+key+"]"),
			HintStyle.Render(fmt.Sprintf("(%d)", i+1)),
			MeaningStyle.Render(opt),
		))
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	b.WriteString(HintStyle.Render(fmt.Sprintf("[%s] Select Answer", KeyStyle.Render("a-g / 1-5"))))
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewForgeShop() string {
	run := m.ForgeRun
	if run == nil {
		return m.viewForgeMenu()
	}
	frameWidth := m.frameWidth()
	var b strings.Builder

	b.WriteString(renderSplitLine(
		ForgeTitleStyle.Render("★ FORGE SHOP ★"),
		ForgeCoinStyle.Render(fmt.Sprintf("Coins: $%d   |   Deck: %d   |   Relics: %d/%d", run.Money, len(run.FullDeck), len(run.Relics), forge.MaxRelics)),
		frameWidth,
	) + "\n")
	if m.ForgeLastReward > 0 {
		b.WriteString(CorrectStyle.Render(fmt.Sprintf("Blind Reward Earned: +$%d", m.ForgeLastReward)) + "\n")
	}
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	// Relics
	b.WriteString(SectionTitleStyle.Render("Relics (Jokers)") + "\n")
	relicKeys := []string{"a", "s"}
	for i, offer := range run.Shop.Relics {
		k := relicKeys[i]
		status := ForgeCoinStyle.Render(fmt.Sprintf("$%d", offer.Def.Cost))
		if offer.Bought {
			status = HintStyle.Render("SOLD")
		}
		b.WriteString(fmt.Sprintf("  %s  %-18s %s  %s\n",
			KeyStyle.Render("["+k+"]"),
			BreadcrumbStyle.Render(offer.Def.Name),
			status,
			HintStyle.Render(offer.Def.Desc),
		))
	}

	// Hanzi Draft Pack
	b.WriteString("\n" + SectionTitleStyle.Render(fmt.Sprintf("Hanzi Draft ($%d — adds to your deck)", forge.DraftCost)) + "\n")
	draftKeys := []string{"d", "f", "g"}
	for i, opt := range run.Shop.Draft {
		k := draftKeys[i]
		h := run.Lexicon.CharsByGlyph[opt.Glyph]
		mat := run.Progress.MaterialForChar(opt.Glyph)
		info := ""
		if h != nil {
			info = fmt.Sprintf("%s  T%d  %s %s", suitStyled(h.Suit), h.Tone, mat.Badge(), mat.Label())
		}
		status := CorrectStyle.Render(fmt.Sprintf("Unlocks %d compounds", opt.UnlocksCount))
		if opt.Bought {
			status = HintStyle.Render("DRAFTED")
		}
		b.WriteString(fmt.Sprintf("  %s  %s   %-26s  %s\n",
			KeyStyle.Render("["+k+"]"),
			CharStyle.Render(opt.Glyph),
			info,
			status,
		))
	}

	// Scroll & Sculpting
	b.WriteString("\n" + SectionTitleStyle.Render("Scrolls & Deck Sculpting") + "\n")
	scrollStatus := ForgeCoinStyle.Render(fmt.Sprintf("$%d", run.Shop.Scroll.Cost))
	if run.Shop.Scroll.Bought {
		scrollStatus = HintStyle.Render("USED")
	}
	b.WriteString(fmt.Sprintf("  %s  Upgrade %-14s %s  %s\n",
		KeyStyle.Render("[h]"),
		run.Shop.Scroll.HandType.Name(),
		scrollStatus,
		HintStyle.Render(fmt.Sprintf("(Lv.%d → Lv.%d)", run.HandLevels[run.Shop.Scroll.HandType], run.HandLevels[run.Shop.Scroll.HandType]+1)),
	))

	pruneStatus := ForgeCoinStyle.Render(fmt.Sprintf("$%d", forge.PruneCost))
	if run.Shop.PrunedThisShop {
		pruneStatus = HintStyle.Render("DONE")
	}
	b.WriteString(fmt.Sprintf("  %s  Prune Weakest Hanzi    %s  %s\n",
		KeyStyle.Render("[j]"), pruneStatus, HintStyle.Render("Remove lowest-connectivity character")))

	dupeStatus := ForgeCoinStyle.Render(fmt.Sprintf("$%d", forge.DuplicateCost))
	if run.Shop.DupedThisShop {
		dupeStatus = HintStyle.Render("DONE")
	}
	b.WriteString(fmt.Sprintf("  %s  Duplicate Best Hub     %s  %s\n",
		KeyStyle.Render("[k]"), dupeStatus, HintStyle.Render("Copy top hub Hanzi with Foil+/Holo/Poly edition")))

	if run.Shop.StatusMessage != "" {
		b.WriteString("\n" + QuizExplanationStyle.Render(run.Shop.StatusMessage) + "\n")
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	footer := fmt.Sprintf("[%s] Buy/Draft  •  [%s] Reroll ($%d)  •  [%s] Next Blind  •  [%s] Menu",
		KeyStyle.Render("a-k"), KeyStyle.Render("r"), forge.RerollCost, KeyStyle.Render("Enter / Space"), KeyStyle.Render("Esc"))
	b.WriteString(HintStyle.Render(footer))
	return renderFrame(frameWidth, b.String())
}

func (m Model) viewForgeRunOver() string {
	run := m.ForgeRun
	frameWidth := m.frameWidth()
	var b strings.Builder

	title := ErrorStyle.Render("RUN OVER")
	if run != nil && run.WonRun {
		title = CorrectStyle.Render("★ VICTORY — MORPHEME MASTER ★")
	}
	b.WriteString(title + "\n")
	b.WriteString(renderDivider(frameWidth) + "\n\n")

	if run != nil {
		b.WriteString(fmt.Sprintf("  Starter Deck:   %s\n", BreadcrumbStyle.Render(run.ArchetypeName)))
		b.WriteString(fmt.Sprintf("  Reached:        Ante %d • %s\n", run.Ante, run.BlindKind.Name()))
		b.WriteString(fmt.Sprintf("  Best Hand:      %s\n", ForgeChipsStyle.Render(fmt.Sprintf("%d pts", run.BestHandScore))))
		b.WriteString(fmt.Sprintf("  Words Proven:   %s\n", CorrectStyle.Render(fmt.Sprintf("%d", run.WordsForged))))
		b.WriteString(fmt.Sprintf("  Final Deck:     %d Hanzi, %d Relics\n", len(run.FullDeck), len(run.Relics)))
	}

	b.WriteString("\n" + renderDivider(frameWidth) + "\n")
	b.WriteString(HintStyle.Render(fmt.Sprintf("[%s] Play Again  •  [%s] Forge Menu",
		KeyStyle.Render("r"), KeyStyle.Render("Enter / Esc"))))
	return renderFrame(frameWidth, b.String())
}
