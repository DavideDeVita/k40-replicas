package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

var _FOLDER string

func rand_01() float32 {
	// Generate a random float32 in the range [0, 1)
	return rand.Float32()
}

func rand_ab_int(a int, b int) int {
	// Generate a random float32 in the range [0, 1)
	return a + rand.Intn(1+b-a)
}

func rand_ab_float(a float32, b float32) float32 {
	// Generate a random float32 in the range [0, 1)
	r := rand_01()
	r *= b - a
	return r + a
}

func rand_10pow(a int, b int) float32 {
	// Generate a random float32 in the range [0, 1)
	r := rand.Float32()
	a10 := float32(math.Pow10(a))
	b10 := float32(math.Pow10(b))
	r *= b10 - a10
	return r + a10
}

func keepSign_centiSqr(base int) float32 {
	/* /100 to keep it more readable */
	var sign float32 = 1.
	fbase := float32(base) / 100.
	if base < 0. {
		sign = -1.
	}
	return (fbase * fbase) * sign
}

func abs(val float32) float32 {
	if val < 0 {
		return -val
	} else {
		return val
	}
}

func log_f32(val float32, base float32) float32 {
	return log10_f32(val) / log10_f32(base)
}

func log10_f32(val float32) float32 {
	return float32(math.Log10(float64(val)))
}

func log10_int(val int) float32 {
	return float32(math.Log10(float64(val)))
}

func float_to_str(num float32, dig int) string {
	// Format the float with 2 decimal places
	var formatted string
	if float32(int(num)) == num {
		formatted = strconv.FormatFloat(float64(num), 'f', 1, 64)
	} else {
		formatted = strconv.FormatFloat(float64(num), 'f', dig, 64)
	}
	// Replace the dot with a comma
	return strings.Replace(formatted, ".", ",", 1)
}

func power_floor(base float32, exp int) int {
	return int(power_f32(base, float32(exp)))
}

func power_f32(base float32, exp float32) float32 {
	return float32(math.Pow(float64(base), float64(exp)))
}

func exppower_f32(exp float32) float32 {
	return float32(math.Pow(math.E, float64(exp)))
}

func matrixToCsv(filename string, matrix [][]float32, header []string, digits int) {
	// Create a new CSV file
	if !strings.HasSuffix(filename, ".csv") {
		log.Println("filename (", filename, ") should be csv..")
	}

	// Removing evt file
	err := os.Remove(filename)
	if err != nil {
		log.Println("Error deleting file:", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Println("Error creating file:", err)
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if header != nil && len(header) > 0 {
		err := writer.Write(header)
		if err != nil {
			log.Println("Error writing header in CSV:", err)
		}
	}

	// Write the matrix to the CSV file
	for _, row := range matrix {
		row_str := make([]string, len(row))
		for i := range row {
			row_str[i] = float_to_str(row[i], digits)
		}

		err := writer.Write(row_str)
		if err != nil {
			log.Println("Error writing row to CSV:", err)
		}
	}

	log.Println("CSV file created successfully")
}

func matrixToCsv_s(filename string, matrix [][]string, header []string, digits int) {
	// Create a new CSV file
	if !strings.HasSuffix(filename, ".csv") {
		log.Println("filename (", filename, ") should be csv..")
	}

	// Removing evt file
	err := os.Remove(filename)
	if err != nil {
		log.Println("Error deleting file:", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Println("Error creating file:", err)
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if header != nil && len(header) > 0 {
		err := writer.Write(header)
		if err != nil {
			log.Println("Error writing header in CSV:", err)
		}
	}

	// Write the matrix to the CSV file
	for _, row := range matrix {
		row_str := make([]string, len(row))
		for i := range row {
			row_str[i] = row[i]
		}

		err := writer.Write(row_str)
		if err != nil {
			log.Println("Error writing row to CSV:", err)
		}
	}

	log.Println("CSV file created successfully")
}

func matrixToCsv_i(filename string, matrix [][]int, header []string) {
	// Create a new CSV file
	if !strings.HasSuffix(filename, ".csv") {
		log.Println("filename (", filename, ") should be csv..")
	}

	// Removing evt file
	err := os.Remove(filename)
	if err != nil {
		log.Println("Error deleting file:", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Println("Error creating file:", err)
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if header != nil && len(header) > 0 {
		err := writer.Write(header)
		if err != nil {
			log.Println("Error writing header in CSV:", err)
		}
	}

	// Write the matrix to the CSV file
	for _, row := range matrix {
		row_str := make([]string, len(row))
		for i := range row {
			row_str[i] = fmt.Sprintf("%d", row[i])
		}

		err := writer.Write(row_str)
		if err != nil {
			log.Println("Error writing row to CSV:", err)
		}
	}

	log.Println("CSV file created successfully")
}

func sortByPrimary_Assurance(primary_ass []float32, secondary_scores []float32, secondary_wn []*WorkerNode, secondary_cns []ClusterNodeState, condition func(a, b float32) bool, reverse bool) {
	n := len(primary_ass)

	// Create an index slice
	indices := make([]int, n)
	for i := 0; i < n; i++ {
		indices[i] = i
	}

	// Sort indices based on primary array
	sort.Slice(indices, func(i, j int) bool {
		if !reverse {
			return condition(primary_ass[indices[i]], primary_ass[indices[j]])
		} else {
			return !condition(primary_ass[indices[i]], primary_ass[indices[j]])
		}
	})

	// Reorder primary array
	tempPrimary := make([]float32, n)
	tempSecondary := make([]float32, n)
	tempNodes := make([]*WorkerNode, n)
	tempStates := make([]ClusterNodeState, n)

	for i, idx := range indices {
		tempPrimary[i] = primary_ass[idx]
		tempSecondary[i] = secondary_scores[idx]
		tempNodes[i] = secondary_wn[idx]
		tempStates[i] = secondary_cns[idx]
	}

	copy(primary_ass, tempPrimary)
	copy(secondary_scores, tempSecondary)
	copy(secondary_wn, tempNodes)
	copy(secondary_cns, tempStates)
}

func const_array(arr []int) bool {
	if len(arr) > 0 {
		st := arr[0]
		for _, x := range arr[1:] {
			if x != st {
				return false
			}
		}
	}
	return true
}

func readableNanoseconds(ns int64) string {
	// Define time units in nanoseconds
	const (
		nanosecond  = 1
		microsecond = 1000 * nanosecond
		millisecond = 1000 * microsecond
		second      = 1000 * millisecond
		minute      = 60 * second
	)

	// Break down the input duration into time components
	minutes := ns / minute
	ns %= minute
	seconds := ns / second
	ns %= second
	milliseconds := ns / millisecond
	ns %= millisecond
	microseconds := ns / microsecond
	nanoseconds := ns % microsecond

	// Build the human-readable string
	result := ""
	if minutes > 0 {
		result += fmt.Sprintf("%d min ", minutes)
	}
	if seconds > 0 || minutes > 0 { // Include seconds if minutes are present
		result += fmt.Sprintf("%d s ", seconds)
	}
	if milliseconds > 0 || seconds > 0 || minutes > 0 {
		result += fmt.Sprintf("%d ms ", milliseconds)
	}
	if microseconds > 0 || milliseconds > 0 || seconds > 0 || minutes > 0 {
		result += fmt.Sprintf("%d μs ", microseconds)
	}
	result += fmt.Sprintf("%d ns", nanoseconds)

	return result
}
