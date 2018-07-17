/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package pme

import (
	"strings"
	"unicode/utf8"
)

// D is the levenshtein distance calculator interface
type D interface {
	// Dist calculates levenshtein distance between two utf-8 encoded strings
	Dist(string, string) int
}

// New creates a new levenshtein distance calculator where indel is increment/deletion cost
// and sub is the substitution cost.
func New(indel, sub int) D {
	return &calculator{indel, sub}
}

type calculator struct {
	Indel_ int
	Sub_   int
}

func (calculator *calculator) Indel() int {
	return calculator.Indel_
}

func (calculator *calculator) SetIndel(Indel int) {
	calculator.Indel_ = Indel
}

func (calculator *calculator) Sub() int {
	return calculator.Sub_
}

func (calculator *calculator) SetSub(Sub int) {
	calculator.Sub_ = Sub
}

// https://en.wikibooks.org/wiki/Algorithm_Implementation/Strings/Levenshtein_distance#C
func (c *calculator) Dist(s1, s2 string) int {
	l := utf8.RuneCountInString(s1)
	m := make([]int, l+1)
	for i := 1; i <= l; i++ {
		m[i] = i * c.Indel()
	}
	lastdiag, x, y := 0, 1, 1
	for _, rx := range s2 {
		m[0], lastdiag, y = x*c.Indel(), (x-1)*c.Indel(), 1
		for _, ry := range s1 {
			m[y], lastdiag = min3(m[y]+c.Indel(), m[y-1]+c.Indel(), lastdiag+c.subCost(rx, ry)), m[y]
			y++
		}
		x++
	}
	return m[l]
}

func (c *calculator) subCost(r1, r2 rune) int {
	if r1 == r2 {
		return 0
	}
	return c.Sub()
}

