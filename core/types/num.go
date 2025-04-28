package types

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

func AbsInt(num int) int {
	if num >= 0 {
		return num
	}

	return num * -1
}

func SumInt(nums ...int) int {
	var res int
	for _, v := range nums {
		res += v
	}

	return res
}

func MinInt(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}

	min := nums[0]
	for _, v := range nums {
		if min > v {
			min = v
		}
	}

	return min
}

func MaxInt(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}

	max := nums[0]
	for _, v := range nums {
		if max < v {
			max = v
		}
	}

	return max
}

func FindKthLargestInt(nums []int, k int) int {
	l := len(nums)
	if k < 1 || k > l {
		return 0
	}

	pivot := nums[0]
	left := 0
	right := len(nums) - 1

	for i := 1; i <= right; {
		if nums[i] < pivot {
			nums[left], nums[i] = nums[i], nums[left]
			i++
			left++
		} else if nums[i] > pivot {
			nums[right], nums[i] = nums[i], nums[right]
			right--
		} else {
			i++
		}
	}

	if k <= left {
		return FindKthLargestInt(nums[:left], k)
	} else if k > right+1 {
		return FindKthLargestInt(nums[right+1:], k-right-1)
	} else {
		return pivot
	}
}

func AbsInt32(num int32) int32 {
	if num >= 0 {
		return num
	}

	return num * -1
}

func SumInt32(nums ...int32) int32 {
	var res int32
	for _, v := range nums {
		res += v
	}

	return res
}

func MinInt32(nums ...int32) int32 {
	if len(nums) == 0 {
		return 0
	}

	min := nums[0]
	for _, v := range nums {
		if min > v {
			min = v
		}
	}

	return min
}

func MaxInt32(nums ...int32) int32 {
	if len(nums) == 0 {
		return 0
	}

	max := nums[0]
	for _, v := range nums {
		if max < v {
			max = v
		}
	}

	return max
}

func FindKthLargestInt32(nums []int32, k int) int32 {
	l := len(nums)
	if k < 1 || k > l {
		return 0
	}

	pivot := nums[0]
	left, right := 0, len(nums)-1

	for i := 1; i <= right; {
		if nums[i] < pivot {
			nums[left], nums[i] = nums[i], nums[left]
			i++
			left++
		} else if nums[i] > pivot {
			nums[right], nums[i] = nums[i], nums[right]
			right--
		} else {
			i++
		}
	}

	if k <= left {
		return FindKthLargestInt32(nums[:left], k)
	} else if k > right+1 {
		return FindKthLargestInt32(nums[right+1:], k-right-1)
	} else {
		return pivot
	}
}

func AbsInt64(num int64) int64 {
	if num >= 0 {
		return num
	}

	return num * -1
}

func SumInt64(nums ...int64) int64 {
	var res int64
	for _, v := range nums {
		res += v
	}

	return res
}

func MinInt64(nums ...int64) int64 {
	if len(nums) == 0 {
		return 0
	}

	min := nums[0]
	for _, v := range nums {
		if min > v {
			min = v
		}
	}

	return min
}

func MaxInt64(nums ...int64) int64 {
	if len(nums) == 0 {
		return 0
	}

	max := nums[0]
	for _, v := range nums {
		if max < v {
			max = v
		}
	}

	return max
}

func FindKthLargestInt64(nums []int64, k int) int64 {
	l := len(nums)
	if k < 1 || k > l {
		return 0
	}

	pivot := nums[0]
	left, right := 0, len(nums)-1

	for i := 1; i <= right; {
		if nums[i] < pivot {
			nums[left], nums[i] = nums[i], nums[left]
			i++
			left++
		} else if nums[i] > pivot {
			nums[right], nums[i] = nums[i], nums[right]
			right--
		} else {
			i++
		}
	}

	if k <= left {
		return FindKthLargestInt64(nums[:left], k)
	} else if k > right+1 {
		return FindKthLargestInt64(nums[right+1:], k-right-1)
	} else {
		return pivot
	}
}

func AbsFloat32(num float32) float32 {
	if num >= 0 {
		return num
	}

	return num * -1
}

func SumFloat32(nums ...float32) float32 {
	var res float32
	for _, v := range nums {
		res += v
	}

	return res
}

func MinFloat32(nums ...float32) float32 {
	if len(nums) == 0 {
		return 0
	}

	min := nums[0]
	for _, v := range nums {
		if min > v {
			min = v
		}
	}

	return min
}

func MaxFloat32(nums ...float32) float32 {
	if len(nums) == 0 {
		return 0
	}

	max := nums[0]
	for _, v := range nums {
		if max < v {
			max = v
		}
	}

	return max
}

func FindKthLargestFloat32(nums []float32, k int) float32 {
	l := len(nums)
	if k < 1 || k > l {
		return 0
	}

	pivot := nums[0]
	left, right := 0, len(nums)-1

	for i := 1; i <= right; {
		if nums[i] < pivot {
			nums[left], nums[i] = nums[i], nums[left]
			i++
			left++
		} else if nums[i] > pivot {
			nums[right], nums[i] = nums[i], nums[right]
			right--
		} else {
			i++
		}
	}

	if k <= left {
		return FindKthLargestFloat32(nums[:left], k)
	} else if k > right+1 {
		return FindKthLargestFloat32(nums[right+1:], k-right-1)
	} else {
		return pivot
	}
}

