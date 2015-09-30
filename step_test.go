package main

import (
	"io/ioutil"
	"runtime/debug"
	"testing"
)

//
// --- TESTS

func TestFindMajorVersion(t *testing.T) {
	// Should find
	versionStr := `
Xcode 7.0
Build version 7A220`

	version, err := findMajorVersion(versionStr)
	if err != nil {
		t.Fatalf("Failed to findMajorVersion: %s", err)
	}
	if version != 7 {
		t.Fatalf("Expected version (7), actual (%d)", version)
	}

	// Should not find
	versionStr = `
	Build version 7A220`

	version, err = findMajorVersion(versionStr)
	if err == nil {
		t.Fatalf("Should failed to findMajorVersion: %s", err)
	}
	if version != -1 {
		t.Fatalf("Expected version (-1), actual (%d)", version)
	}
}

func Test_isStringFoundInOutput(t *testing.T) {
	// Should NOT find
	searchPattern := "something"
	isShouldFind := false
	for _, anOutStr := range []string{
		"",
		"a",
		"1",
		"somethin",
		"somethinx",
		"TEST FAILED",
	} {
		if isFound := isStringFoundInOutput(searchPattern, anOutStr); isFound != isShouldFind {
			t.Logf("Search pattern was: %s", searchPattern)
			t.Logf("Input string to search in was: %s", anOutStr)
			t.Fatalf("Expected (%v), actual (%v)", isShouldFind, isFound)
		}
	}

	// Should find
	searchPattern = "search for this"
	isShouldFind = true
	for _, anOutStr := range []string{
		"search for this",
		"-search for this",
		"search for this-",
		"-search for this-",
	} {
		if isFound := isStringFoundInOutput(searchPattern, anOutStr); isFound != isShouldFind {
			t.Logf("Search pattern was: %s", searchPattern)
			t.Logf("Input string to search in was: %s", anOutStr)
			t.Fatalf("Expected (%v), actual (%v)", isShouldFind, isFound)
		}
	}

	// Should find - empty pattern - always "yes"
	searchPattern = ""
	isShouldFind = true
	for _, anOutStr := range []string{
		"",
		"a",
		"1",
		"TEST FAILED",
	} {
		if isFound := isStringFoundInOutput(searchPattern, anOutStr); isFound != isShouldFind {
			t.Logf("Search pattern was: %s", searchPattern)
			t.Logf("Input string to search in was: %s", anOutStr)
			t.Fatalf("Expected (%v), actual (%v)", isShouldFind, isFound)
		}
	}
}

func Test_findTestSummaryInOutput(t *testing.T) {
	// load sample logs
	sampleIPhoneSimulatorLog, err := loadFileContent("./_samples/xcodebuild-iPhoneSimulator-timeout.txt")
	if err != nil {
		t.Fatalf("Failed to load error sample log : %s", err)
	}
	sampleUITestTimeoutLog, err := loadFileContent("./_samples/xcodebuild-UITest-timeout.txt")
	if err != nil {
		t.Fatalf("Failed to load (UITest timeout) error sample log : %s", err)
	}
	sampleOKBuildLog, err := loadFileContent("./_samples/xcodebuild-ok.txt")
	if err != nil {
		t.Fatalf("Failed to load xcodebuild-ok.txt : %s", err)
	}

	// Should NOT find
	for _, anOutStr := range []string{
		"",
		"xyz",
		"TEST SUCCEEDED",
		"TEST FAILED",
		sampleIPhoneSimulatorLog,
	} {
		testFindTestSummaryInOutput(t, false, anOutStr, true)
		testFindTestSummaryInOutput(t, false, anOutStr, false)
	}

	// should find
	testFindTestSummaryInOutput(t, true, "** TEST SUCCEEDED **", true)
	testFindTestSummaryInOutput(t, true, "** TEST FAILED **", false)
	testFindTestSummaryInOutput(t, true, "Failing tests:", false)
	testFindTestSummaryInOutput(t, true, "Testing failed:", false)
	testFindTestSummaryInOutput(t, true, sampleOKBuildLog, true)
	testFindTestSummaryInOutput(t, true, sampleUITestTimeoutLog, false)
}

