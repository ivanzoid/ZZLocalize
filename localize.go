package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	localizeFunctionFlag     string
	outputDirectoryFlag      string
	languagesFlag            string
	forceRescanFlag          bool
	extensionsFlag           string
	localizationFileNameFlag string
	verboseFlag              bool
)

var (
	extensions                   []string
	languages                    []string
	languagesCount               int
	localizationModificationTime time.Time
	localizeRegexp               *regexp.Regexp
	stripCommentsRegexp          *regexp.Regexp
	localization                 map[string][]string
)

const (
	languageKey = "language"
)

func init() {
	const (
		localizeFunctionDefault     = "Localize"
		localizeFunctionUsage       = "Name of localization routine."
		outputDirectoryDefault      = "."
		outputDirectoryUsage        = "Specifies what directory localization file should be created in. Default is current directory."
		languagesFlagDefault        = "en,ru"
		languagesFlagUsage          = "Comma-separated list of localization languages."
		localizationFileNameDefault = "Localization.csv"
		localizationFileNameUsage   = "Name of localization CSV file."
		forceRescanDefault          = false
		forceRescanUsage            = "By default, source file will be (re)scanned only if its modification time is newer than of " + localizationFileNameDefault + ". If this flag is present, all files are rescanned."
		extensionsFlagDefault       = "m,mm"
		extensionsFlagUsage         = "Comma-separated list of extensions of files to be scanned."
		verboseDefault              = false
		verboseUsage                = "Use verbose output."
	)
	flag.StringVar(&localizeFunctionFlag, "s", localizeFunctionDefault, localizeFunctionUsage)
	flag.StringVar(&outputDirectoryFlag, "o", outputDirectoryDefault, outputDirectoryUsage)
	flag.StringVar(&languagesFlag, "l", languagesFlagDefault, languagesFlagUsage)
	flag.BoolVar(&forceRescanFlag, "f", forceRescanDefault, forceRescanUsage)
	flag.StringVar(&extensionsFlag, "e", extensionsFlagDefault, extensionsFlagUsage)
	flag.StringVar(&localizationFileNameFlag, "n", localizationFileNameDefault, localizationFileNameUsage)
	flag.BoolVar(&verboseFlag, "v", verboseDefault, verboseUsage)
}

func usage() {
	fmt.Fprintf(os.Stderr, "%s is a tool to generate/merge csv-based localization file for Objective-C source code.\n", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "\t%s [options] path\n", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Options are:\n")
	fmt.Fprintf(os.Stderr, "\n")
	flag.PrintDefaults()
}

func walkFunc(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}
	for _, extension := range extensions {
		if filepath.Ext(info.Name()) == extension {
			if forceRescanFlag || info.ModTime().After(localizationModificationTime) {
				processFile(path)
			}
		}
	}
	return nil
}

func processFile(filePath string) {
	if verboseFlag {
		fmt.Println("Processing " + filePath)
	}
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	fileContents := string(bytes)
	fileContents = stripComments(fileContents)
	processFileContents(fileContents)
}

func processFileContents(fileContents string) {
	var matches [][]string = localizeRegexp.FindAllStringSubmatch(fileContents, -1)
	for _, match := range matches {
		if len(match) > 1 {
			key := match[1]
			processKey(key)
		}
	}
	if verboseFlag {
		if len(matches) > 0 {
			fmt.Printf("Parsed %d keys.", len(matches))
		}
	}
}

func processKey(key string) {
	if _, ok := localization[key]; !ok {
		localization[key] = make([]string, languagesCount)
	}
}

func stripComments(fileContents string) string {
	return stripCommentsRegexp.ReplaceAllString(fileContents, "")
}

func compileLocalizeRegexp() {
	regexpString := fmt.Sprintf("%s\\s*\\(\\s*@\"(.*)\"\\s*\\)", localizeFunctionFlag)
	localizeRegexp = regexp.MustCompile(regexpString)
}