func min3(a, b, c int) int {
	return min(a, min(b, c))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var defaultCalculator = New(1, 1)

// Dist is a convenience function for a levenshtein distance calculator with equal costs.
func getEDScore(s1, s2 string) int {
	return defaultCalculator.Dist(s1, s2)
}

// //function to calculate Edit distance
// func getEDScore(str1 string, str2 string, m int, n int) int {
// 	// If first string is empty, the only option is to
// 	// insert all characters of second string into first
// 	if m == 0 {
// 		return n
// 	}
// 	// If second string is empty, the only option is to
// 	// remove all characters of first string
// 	if n == 0 {
// 		return m
// 	}
// 	// If last characters of two strings are same, nothing
// 	// much to do. Ignore last characters and get count for
// 	// remaining strings.
// 	if str1[m-1] == str2[n-1] {
// 		return getEDScore(str1, str2, m-1, n-1)
// 	}
// 	// If last characters are not same, consider all three(Insert,Remove, Replace)
// 	// operations on last character of first string, recursively
// 	// compute minimum cost for all three operations and take
// 	// minimum of three values.
// 	return 1 + min(getEDScore(str1, str2, m, n-1), getEDScore(str1, str2, m-1, n), getEDScore(str1, str2, m-1, n-1))
// }
//
// func min(x int, y int, z int) int {
// 	if x < y && x < z {
// 		return x
// 	}
// 	if y < x && y < z {
// 		return y
// 	}
// 	return z
// }
//function to get sound-marker
func getPhonetic(name string) string {
	soundex := getSoundMarker(name)
	// myLogger.Debugf("Stdl2 inside phonetic return : ",soundex)
	return soundex
}

func getSoundMarker(word string) string {
	if word == "" {
		return "0000"
	}
	input := strings.ToLower(word)
	result := strings.ToUpper(input[0:1])
	code := ""
	lastCode := ""
	for _, rune := range input[1:] {
		switch rune {
		case 'b', 'f', 'p', 'v':
			code = "1"
		case 'c', 'g', 'j', 'k', 'q', 's', 'x', 'z':
			code = "2"
		case 'd', 't':
			code = "3"
		case 'l':
			code = "4"
		case 'm', 'n':
			code = "5"
		case 'r':
			code = "6"
		}
		if lastCode != code {
			lastCode = code
			result = result + lastCode
			if len(result) == 4 {
				break
			}
		}
	}
	return result + strings.Repeat("0", 4-len(result))
}

//standardization function
func Standardize(value string) string {
	// Trim spaces and special charecters
	str := trim(value)
	return str
}

func trim(value string) string {
	for _, runValue := range value {
		if runValue < 'A' || runValue > 'z' {
			value = strings.Replace(value, string(runValue), "", -1)
		}
	}
	return strings.ToUpper(value)
}

//scoring two tokens
func compareNametokens(t1 string, t2 string) int {
	if t1 != "" && t2 != "" && t1 == t2 {
		return 4
	}
	// if getEquivalant(t1, "NAME")==t2{
	// 	myLogger.Debugf("inside equivalent:")
	// 	return 3
	// }
	if getPhonetic(t1) == getPhonetic(t2) {
		return 2
	}
	ed_score := getEDScore(t1, t2)
	var base_len int
	if len(t1) > len(t2) {
		base_len = len(t1)
	} else {
		base_len = len(t2)
	}
	if float64(ed_score) <= 0.2*float64(base_len) {
		return 2
	}
	if strings.Contains(t1, t2) || strings.Contains(t2, t1) {
		return 1
	}
	return 0
}

func foldTokens(tl1 []string, tl2 []string) []string {
	temp := ""
	folded := tl2
	for i := 0; i < len(tl1); i++ {
		t1 := tl1[i]
		pos := -1
		for k := 0; k < len(tl2); k++ {
			if tl2[k] == t1 {
				pos = k
				break
			}
		}
		if pos != -1 {
			continue
		} else {
			found := false
			for j := 0; j < len(folded); j++ {
				temp = ""
				for ind := j; ind < len(folded); ind++ {
					t2 := folded[ind]
					temp = temp + t2
					if temp == t1 {
						found = true
						for ex := j; ex < ind; ex++ {
							folded = append(folded[:ex], folded[ex+1:]...)
						}
						folded[j] = temp
						break
					}
				}
				if found == true {
					break
				}
			}
		}
	}
	return folded
}

func compareNameTokenLists(tl1 []string, tl2 []string) int {
	if len(tl1) == 0 || len(tl2) == 0 {
		return 0
	}
	tl2 = foldTokens(tl1, tl2)
	tl1 = foldTokens(tl2, tl1)

	nrow := len(tl1)
	ncol := len(tl2)

	compCell := make([][]int, nrow)
	for ii := range compCell {
		compCell[ii] = make([]int, ncol)
	}
	//creating the grid of scores after comparing different tokens
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			compCell[i][j] = compareNametokens(tl1[i], tl2[j])
		}
	}
	//getting the single score from the score-grid
	max := 0
	k, r := 0, 0
	score, offset := 0, 0
	for {
		max = 0
		for i := 0; i < nrow; i++ {
			for j := 0; j < ncol; j++ {
				if compCell[i][j] > max {
					max = compCell[i][j]
					k = i
					r = j
				}
			}
		}
		if max <= 0 {
			break
		} else {
			var abs_diff int
			if k-r > 0 {
				abs_diff = k - r
			} else {
				abs_diff = r - k
			}
			if offset < abs_diff {
				offset = abs_diff
			}
			score = score + max
			for i := 0; i < nrow; i++ {
				compCell[i][r] = -1
			}
			for j := 0; j < ncol; j++ {
				compCell[k][j] = -1
			}
		}
	}

	a := float64(score) + (-0.5 * float64(offset))
	//  myLogger.Debugf("value of score: ", score)
	//  myLogger.Debugf("value of offset: ", offset)
	var temp int
	if nrow > ncol {
		temp = nrow
	} else {
		temp = ncol
	}

	b := 4 * temp
	c := (9.0 * (a / float64(b))) + 1.0
	//score out of 100
	d := int(c+0.5) * 10
	return d
}

//Main Score function
//Input strings are having tokens which are "~" separated.
func Score(str1 string, str2 string) int {
	var score int
	if str1 == "" || str2 == "" {
		score = -1
		return score
	}
	token_str1 := strings.Split(str1, "~")
	token_str2 := strings.Split(str2, "~")

	//standardization of tokens before they go for comparing.
	for i := 0; i < len(token_str1); i++ {
		token_str1[i] = Standardize(token_str1[i])
	}
	for j := 0; j < len(token_str2); j++ {
		token_str2[j] = Standardize(token_str2[j])
	}
	//calling the main scoring function
	score = compareNameTokenLists(token_str1, token_str2)
	return score
}
