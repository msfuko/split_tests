package main

import (
	"encoding/xml"
	"io"
	"os"
	"path"

	"github.com/bmatcuk/doublestar"
)

type junitXML struct {
	TestCases []struct {
		File string  `xml:"file,attr"`
		Time float64 `xml:"time,attr"`
	} `xml:"testcase"`
}
type Testcase struct {
	Name      string `xml:"name,attr"`
	ClassName string `xml:"classname,attr"`
	Time      string `xml:"time,attr"`
}

type Testsuite struct {
	Name      string     `xml:"name,attr"`
	Timestamp string     `xml:"timestamp,attr"`
	Hostname  string     `xml:"hostname,attr"`
	Tests     int        `xml:"tests,attr"`
	Failures  int        `xml:"failures,attr"`
	Skipped   int        `xml:"skipped,attr"`
	Time      float64    `xml:"time,attr"`
	Errors    int        `xml:"errors,attr"`
	Testcases []Testcase `xml:"testcase"`
}
type Testsuites struct {
	Id        string      `xml:"id,attr"`
	Name      string      `xml:"name,attr"`
	Tests     int         `xml:"tests,attr"`
	Failures  int         `xml:"failures,attr"`
	Skipped   int         `xml:"skipped,attr"`
	Errors    int         `xml:"errors,attr"`
	Time      float64     `xml:"time,attr"`
	Testsuite []Testsuite `xml:"testsuite"`
}

func loadJUnitXML(reader io.Reader) *junitXML {
	var junitXML junitXML

	decoder := xml.NewDecoder(reader)
	err := decoder.Decode(&junitXML)
	if err != nil {
		fatalMsg("failed to parse junit xml: %v\n", err)
	}

	return &junitXML
}

func addFileTimesFromIOReader(fileTimes map[string]float64, reader io.Reader) {
	junitXML := loadJUnitXML(reader)
	for _, testCase := range junitXML.TestCases {
		filePath := path.Clean(testCase.File)
		fileTimes[filePath] += testCase.Time
	}
}

func getFileTimesFromJUnitXML(fileTimes map[string]float64) {
	if junitXMLPath != "" {
		filenames, err := doublestar.Glob(junitXMLPath)
		if err != nil {
			fatalMsg("failed to match jUnit filename pattern: %v", err)
		}
		for _, junitFilename := range filenames {
			file, err := os.Open(junitFilename)
			if err != nil {
				fatalMsg("failed to open junit xml: %v\n", err)
			}
			defer file.Close()
			printMsg("using test times from JUnit report %s\n", junitFilename)
			xmlData, err := io.ReadAll(file)
			if err != nil {
				fatalMsg("Error reading file: %v\n", err)
				return
			}
			var testsuites Testsuites

			// Unmarshal the XML data into the testsuites variable
			err = xml.Unmarshal(xmlData, &testsuites)
			if err != nil {
				fatalMsg("Error unmarshaling XML: %v\n", err)
				return
			}

			for _, suite := range testsuites.Testsuite {
				for _, testcase := range suite.Testcases {
					fileTimes[testcase.ClassName] = testsuites.Time
				}
			}

		}
	} else {
		printMsg("using test times from JUnit report at stdin\n")
		addFileTimesFromIOReader(fileTimes, os.Stdin)
	}
}
