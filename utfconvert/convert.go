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
	'й': "y",
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
	'х': "kh",
	'ц': "ts",
	'ч': "ch",
	'ш': "sh",
	'щ': "shch",
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
			result += string(char)
		}
	}
	return result
}
