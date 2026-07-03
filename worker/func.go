package main

func fMax(in []int) []*int {
	if len(in) == 0 {
		return make([]*int, 2)
	}

	if len(in) == 1 {
		return []*int{new(in[0]), new(in[0])}
	}

	max1 := &in[0]
	max2 := (*int)(nil)

	for i := 1; i < len(in); i++ {
		if *max1 < in[i] {
			max2 = max1
			max1 = &in[i]

			continue
		}

		if max2 == nil && in[i] != *max1 {
			max2 = &in[i]
			continue
		}

		if max2 != nil && *max2 < in[i] && in[i] != *max1 {
			max2 = &in[i]
		}
	}

	if max2 == nil {
		max2 = max1
	}

	return []*int{max1, max2}
}