func TestIsStringFoundInOutput_timeOutMessageIPhoneSimulator(t *testing.T) {
	// load sample logs
	sampleIPhoneSimulatorLog, err := loadFileContent("./_samples/xcodebuild-iPhoneSimulator-timeout.txt")
	if err != nil {
		t.Fatalf("Failed to load error sample log : %s", err)
	}
	sampleOKBuildLog, err := loadFileContent("./_samples/xcodebuild-ok.txt")
	if err != nil {
		t.Fatalf("Failed to load xcodebuild-ok.txt : %s", err)
	}

	// Should find
	for _, anOutStr := range []string{
		"iPhoneSimulator: Timed out waiting",
		"iphoneSimulator: timed out waiting",
		"iphoneSimulator: timed out waiting, test test test",
		"aaaiphoneSimulator: timed out waiting, test test test",
		"aaa iphoneSimulator: timed out waiting, test test test",
		sampleIPhoneSimulatorLog,
	} {
		testIPhoneSimulatorTimeoutWith(t, anOutStr, true)
	}

	// Should not
	for _, anOutStr := range []string{
		"",
		"iphoneSimulator:",
		sampleOKBuildLog,
	} {
		testIPhoneSimulatorTimeoutWith(t, anOutStr, false)
	}
}

func TestIsStringFoundInOutput_timeOutMessageUITest(t *testing.T) {
	// load sample logs
	sampleUITestTimeoutLog, err := loadFileContent("./_samples/xcodebuild-UITest-timeout.txt")
	if err != nil {
		t.Fatalf("Failed to load error sample log : %s", err)
	}
	sampleOKBuildLog, err := loadFileContent("./_samples/xcodebuild-ok.txt")
	if err != nil {
		t.Fatalf("Failed to load xcodebuild-ok.txt : %s", err)
	}

	// Should find
	for _, anOutStr := range []string{
		"Terminating app due to uncaught exception '_XCTestCaseInterruptionException'",
		"terminating app due to uncaught exception '_XCTestCaseInterruptionException'",
		"aaTerminating app due to uncaught exception '_XCTestCaseInterruptionException'aa",
		"aa Terminating app due to uncaught exception '_XCTestCaseInterruptionException' aa",
		sampleUITestTimeoutLog,
	} {
		testTimeOutMessageUITestWith(t, anOutStr, true)
	}

	// Should not
	for _, anOutStr := range []string{
		"",
		"Terminating app due to uncaught exception",
		"_XCTestCaseInterruptionException",
		sampleOKBuildLog,
	} {
		testTimeOutMessageUITestWith(t, anOutStr, false)
	}
}

func Test_printableCommandArgs(t *testing.T) {
	orgCmdArgs := []string{
		"xcodebuild", "-project", "MyProj.xcodeproj", "-scheme", "MyScheme",
		"build", "test",
		"-destination", "platform=iOS Simulator,name=iPhone 6,OS=latest",
		"-sdk", "iphonesimulator",
	}
	resStr := printableCommandArgs(orgCmdArgs)
	expectedStr := `xcodebuild "-project" "MyProj.xcodeproj" "-scheme" "MyScheme" "build" "test" "-destination" "platform=iOS Simulator,name=iPhone 6,OS=latest" "-sdk" "iphonesimulator"`

	if resStr != expectedStr {
		t.Log("printableCommandArgs failed to generate the expected string!")
		t.Logf(" -> expectedStr: %s", expectedStr)
		t.Logf(" -> resStr: %s", resStr)
		t.Fatalf("Expected string does not match the generated string. Original args: (%#v)", orgCmdArgs)
	}
}

