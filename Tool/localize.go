package main

import (
	"bytes"
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
	forceRescanFlag          bool
	extensionsFlag           string
	localizationFileNameFlag string
	convertStringsModeFlag   bool
	cleanFlag                bool
	verboseFlag              bool
)

var (
	extensions                   []string
	languages                    []string
	languagesCount               int
	localizationModificationTime time.Time
	localizeRegexp               *regexp.Regexp
	stripCommentsRegexp          *regexp.Regexp
	stringsRegexp                *regexp.Regexp
	localization                 map[string][]string
)

const (
	languageKey     = "language"
	defaultLanguage = "en"
)

func init() {
	const (
		localizeFunctionDefault     = "Localize"
		localizeFunctionUsage       = "Name of localization routine"
		outputDirectoryDefault      = "."
		outputDirectoryUsage        = "Output directory for localization file"
		localizationFileNameDefault = "Localization.csv"
		localizationFileNameUsage   = "Name of localization file"
		forceRescanDefault          = false
		forceRescanUsage            = "Force rescan of all files (modification time will be ignored)"
		extensionsFlagDefault       = "m,mm"
		extensionsFlagUsage         = "Comma-separated list of extensions of files which should be scanned"
		convertStringsModeDefault   = false
		convertStringsModeUsage     = "Enables 'conversion' mode. Recursively converts .strings files found in <path> to single .csv file"
		cleanDefault                = false
		cleanUsage                  = "Clean unused localization strings. Implies -r (full rescan)"
		verboseDefault              = false
		verboseUsage                = "Use verbose output"
	)
	flag.StringVar(&localizeFunctionFlag, "s", localizeFunctionDefault, localizeFunctionUsage)
	flag.StringVar(&outputDirectoryFlag, "o", outputDirectoryDefault, outputDirectoryUsage)
	flag.BoolVar(&forceRescanFlag, "r", forceRescanDefault, forceRescanUsage)
	flag.StringVar(&extensionsFlag, "e", extensionsFlagDefault, extensionsFlagUsage)
	flag.StringVar(&localizationFileNameFlag, "n", localizationFileNameDefault, localizationFileNameUsage)
	flag.BoolVar(&convertStringsModeFlag, "k", convertStringsModeDefault, convertStringsModeUsage)
	flag.BoolVar(&cleanFlag, "c", cleanDefault, cleanUsage)
	flag.BoolVar(&verboseFlag, "v", verboseDefault, verboseUsage)
}

func usage() {
	fmt.Fprintf(os.Stderr, "%s is a tool to generate/merge CSV-based localization file for Objective-C source code.\n", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "\t%s [options] <sourcePath>\n", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "or\n")
	fmt.Fprintf(os.Stderr, "\t%s -k [options] <sourcePath>\n", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Options are:\n(text in [brackets] are default values)\n")
	fmt.Fprintf(os.Stderr, "\n")

	flag.CommandLine.VisitAll(func(flag *flag.Flag) {
		defaultValue := ""
		if len(flag.DefValue) > 0 {
			defaultValue = fmt.Sprintf(" [%v]", flag.DefValue)
		}
		fmt.Fprintf(os.Stderr, "\t-%s:\t%s%s\n", flag.Name, flag.Usage, defaultValue)
	})
}

func processSources(sourcePath string) {
	filepath.Walk(sourcePath, sourceWalkFunc)
}

func sourceWalkFunc(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}
	fileExtension := filepath.Ext(info.Name())
	if len(fileExtension) == 0 {
		return nil
	}
	fileExtension = fileExtension[1:]
	for _, extension := range extensions {
		if fileExtension == extension {
			modified := info.ModTime().After(localizationModificationTime)
			if forceRescanFlag || modified {
				processSourceFile(path)
			} else {
				if verboseFlag {
					fmt.Printf("File %v was not modified.\n", path)
				}
			}
		}
	}
	return nil
}

func processSourceFile(filePath string) {
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
	processSourceFileContents(fileContents)
}

