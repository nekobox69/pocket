// Package kit Create at 2020-11-06 10:20
package pocket

// Pow 次方
func Pow(x, n int64) int64 {
	ret := int64(1) // 结果初始为0次方的值，整数0次方为1。如果是矩阵，则为单元矩阵。
	for n != 0 {
		if n%2 != 0 {
			ret = ret * x
		}
		n /= 2
		x = x * x
	}
	return ret
}

// IntMax int max
func IntMax(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// IntMin int min
func IntMin(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

// FloatMax float max
func FloatMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// FloatMin float min
func FloatMin(a, b float64) float64 {
	if a > b {
		return b
	}
	return a
}
