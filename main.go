package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

/*
	Several things are required to fix the issue:
	1. How to read data from CSV file.
	2. How to filter datetime RFC3339 based on the time range.
	3. How to save data to CSV file.
*/

func readReports(reportDir string) ([][]string, error) {
	file, err := os.Open(reportDir)

	if err != nil {
		return nil, fmt.Errorf("error while reading the file: %v", err)
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	if err != nil {
		errMsg := fmt.Errorf("error reading records: %v", err)
		return nil, errMsg
	}

	defer func(file *os.File) {
		err = file.Close()
	}(file)

	return records, err
}

func writeReports(targetDir string, records [][]string) error {
	file, err := os.Create(fmt.Sprintf("%s.csv", targetDir))
	if err != nil {
		errMsg := fmt.Errorf("failure to write file: %v", err)
		return errMsg
	}

	defer func(file *os.File) {
		err = file.Close()
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.WriteAll(records)
	if err != nil {
		return err
	}
	return nil
}

// checking time range whether it's between the time range or not.
// left, equal, right is the return value. 'left' mean that the input time range is less than from the lowest value of the file.
// 'right' mean that the input time range is more than from the highest value of the file.
// 'equal' mean that the input time range is between the lowest and highest value of the file.
func checkTimeRange(start, end, check time.Time) string {
	checkerResult := "equal"

	if (check.Equal(start) || check.After(start)) && check.Before(end) {
		return checkerResult
	} else if check.Before(start) {
		checkerResult = "left"
	} else if check.After(end) {
		checkerResult = "right"
	}

	return checkerResult
}

// search which file that contains value from the searched date. the search algorithm using binary search.
// the output are int and string value. int represent the file identifier where the search process step.
// string represent the end result, whether the searched date is in the file ('equal') or in another file ('left' or 'right')
func searchFile(mainDir string, totalFiles int, searchedDate time.Time) (int, string, error) {
	minRange := 1
	maxRange := totalFiles
	midIndex := minRange + (maxRange-minRange)/2

	var position string

	for minRange <= maxRange {
		midIndex = minRange + (maxRange-minRange)/2

		records, err := readReports(fmt.Sprintf("%s/%d_report.csv", mainDir, midIndex))
		if err != nil {
			return 0, "", err
		}

		firstDate, err := time.Parse(time.RFC3339, records[0][1])
		if err != nil {
			errMsg := fmt.Errorf("invalid date time: %v", err)
			return 0, "", errMsg
		}

		recordLen := len(records)
		lastDate, err := time.Parse(time.RFC3339, records[recordLen-1][1])
		if err != nil {
			errMsg := fmt.Errorf("invalid date time: %v", err)
			return 0, "", errMsg
		}

		position = checkTimeRange(firstDate, lastDate, searchedDate)

		if position == "equal" {
			return midIndex, position, nil
		} else if position == "left" {
			maxRange = midIndex - 1
		} else {
			minRange = midIndex + 1
		}

	}
	return midIndex, position, nil
}

// search the index of the searched date within the file. a file contain 20 lines of data.
// the function will help to search the index/line number where the searched date matched.
func searchDate(records [][]string, searchedDate time.Time) (int, error) {
	minRange, maxRange := 0, len(records)
	midIndex := minRange + (maxRange-minRange)/2

	for minRange <= maxRange {
		midIndex = minRange + (maxRange-minRange)/2

		currentDate, err := time.Parse(time.RFC3339, records[midIndex][1])
		if err != nil {
			errMsg := fmt.Errorf("invalid date time: %v", err)
			return -1, errMsg
		}

		if searchedDate.Equal(currentDate) {
			return midIndex, nil
		} else if searchedDate.After(currentDate) {
			minRange = midIndex + 1
		} else {
			maxRange = midIndex - 1
		}
	}
	return midIndex, nil
}

func main() {
	var (
		directory    string
		startTimeStr string
		endTimeStr   string
	)

	flag.StringVar(&directory, "d", "files_report", "Define the directory")
	flag.StringVar(&startTimeStr, "s", "2023-12-10T16:39:13+07:00", "Start time: RFC3339")
	flag.StringVar(&endTimeStr, "e", "2024-02-20T19:04:18+07:00", "End time: RFC3339")

	flag.Parse()

	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	totalFiles := len(files)
	startTime, err := time.Parse(time.RFC3339, startTimeStr)

	if err != nil {
		fmt.Printf("invalid date time: %v\n", err)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)

	if err != nil {
		fmt.Printf("invalid date time: %v\n", err)
		return
	}

	if startTime.After(endTime) {
		fmt.Println("incorrect start time: start time more than end time")
		return
	} else if endTime.Before(startTime) {
		fmt.Println("incorrect end time: end time less than start time")
		return
	} else if startTime.Equal(endTime) || endTime.Equal(startTime) {
		fmt.Println("incorrect start/end time: start and end time is equal")
		return
	}

	fileIndexStart, searchDescStart, err := searchFile(directory, totalFiles, startTime)

	if err != nil {
		fmt.Println(err)
		return
	}

	fileIndexEnd, searchDescEnd, err := searchFile(directory, totalFiles, endTime)

	if err != nil {
		fmt.Println(err)
		return
	}

	if fileIndexStart == 0 || fileIndexEnd == 0 {
		fmt.Println("error during search file")
		return
	}

	if (searchDescStart == "left" && searchDescEnd == "left") || (searchDescStart == "right" && searchDescEnd == "right") {
		fmt.Println("data not found!")
		return
	}

	finalReports := make([][]string, 0)

	recordsBegin, err := readReports(fmt.Sprintf("%s/%d_report.csv", directory, fileIndexStart))
	if err != nil {
		fmt.Println(err)
		return
	}

	startIndex := 0
	if searchDescStart == "equal" {
		startIndex, err = searchDate(recordsBegin, startTime)

		if err != nil {
			fmt.Println(err)
			return
		}

	}

	recordsEnd, err := readReports(fmt.Sprintf("%s/%d_report.csv", directory, fileIndexEnd))

	if err != nil {
		fmt.Println(err)
		return
	}

	endIndex := len(recordsEnd) + 1
	if searchDescEnd == "equal" {
		endIndex, err = searchDate(recordsEnd, endTime)

		if err != nil {
			fmt.Println(err)
			return
		}
	}

	isSameFile := fileIndexEnd - fileIndexStart

	if isSameFile == 0 {
		finalReports = append(finalReports, recordsBegin[startIndex:endIndex]...)
	} else {
		finalReports = append(finalReports, recordsBegin[startIndex:]...)

		for i := fileIndexStart + 1; i < fileIndexEnd; i++ {
			records, err := readReports(fmt.Sprintf("%s/%d_report.csv", directory, i))
			if err != nil {
				fmt.Println(err)
				return
			}
			finalReports = append(finalReports, records...)
		}

		finalReports = append(finalReports, recordsEnd[:endIndex]...)
	}

	err = writeReports("final_result", finalReports)

	if err != nil {
		fmt.Println(err)
		return
	}
}
