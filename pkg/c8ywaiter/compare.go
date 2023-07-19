package c8ywaiter

func CompareCount(count, minimum, maximum int64) (done bool) {
	if minimum > -1 && maximum > -1 {
		done = count >= minimum && count <= maximum
	} else if minimum > -1 {
		done = count >= minimum
	} else if maximum > -1 {
		done = count <= maximum
	}
	return
}
