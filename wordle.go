package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
)

const wordSize = 5

type Tip interface {
	rule(string) bool
}

type Green struct {
	letter   rune
	position int
}

type Yellow struct {
	letter    rune
	positions []int
}

type Black struct {
	letter rune
}

// implements sort.Interface for []Tip based on the Tip type.
type ByTipType []Tip

func (t ByTipType) Len() int      { return len(t) }
func (t ByTipType) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

// Black < Yellow < Green
func (t ByTipType) Less(i, j int) bool {
	switch t[i].(type) {
	case Black:
		switch t[j].(type) {
		case Black:
			return true
		default:
			// i is always smaller if j is not Black
			return false
		}
	case Yellow:
		switch t[j].(type) {
		case Black:
			return true
		case Yellow:
			return true
		case Green:
			return false
		}
	case Green:
		return true
	}
	return false
}

func (g Green) String() string {
	return fmt.Sprintf("G(%c, %d)", g.letter, g.position)
}

func (y Yellow) String() string {
	return fmt.Sprintf("Y(%c, %v)", y.letter, y.positions)
}

func (b Black) String() string {
	return fmt.Sprintf("B(%c)", b.letter)
}

func (g Green) rule(word string) bool {
	if rune(word[g.position]) == g.letter {
		return true
	} else {
		return false
	}
}

func (y Yellow) rule(word string) bool {
	for i, c := range word {
		if c == y.letter && !containsInt(i, y.positions) {
			return true
		}
	}
	return false
}

func (b Black) rule(word string) bool {
	for _, c := range word {
		if c == b.letter {
			return false
		}
	}
	return true
}

func loadWords() *[]string {
	file, err := os.Open("words_en.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(file)

	initialWords := []string{}

	// TODO: support unicode chars
	isLowerAscii := func(word string) bool {
		for _, r := range word {
			if r < 'a' || r > 'z' {
				return false
			}
		}
		return true
	}

	for scanner.Scan() {
		word := scanner.Text()
		if len(word) == wordSize && isLowerAscii(word) {
			initialWords = append(initialWords, word)
		}
	}
	fmt.Printf("Starting the game with %d candidates\n", len(initialWords))
	return &initialWords
}

func filterLstStr(words *[]string, predicate func(string) bool) *[]string {
	filteredLst := []string{}
	for _, word := range *words {
		if predicate(word) {
			filteredLst = append(filteredLst, word)
		}
	}
	return &filteredLst
}

func containsInt(el int, list []int) bool {
	for _, e := range list {
		if e == el {
			return true
		}
	}
	return false
}

func containsTip(el Tip, list []Tip) bool {
	for _, e := range list {
		if reflect.DeepEqual(e, el) {
			return true
		}
	}
	return false
}

func bestScoringWord(words *[]string) (bestWord string) {
	scoreMatrix := [26][wordSize]int{}

	// compute frequence for each letter at each position
	for _, word := range *words {
		for i, c := range word {
			if c > 'z' || c > 122 {
				fmt.Println("C:", c, string(c), c-'a', string(c-'a'))
				panic("help!")
			}
			scoreMatrix[c-'a'][i] += 1
		}
	}

	// find max scoring word
	maxScore := -1
	for _, word := range *words {
		wordScore := 0
		for i, c := range word {
			wordScore += scoreMatrix[c-'a'][i]
		}
		if wordScore > maxScore {
			maxScore = wordScore
			bestWord = word
		}
	}
	// TODO: random draw from all words having max score
	return
}

func applyTips(words *[]string, tips []Tip) *[]string {
	for _, tip := range tips {
		words = filterLstStr(words, tip.rule)
		// fmt.Println("after filter: ", tip, len(*words))
	}
	return words
}

func processTips(attemptedWord []rune, existingTips []Tip) (tips []Tip) {
	// read wordle output
	var gameResponse string
	for {
		fmt.Printf("What did you get? (+ for green, * for yellow, - for black)\n")
		fmt.Println(string(attemptedWord))
		_, err := fmt.Scan(&gameResponse)
		if err == nil && len(gameResponse) == wordSize {
			break
		}
		fmt.Println("Something was wrong with your gameResponse. Try again. Got: ", gameResponse)
	}

	// create tips objects
	newTips := []Tip{}
	for i := range gameResponse {
		var newTip Tip
		switch gameResponse[i] {
		case '+':
			newTip = Green{attemptedWord[i], i}
		case '*':
			newTip = Yellow{attemptedWord[i], []int{i}}
		case '-':
			newTip = Black{attemptedWord[i]}
		}
		newTips = append(newTips, newTip)
	}

	// sort so that the list has green > yellow > black elems
	sort.Sort(ByTipType(newTips))

	// merge new tips with existing tips
	tips = existingTips
	for _, newTip := range newTips {
		if containsTip(newTip, tips) {
			continue
		}
		switch tip := newTip.(type) {
		case Green:
			tips = append(tips, tip)
		case Yellow:
			tips = append(tips, tip)
		case Black:
			// check if list already contains a green with same letter
			foundGreen := false
			for _, eTip := range tips {
				if eTipTyped, ok := eTip.(Green); ok && eTipTyped.letter == tip.letter {
					foundGreen = true
				}
			}
			// just add this rule if green was not found
			if !foundGreen {
				tips = append(tips, tip)
			}
		}
	}

	return
}

func gameLoop(words *[]string, tips []Tip) (*[]string, []Tip) {
	filteredWords := applyTips(words, tips)
	fmt.Printf("Guessing among %d words\n", len(*filteredWords))
	bestWord := bestScoringWord(filteredWords)
	fmt.Printf("Try: %s\n", bestWord)
	tips = processTips([]rune(bestWord), tips)
	return filteredWords, tips
}

func main() {
	words := loadWords()
	tips := []Tip{}
	for i := 0; i < 6; i++ {
		if len(*words) == 0 {
			fmt.Println("Was that correct?")
			break
		}
		words, tips = gameLoop(words, tips)
		fmt.Println("Current tips:", tips)
	}
}
