package point

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

func primeFactorization(n int) (factors map[int]int) {
	factors = make(map[int]int, 0)
	if n == 1 {
		factors[n] = 1
		return
	}

	i := 2
	for n != 1 {

		if n%i == 0 {
			n = n / i
			factors[i]++
		} else {
			i++
		}

	}

	return
}

func tagCardinalityPartition(numTags int, factors map[int]int) []int {
	buckets := make([]int, numTags)

	for i := range buckets {
		buckets[i] = 1
	}

	orderedFactors := []int{}
	for factor := range factors {
		orderedFactors = append(orderedFactors, factor)
	}
	sort.Ints(orderedFactors)

	i := 0
	for _, factor := range orderedFactors {
		power := factors[factor]
		buckets[i%len(buckets)] *= int(math.Pow(float64(factor), float64(power)))
		i++
	}

	return buckets
}

func generateSeriesKeys(measurement, tmplt string, card int) [][]byte {
	fmtTmplt, numTags := formatTemplate(measurement, tmplt)
	tagCardinalities := tagCardinalityPartition(numTags, primeFactorization(card))

	series := []string{}
	seriesAsBytes := [][]byte{}

	for i := 0; i < card; i++ {
		mods := sliceMod(i, tagCardinalities)
		serie := fmt.Sprintf(fmtTmplt, mods...)
		series = append(series, serie)
		seriesAsBytes = append(seriesAsBytes, []byte(serie))
	}

	return seriesAsBytes
}

func formatTemplate(m, s string) (string, int) {
	parts := strings.Split(s, ",")

	for i, part := range parts {
		parts[i] = part + "-%v"
	}

	return m + "," + strings.Join(parts, ","), len(parts)
}

func sliceMod(m int, mods []int) []interface{} {
	ms := []interface{}{}
	for _, mod := range mods {
		ms = append(ms, m%mod)
	}

	return ms
}