func processSourceFileContents(fileContents string) {
	var matches [][]string = localizeRegexp.FindAllStringSubmatch(fileContents, -1)
	for _, match := range matches {
		if len(match) > 1 {
			key := match[1]
			processSourceFileKey(key)
		}
	}
	if verboseFlag {
		if len(matches) > 0 {
			fmt.Printf("Parsed %d keys.\n", len(matches))
		}
	}
}

func processSourceFileKey(key string) {
	if _, ok := localization[key]; !ok {
		localization[key] = make([]string, languagesCount)
	}
}

func stripComments(fileContents string) string {
	indicesToRemove := make([]int, 0)

	const commentMatchingGroup = 2
	const commentMatchingStartIndex = 2 * commentMatchingGroup
	const commentMatchingEndIndex = 2*commentMatchingGroup + 1

	for _, matchIndices := range stripCommentsRegexp.FindAllStringSubmatchIndex(fileContents, -1) {
		if len(matchIndices)/2 >= commentMatchingGroup+1 {
			if matchIndices[commentMatchingStartIndex] != -1 && matchIndices[commentMatchingEndIndex] != -1 {
				indicesToRemove = append(indicesToRemove, matchIndices[commentMatchingStartIndex], matchIndices[commentMatchingEndIndex])
			}
		}
	}

	var buffer bytes.Buffer
	index := 0
	for i := 0; i < len(indicesToRemove); i += 2 {
		startingIndex := indicesToRemove[i]
		endingIndex := indicesToRemove[i+1]
		buffer.WriteString(fileContents[index:startingIndex])
		index = endingIndex
	}
	buffer.WriteString(fileContents[index:len(fileContents)])
	result := buffer.String()
	return result
}

func findAllStringsLanguages(stringsFilesPath string) {
	filepath.Walk(stringsFilesPath, stringsLanguagesWalkFunc)
}

func stringsLanguagesWalkFunc(path string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		return nil
	}
	if filepath.Ext(info.Name()) != ".lproj" {
		return nil
	}
	language := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
	for _, existingLanguage := range languages {
		if existingLanguage == language {
			return nil
		}
	}
	if language == defaultLanguage {
		languages = append([]string{language}, languages...)
	} else {
		languages = append(languages, language)
	}
	return nil
}

func processStringsFiles(stringsFilesPath string) {
	filepath.Walk(stringsFilesPath, stringsWalkFunc)
}

var stringsWalkCurrentLanguageIndex int

func stringsWalkFunc(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		if filepath.Ext(info.Name()) != ".lproj" {
			return nil
		}
		stringsWalkCurrentLanguageIndex = -1
		for index, language := range languages {
			if strings.HasPrefix(info.Name(), language) {
				stringsWalkCurrentLanguageIndex = index
				return nil
			}
		}
	} else {
		if filepath.Ext(info.Name()) != ".strings" {
			return nil
		}
		if stringsWalkCurrentLanguageIndex == -1 {
			return nil
		}
		processStringFile(path)
	}
	return nil
}

func processStringFile(filePath string) {
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
	translationsCount := processStringFileContents(fileContents)

	if verboseFlag {
		fmt.Printf("Parsed %d translations from %s\n", translationsCount, filePath)
	}
}

func unescapeLocalizedString(s string) string {
	result := strings.Replace(s, "\\\"", "\"", -1)
	result = strings.Replace(result, "\\'", "'", -1)
	return result
}

func processStringFileContents(fileContents string) int {
	const expectedGroupCount = 2
	const keyGroupStartPos = 1 * 2
	const keyGroupEndPos = 1*2 + 1
	const valueGroupStartPos = 2 * 2
	const valueGroupEndPos = 2*2 + 1

	translationsCount := 0

	for _, matchIndices := range stringsRegexp.FindAllStringSubmatchIndex(fileContents, -1) {
		if len(matchIndices)/2 >= expectedGroupCount+1 {
			key := ""
			if matchIndices[keyGroupStartPos] != -1 && matchIndices[keyGroupEndPos] != -1 {
				key = fileContents[matchIndices[keyGroupStartPos]:matchIndices[keyGroupEndPos]]
			}
			value := ""
			if matchIndices[valueGroupStartPos] != -1 && matchIndices[valueGroupEndPos] != -1 {
				value = fileContents[matchIndices[valueGroupStartPos]:matchIndices[valueGroupEndPos]]
			}
			key = unescapeLocalizedString(key)
			value = unescapeLocalizedString(value)
			if len(key) > 0 && len(value) > 0 {
				processStringsFileKeyValue(key, value)
				translationsCount++
			}
		}
	}

	return translationsCount
}