func AbsFloat64(num float64) float64 {
	return math.Abs(num)
}

func IsZeroFloat(v float64) bool {
	return math.Abs(v) < epsilon
}

func IsEqFloat(v1, v2 float64) bool {
	return IsZeroFloat(AbsFloat64(v1 - v2))
}

func Truncate(num ...*float64) {
	for _, v := range num {
		if AbsFloat64(*v) < epsilon {
			*v = 0
		}
	}
}

func SumFloat64(nums ...float64) float64 {
	var res float64
	for _, v := range nums {
		res += v
	}

	return res
}

func MinFloat64(nums ...float64) float64 {
	if len(nums) == 0 {
		return 0
	}

	min := nums[0]
	for _, v := range nums {
		if min > v {
			min = v
		}
	}

	return min
}

func MaxFloat64(nums ...float64) float64 {
	if len(nums) == 0 {
		return 0
	}

	max := nums[0]
	for _, v := range nums {
		if max < v {
			max = v
		}
	}

	return max
}

func FindKthLargestFloat64(nums []float64, k int) float64 {
	l := len(nums)
	if k < 1 || k > l {
		return 0
	}

	pivot := nums[0]
	left, right := 0, l-1

	for i := 1; i <= right; {
		if nums[i] < pivot {
			nums[left], nums[i] = nums[i], nums[left]
			i++
			left++
		} else if nums[i] > pivot {
			nums[right], nums[i] = nums[i], nums[right]
			right--
		} else {
			i++
		}
	}

	if k <= left {
		return FindKthLargestFloat64(nums[:left], k)
	} else if k > right+1 {
		return FindKthLargestFloat64(nums[right+1:], k-right-1)
	} else {
		return pivot
	}
}

func AvgFloat64(nums []float64) float64 {
	l := len(nums)
	if l == 0 {
		return math.NaN()
	}

	return SumFloat64(nums...) / float64(l)
}

func VarFloat64(nums []float64, mean float64) float64 {
	l := len(nums)
	if l == 0 {
		return math.NaN()
	}

	sum := 0.0
	for _, v := range nums {
		sum += math.Pow(v-mean, 2)
	}

	return sum / float64(l)
}

func IsSmoothSeq(seq ...float64) bool {
	mean := AvgFloat64(seq)
	if mean == 0 {
		return false
	}

	variance := VarFloat64(seq, mean)
	return math.Abs(variance/mean) < 0.1
}

func CalcRate(num, total int, di int) string {
	if total == 0 {
		return ""
	}

	tmp := num * 100 / total

	return fmt.Sprint(tmp) + "%"
}

//随机数种子
var Rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}

	return Rnd.Intn(max-min) + min
}

func RandomInt32(min, max int32) int32 {
	if min >= max || min == 0 || max == 0 {
		return max
	}

	return Rnd.Int31n(max-min) + min
}

func RandomInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}

	return Rnd.Int63n(max-min) + min
}

func RandomFloat64(min, max float64) float64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}

	return Rnd.Float64()*(max-min) + min
}

func RandomRatio(min, max int) float64 {
	return float64(RandomInt(min, max)) / 100
}

func Dice20() int {
	return RandomInt(1, 21)
}

func Dice100() int {
	return RandomInt(1, 101)
}

// 获取随机数
func GetRandNumber(n int) int {
	return Rnd.Intn(n)
}

const base90Chars = "0123456789" +
	"abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"!@#$%^&*()" +
	"[]{}<>+-/=_,.:;?'~"

func DecimalToBase90(decimal int) string {
	if decimal == 0 {
		return string(base90Chars[0])
	}

	var result []byte
	dividend := decimal

	for dividend > 0 {
		remainder := dividend % 90
		result = append([]byte{base90Chars[remainder]}, result...)
		dividend = dividend / 90
	}

	return string(result)
}

func Base90ToDecimal(number string) int {
	res := 0
	for _, char := range number {
		digit := strings.IndexRune(base90Chars, char)
		if digit > -1 {
			res = res*90 + digit
		}
	}

	return res
}

func Decimal(value float64) float64 {
	return math.Trunc(value*1e2+0.5) * 1e-2
}

func Round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

func TakeDigits(num float64, decimal int) float64 {
	if decimal == 0 {
		return num
	}

	shift := math.Pow10(decimal)
	return math.Round(num*shift) / shift
}

func TruncDigits(num float64, decimal int) float64 {
	if decimal == 0 {
		return num
	}

	shift := math.Pow10(decimal)
	return math.Trunc(num*shift) / shift
}

func DecimalPlaces(num float64, decimal int) float64 {
	shift := math.Pow10(decimal)
	return num / shift
}

func ComputeRatio(n1, n2 float64, decimal int) float64 {
	if n2 == 0 {
		return 0
	}

	return TakeDigits(n1/n2, decimal)
}

func PadInt64(num int64, n int) string {
	return fmt.Sprintf("%0*d", n, num)
}

func PadNumStr(raw string, n int) string {
	num, err := ParseInt64FromStr(raw)
	if err != nil {
		return raw
	}

	return PadInt64(num, n)
}
