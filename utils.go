package main

import (
	"encoding/csv"
	"log"
	"math/rand"
	"os"
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
	r := rand_01()
	r *= float32(b - a)
	return int(r) + a
}

func rand_ab_float(a float32, b float32) float32 {
	// Generate a random float32 in the range [0, 1)
	r := rand_01()
	r *= b - a
	return r + a
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