func processStringsFileKeyValue(key, value string) {
	if _, ok := localization[key]; !ok {
		localization[key] = make([]string, languagesCount)
	}
	localization[key][stringsWalkCurrentLanguageIndex] = value
}

func compileLocalizeRegexp() {
	regexpString := fmt.Sprintf("(?ms)%s\\s*\\(\\s*@\"(.*?)\"\\s*(?:,\\s*[\\w]*|@\"(.*?)\"\\s*)*\\s*\\)", localizeFunctionFlag)
	localizeRegexp = regexp.MustCompile(regexpString)
}

func compileStripCommentsRegexp() {
	stripCommentsRegexp = regexp.MustCompile("(?ms)(\\\".*?\\\"|\\'.*?\\')|(/\\*.*?\\*/|//[^\\r\\n]*$)")
}

func compileStringsRegexp() {
	stringsRegexp = regexp.MustCompile("(?ms)\\s*\\\"(.*?)\\\"\\s*=\\s*\\\"(.*?)\\\"\\s*;")
}

func initLocalization() {
	localization = make(map[string][]string)
}

func loadLocalization(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
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
		keysEnding := ""
		if len(localization) > 1 {
			keysEnding = "s"
		}
		fmt.Printf("Loaded %d key%s from %s\n", len(localization), keysEnding, filePath)
	}

	languagesValues, ok := localization[languageKey]
	if !ok {
		fmt.Fprintf(os.Stderr, "%s:1: error: Missing line with 'language' key.\n", filePath)
	} else {
		languages = languagesValues[1:]
		languagesCount = len(languages)
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
	if verboseFlag {
		fmt.Printf("localization table:\n%v\n", localization)
	}
	for i, key := range keys {
		languageValues := localization[key]
		if len(languageValues) > languagesCount+1 {
			fmt.Fprintf(os.Stderr, "%s:%d: warning: Key '%s' has more translations (%d) than languages specified (%d)\n",
				outputFilePath, i+1, key, len(languageValues), languagesCount)
		} else if len(languageValues) < languagesCount+1 {
			fmt.Fprintf(os.Stderr, "%s:%d: warning: Key '%s' has less translations (%d) than languages specified (%d)\n",
				outputFilePath, i+1, key, len(languageValues), languagesCount)
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
			fmt.Fprintf(os.Stderr, "%s:%d: warning: Missing translation%s for key '%s' for language%s %s\n",
				outputFilePath, i+1, translationsEnding, key, languagesEnding, missingLanguages)
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
		values := append([]string{key}, localization[key]...)
		csvWriter.Write(values)
	}

	csvWriter.Flush()

	if verboseFlag {
		keysEnding := ""
		if len(keys) > 1 {
			keysEnding = "s"
		}
		fmt.Printf("Saved file with %d key%s.\n", len(keys), keysEnding)
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

	extensions = strings.Split(extensionsFlag, ",")

	return
}

func main() {
	sourcePath, outputFilePath := parseArguments()
	initLocalization()
	compileStripCommentsRegexp()
	if !convertStringsModeFlag {
		compileLocalizeRegexp()
		loadLocalization(outputFilePath)
		processSources(sourcePath)
	} else {
		compileStringsRegexp()
		processStringsFiles(sourcePath)
	}
	keys := sortedKeys()
	checkLocalization(keys, outputFilePath)
	saveLocalization(keys, outputFilePath)
}
