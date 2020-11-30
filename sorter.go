package sdiffer

type Sorter interface {
	Match(fieldPath string) bool
	Less(i, j int) bool
}

func qSort(slice []int, start, end int) {
	if start < end {
		m := slice[start]
		l := start
		r := end

		for l < r {
			for l < r && slice[r] >= m {
				r--
			}

			if l < r {
				slice[l] = slice[r]
				l++
			}

			for l < r && slice[l] <= m {
				l++
			}

			if l < r {
				slice[r] = slice[l]
				r--
			}
		}

		slice[l] = m

		qSort(slice, start, l-1)
		qSort(slice, l+1, end)
	}
}
