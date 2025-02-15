package utfconvert

var utfMapping = map[rune]string{
	'а': "a",
	'б': "b",
	'в': "v",
	'г': "g",
	'д': "d",
	'е': "ye",
	'ж': "j",
	'з': "z",
	'и': "i",
	'й': "hi",
	'к': "k",
	'л': "l",
	'м': "m",
	'н': "n",
	'о': "o",
	'п': "p",
	'р': "r",
	'с': "s",
	'т': "t",
	'у': "u",
	'ф': "f",
	'х': "h",
	'ц': "ts",
	'ч': "ch",
	'ш': "sh",
	'щ': "shc",
	'ъ': "qi",
	'ы': "yi",
	'ь': "zi",
	'э': "e",
	'ю': "yu",
	'я': "ya",
	' ': " ",
	'ё': "yo",
	'ү': "w",
	'ө': "q",
}

func UtfConvert(word string) string {
	var result string
	for _, char := range word {
		if converted, exists := utfMapping[char]; exists {
			result += converted
		} else {
			panic("unrecognized rune")
		}
	}
	return result
}

// - а -> 0x0430 -> a
// - б -> 0x0431 -> b
// - в -> 0x0432 -> v
// - г -> 0x0433 -> g
// - д -> 0x0434 -> d
// - е -> 0x0435 -> ye
// - ж -> 0x0436 -> j
// - з -> 0x0437 -> z
// - и -> 0x0438 -> i
// - й -> 0x0439 -> hi
// - к -> 0x043a -> k
// - л -> 0x043b -> l
// - м -> 0x043c -> m
// - н -> 0x043d -> n
// - о -> 0x043e -> o
// - п -> 0x043f -> p
// - р -> 0x0440 -> r
// - с -> 0x0441 -> s
// - т -> 0x0442 -> t
// - у -> 0x0443 -> u
// - ф -> 0x0444 -> f
// - х -> 0x0445 -> h
// - ц -> 0x0446 -> ts
// - ч -> 0x0447 -> ch
// - ш -> 0x0448 -> sh
// - щ -> 0x0449 -> shc
// - ъ -> 0x044a -> qi
// - ы -> 0x044b -> yi
// - ь -> 0x044c -> zi
// - э -> 0x044d -> e
// - ю -> 0x044e -> yu
// - я -> 0x044f -> ya
// - (space) -> 0x0020
// - ё -> 0x0451 -> yo
// - ү -> 0x04af -> w
// - ө -> 0x04e9 -> q
