package console

// lengthLastWord calculates the length of the last word in a string.
// A word is defined as a sequence of non-space characters.
// Returns 0 if the string contains no words.
func lengthLastWord(str string) int {
	metLastWord := false
	metFirstChar := false
	space := uint8(' ')
	for id := len(str) - 1; id >= 0; id-- {
		if id == 0 {
			metFirstChar = true
		}
		char := str[id]
		if char != space {
			metLastWord = true
		}
		if char == space && metLastWord {
			return (len(str) - 1) - id
		}
		if metFirstChar && metLastWord {
			return len(str)
		}
	}
	return 0 // no word in this string
}

// lengthFirstWord calculates the length of the first word in a string.
// A word is defined as a sequence of non-space characters.
// Returns 0 if the string contains no words.
func lengthFirstWord(str string) int {
	metFirstWord := false
	metLastChar := false
	space := uint8(' ')
	for id := 0; id < len(str); id++ {
		if id == len(str)-1 {
			metLastChar = true
		}
		char := str[id]
		if char != space {
			metFirstWord = true
		}
		if char == space && metFirstWord {
			return id
		}
		if metLastChar && metFirstWord {
			return len(str)
		}
	}
	return 0 // no word in this string
}
