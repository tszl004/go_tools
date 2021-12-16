package slice

import "github.com/pkg/errors"

func MinInt(list []int) (int, error) {
	if len(list) == 0 {
		return 0, errors.New("empty slice")
	}
	min := int(^uint(0) >> 1)
	for _, num := range list {
		if num < min {
			min = num
		}
	}
	return min, nil
}
