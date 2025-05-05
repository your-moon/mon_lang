package roundingutil

func RoundAwayFromZero(n, x int) int {
	if x%n == 0 {
		return x
	} else if x < 0 {
		return x - n - (x % n)
	} else {
		return x + n - (x % n)
	}
}