func Test_findFirstDelimiter(t *testing.T) {
	test := func(strToSearchIn string, searchDelims []string, expectedDelim string, expectedIdx int) {
		foundIdx, foundDelim := findFirstDelimiter(strToSearchIn, searchDelims)

		if foundDelim != expectedDelim {
			t.Log("foundDelim != expectedDelim")
			t.Logf(" -> foundDelim: %s", foundDelim)
			t.Logf(" -> expectedDelim: %s", expectedDelim)
			t.Fatalf("searchDelims: (%#v) | strToSearchIn: %s", searchDelims, strToSearchIn)
		}
		if foundIdx != expectedIdx {
			t.Log("foundIdx != expectedIdx")
			t.Logf(" -> foundIdx: %v", foundIdx)
			t.Logf(" -> expectedIdx: %v", expectedIdx)
			t.Fatalf("searchDelims: (%#v) | strToSearchIn: %s", searchDelims, strToSearchIn)
		}
	}

	strToSearchIn := `Touch /Users/bitrise/Library/Developer/Xcode/DerivedData/BitriseSampleWithYML-duyungvabqzmagefbqehoqyodaqz/Build/Products/Debug-iphonesimulator/BitriseSampleWithYMLTests.xctest
    cd /Users/bitrise/develop/sample-apps-ios-with-bitrise-yml
    export PATH="/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneSimulator.platform/Developer/usr/bin:/Applications/Xcode.app/Contents/Developer/usr/bin:/usr/local/heroku/bin:/usr/local/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Users/bitrise/develop/go/bin:/usr/local/opt/go/libexec/bin:/Users/bitrise/bin:/Users/bitrise/develop/go/bin:/usr/local/opt/go/libexec/bin:/Users/bitrise/bin"
    /usr/bin/touch -c /Users/bitrise/Library/Developer/Xcode/DerivedData/BitriseSampleWithYML-duyungvabqzmagefbqehoqyodaqz/Build/Products/Debug-iphonesimulator/BitriseSampleWithYMLTests.xctest

** TEST SUCCEEDED **

Test Suite 'All tests' started at 2015-09-20 16:01:13.995
Test Suite 'BitriseSampleWithYMLTests.xctest' started at 2015-09-20 16:01:13.997`

	// found
	// note: order of delimiters doesn't matter, first occurance will be used
	test(strToSearchIn, []string{"TEST SUCCEEDED", "Test Suite"}, "TEST SUCCEEDED", 842)
	test(strToSearchIn, []string{"Test Suite", "TEST SUCCEEDED"}, "TEST SUCCEEDED", 842)
	test("aaaa Test Suite aaa", []string{"Test Suite", "TEST SUCCEEDED"}, "Test Suite", 5)
	test("aaaa Test Suite aaa", []string{"TEST SUCCEEDED", "Test Suite"}, "Test Suite", 5)

	// not found
	test("aaaa", []string{"TEST SUCCEEDED", "Test Suite"}, "", -1)
}

//
// --- TESTING UTILITY FUNCS

func testIsFoundWith(t *testing.T, searchPattern, outputToSearchIn string, isShouldFind bool) {
	if isFound := isStringFoundInOutput(searchPattern, outputToSearchIn); isFound != isShouldFind {
		t.Logf("Search pattern was: %s", searchPattern)
		t.Logf("Input string to search in was: %s", outputToSearchIn)
		t.Fatalf("Expected (%v), actual (%v)", isShouldFind, isFound)
	}
}
func testIPhoneSimulatorTimeoutWith(t *testing.T, outputToSearchIn string, isShouldFind bool) {
	testIsFoundWith(t, timeOutMessageIPhoneSimulator, outputToSearchIn, isShouldFind)
}

func testTimeOutMessageUITestWith(t *testing.T, outputToSearchIn string, isShouldFind bool) {
	testIsFoundWith(t, timeOutMessageUITest, outputToSearchIn, isShouldFind)
}

func testFindTestSummaryInOutput(t *testing.T, isExpectToFind bool, fullOutput string, isRunSucess bool) {
	resStr := findTestSummaryInOutput(fullOutput, isRunSucess)
	if isExpectToFind && resStr == "" {
		t.Logf("Expected to find Test Summary in provided log.")
		debug.PrintStack()
		t.Fatalf("Provided output was: %s", fullOutput)
	}
	if !isExpectToFind && resStr != "" {
		t.Logf("Expected to NOT find Test Summary in provided log.")
		debug.PrintStack()
		t.Fatalf("Provided output was: %s", fullOutput)
	}
}

func loadFileContent(filePth string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filePth)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}