func compileStripCommentsRegexp() {
	stripCommentsRegexp = regexp.MustCompile("(?ms)//.*?$|/\\*.*?\\*/|'(?:\\.|[^\\'])*'|\"(?:\\.|[^\\\"])*\"")
}

func loadLocalization(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer file.Close()

	info, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Warning: ", err)
	}
	localizationModificationTime = info.ModTime()

	reader := csv.NewReader(file)
	for {
		values, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		if len(values) >= 1 {
			localization[values[0]] = values[1:]
		}
	}

	if verboseFlag {
		fmt.Printf("Loaded %d keys from %s", filePath)
	}
}

func localizationKeys() []string {
	keys := make([]string, len(localization))
	i := 0
	for key := range localization {
		keys[i] = key
		i++
	}
	return keys
}

func checkLocalization(keys []string, outputFilePath string) {
	for i, key := range keys {
		languageValues := localization[key]
		if len(languageValues) > languagesCount {
			fmt.Fprintf(os.Stderr, "%s:%d: warn : Key '%s' has more translations (%d) than languages specified (%d)\n", outputFilePath, i, len(languageValues), languagesCount)
		} else if len(languageValues) < languagesCount {
			fmt.Fprintf(os.Stderr, "%s:%d: warn : Key '%s' has less translations (%d) than languages specified (%d)\n", outputFilePath, i, len(languageValues), languagesCount)
		}
		missingTranslations := make([]string, 0)
		for index, language := range languages {
			if index < len(languageValues) && languageValues[index] == "" {
				missingTranslations = append(missingTranslations, language)
			}
		}
		if len(missingTranslations) > 0 {
			missingLanguages := strings.Join(missingTranslations, ", ")
			translationsEnding := ""
			languagesEnding := ""
			if len(missingTranslations) > 1 {
				translationsEnding = "s"
				languagesEnding = "s:"
			}
			fmt.Fprintf(os.Stderr, "%s:%d: warn : Missing translation%s for key '%s' for language%s %s", outputFilePath, i, translationsEnding, key, languagesEnding, missingLanguages)
		}
	}
}

func sortedKeys() []string {
	languageValues := localization[languageKey]
	delete(localization, languageKey)

	keys := localizationKeys()
	sort.Strings(keys)

	if languageValues == nil {
		languageValues = languages
	}

	localization[languageKey] = languageValues
	keys = append([]string{languageKey}, keys...)
	return keys
}

func saveLocalization(keys []string, outputFilePath string) {
	file, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer file.Close()

	if verboseFlag {
		fmt.Println("Saving " + outputFilePath)
	}

	csvWriter := csv.NewWriter(file)
	for _, key := range keys {
		csvWriter.Write(localization[key])
	}

	if verboseFlag {
		fmt.Println("Done")
	}
}

func parseArguments() (sourcePath, outputFilePath string) {
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}
	sourcePath = flag.Arg(0)

	outputDir := ""
	if len(outputDirectoryFlag) > 0 {
		if filepath.IsAbs(outputDirectoryFlag) {
			outputDir = outputDirectoryFlag
		} else {
			pwd, _ := os.Getwd()
			outputDir = filepath.Join(pwd, outputDirectoryFlag)
		}
	} else {
		pwd, _ := os.Getwd()
		outputDir = pwd
	}

	outputFilePath = filepath.Join(outputDir, localizationFileNameFlag)

	languages = strings.Split(languagesFlag, ",")
	languagesCount = len(languages)

	extensions = strings.Split(extensionsFlag, ",")

	return
}

func main() {
	sourcePath, outputFilePath := parseArguments()
	compileLocalizeRegexp()
	compileStripCommentsRegexp()
	loadLocalization(outputFilePath)
	filepath.Walk(sourcePath, walkFunc)
	keys := sortedKeys()
	checkLocalization(keys, outputFilePath)
	saveLocalization(keys, outputFilePath)
}
