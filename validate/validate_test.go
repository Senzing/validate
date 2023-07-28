package validate

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------------------
// test Read method
func TestRead(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
}

func TestRead_bad(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 10; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
}

func TestRead_bad_loglevel(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
		LogLevel: "BAD",
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 10; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Unable to set log level to BAD"
	assert.Contains(t, got, msg)
	assert.Equal(t, true, result)
}

func TestRead_bad_url(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputUrl: "BAD",
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 1; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error, Check the input-url parameter: BAD"
	assert.Contains(t, got, msg)
	assert.Equal(t, false, result)
}

func TestRead_bad_url_parse(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputUrl: "http://bad:bad{BAD=bad@example.com",
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 2; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error parsing input-url"
	assert.Contains(t, got, msg)
	assert.Equal(t, false, result)
}

func TestRead_bad_drop_through(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputUrl: "BAD,Really bad",
	}
	result := validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 2; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error unable to handle"
	assert.Contains(t, got, msg)
	assert.Equal(t, false, result)
}

func TestRead_bad_file_doesnt_exist(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{
		InputUrl: "file:///badfile.jsonl",
	}
	validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Fatal error opening input file: /badfile.jsonl"
	assert.Contains(t, got, msg)
}

func TestRead_bad_file_type(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "txt")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}

	validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 2; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "If this is a valid JSONL file, please rename with the .jsonl extension or use the file type override (--file-type)"
	assert.Contains(t, got, msg)
}

func TestRead_jsonOutput(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl:   fmt.Sprintf("file://%s", filename),
		JsonOutput: true,
	}
	validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 3; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
}

func TestRead_jsonOutput_bad(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl:   fmt.Sprintf("file://%s", filename),
		JsonOutput: true,
	}
	validator.Read(context.Background())

	var got string = ""
	for i := 0; i < 10; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
}

// ----------------------------------------------------------------------------
// test jsonl file read

func TestReadJsonlFile(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	validator.readJSONLFile(filename)

	scanner.Scan() // blocks until a new line is written to the pipe

	got := scanner.Text() // the last line written to the scanner
	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
}

func TestReadJsonlFile_bad(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", filename),
	}
	validator.readJSONLFile(filename)

	var got string = ""
	for i := 0; i < 8; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
}

func TestReadJsonlFile_jsonOutput(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl:   fmt.Sprintf("file://%s", filename),
		JsonOutput: true,
	}
	validator.readJSONLFile(filename)

	scanner.Scan() // blocks until a new line is written to the pipe

	got := scanner.Text() // the last line written to the scanner
	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
}

func TestReadJsonlFile_jsonOutput_bad(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
	defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl:   fmt.Sprintf("file://%s", filename),
		JsonOutput: true,
	}
	validator.readJSONLFile(filename)

	var got string = ""
	for i := 0; i < 8; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
}

// ----------------------------------------------------------------------------
// test gzip file read

func TestReadGzipFile(t *testing.T) {
	scanner, cleanUp := mockStdout(t)
	defer cleanUp()
	tmpfile, err := ioutil.TempFile("", "test.*.jsonl.gz")
	// tmpfile, err := os.OpenFile("/tmp/q.gz", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		t.Fatal(err)
	}
	gf := gzip.NewWriter(tmpfile)
	fw := bufio.NewWriter(gf)
	fw.WriteString(testGoodData)
	fw.Flush()
	gf.Close()
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	// createTempGzDataFile(t, testGoodData)
	// tmpfile, _ := createTempGzDataFile(t, testGoodData)
	// defer moreCleanUp()

	validator := &ValidateImpl{
		InputUrl: fmt.Sprintf("file://%s", tmpfile.Name()),
	}
	validator.readGZFile(tmpfile.Name())

	scanner.Scan() // blocks until a new line is written to the pipe

	got := scanner.Text() // the last line written to the scanner
	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
}

// func TestReadJsonlFile_bad(t *testing.T) {

// 	scanner, cleanUp := mockStdout(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputUrl: fmt.Sprintf("file://%s", filename),
// 	}
// 	validator.readJSONLFile(filename)

// 	var got string = ""
// 	for i := 0; i < 8; i++ {
// 		scanner.Scan()
// 		got += scanner.Text()
// 		got += "\n"
// 	}

// 	msg := "Validated 16 lines, 4 were bad"
// 	assert.Contains(t, got, msg)
// }

// func TestReadJsonlFile_jsonOutput(t *testing.T) {

// 	scanner, cleanUp := mockStderr(t)
// 	defer cleanUp()

// 	tmpfile, moreCleanUp := createTempDataFile(t, testGoodData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputUrl:   fmt.Sprintf("file://%s", filename),
// 		JsonOutput: true,
// 	}
// 	validator.readJSONLFile(filename)

// 	scanner.Scan() // blocks until a new line is written to the pipe

// 	got := scanner.Text() // the last line written to the scanner
// 	msg := "Validated 12 lines, 0 were bad"
// 	assert.Contains(t, got, msg)
// }

// func TestReadJsonlFile_jsonOutput_bad(t *testing.T) {

// 	scanner, cleanUp := mockStderr(t)
// 	defer cleanUp()

// 	filename, moreCleanUp := createTempDataFile(t, testBadData, "jsonl")
// 	defer moreCleanUp()

// 	validator := &ValidateImpl{
// 		InputUrl:   fmt.Sprintf("file://%s", filename),
// 		JsonOutput: true,
// 	}
// 	validator.readJSONLFile(filename)

// 	var got string = ""
// 	for i := 0; i < 8; i++ {
// 		scanner.Scan()
// 		got += scanner.Text()
// 		got += "\n"
// 	}

// 	msg := "Validated 16 lines, 4 were bad"
// 	assert.Contains(t, got, msg)
// }
// ----------------------------------------------------------------------------
// test validateLines

func TestValidateLines(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{}
	validator.validateLines(strings.NewReader(testGoodData))

	scanner.Scan() // blocks until a new line is written to the pipe

	got := scanner.Text() // the last line written to the scanner

	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
}

func TestValidateLines_bad(t *testing.T) {

	scanner, cleanUp := mockStdout(t)
	defer cleanUp()

	validator := &ValidateImpl{}
	validator.validateLines(strings.NewReader(testBadData))

	var got string = ""
	for i := 0; i < 8; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
}

func TestValidateLines_jsonOutput(t *testing.T) {
	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	validator := &ValidateImpl{
		JsonOutput: true,
	}
	validator.validateLines(strings.NewReader(testGoodData))

	scanner.Scan() // blocks until a new line is written to the pipe

	got := scanner.Text() // the last line written to the scanner
	msg := "Validated 12 lines, 0 were bad"
	assert.Contains(t, got, msg)
}

func TestValidateLines_jsonOutput_bad(t *testing.T) {

	scanner, cleanUp := mockStderr(t)
	defer cleanUp()

	validator := &ValidateImpl{
		JsonOutput: true,
	}
	validator.validateLines(strings.NewReader(testBadData))

	var got string = ""
	for i := 0; i < 8; i++ {
		scanner.Scan()
		got += scanner.Text()
		got += "\n"
	}

	msg := "Validated 16 lines, 4 were bad"
	assert.Contains(t, got, msg)
}

// ----------------------------------------------------------------------------
// Helper functions
// ----------------------------------------------------------------------------

// create a tempdata baile with the given content and extension
func createTempDataFile(t *testing.T, content string, fileextension string) (filename string, cleanUp func()) {
	t.Helper()
	tmpfile, err := ioutil.TempFile(t.TempDir(), "test.*."+fileextension)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}

	filename = tmpfile.Name()

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	return filename,
		func() {
			os.Remove(filename)
		}
}

func createTempGzDataFile(t *testing.T, content string) (tmpfile *os.File, cleanUp func()) {
	t.Helper()

	// tmpfile, err := ioutil.TempFile(t.TempDir(), "test.*.jsonl.gz")
	tmpfile, err := os.OpenFile("/tmp/q.gz", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		t.Fatal(err)
	}
	gf := gzip.NewWriter(tmpfile)
	fw := bufio.NewWriter(gf)
	fw.WriteString(testGoodData)
	fw.Flush()
	gf.Close()
	tmpfile.Close()

	// gzipWriter := gzip.NewWriter(tmpfile)
	// writer := bufio.NewWriter(gzipWriter)
	// writer := bufio.NewWriter(tmpfile)
	// i, err := writer.WriteString(content)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// writer.Flush()
	// fmt.Println("bytes written:", i)
	// if err := gzipWriter.Close(); err != nil {
	// 	t.Fatal(err)
	// }
	// if err := tmpfile.Close(); err != nil {
	// 	t.Fatal(err)
	// }
	return tmpfile,
		func() {

			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}
			// os.Remove(tmpfile.Name())
		}
}

func mockStdout(t *testing.T) (buffer *bufio.Scanner, cleanUp func()) {
	t.Helper()
	origStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		assert.Fail(t, "couldn't get os Pipe: %v", err)
	}
	os.Stdout = writer

	return bufio.NewScanner(reader),
		func() {
			//clean-up
			os.Stdout = origStdout
		}
}

func mockStderr(t *testing.T) (buffer *bufio.Scanner, cleanUp func()) {
	t.Helper()
	origStderr := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		assert.Fail(t, "couldn't get os Pipe: %v", err)
	}
	os.Stderr = writer

	return bufio.NewScanner(reader),
		func() {
			//clean-up
			os.Stderr = origStderr
		}
}

var testGoodData string = `{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000001", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000002", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "SUITE E-2,UNION COURT BUILDING, P.O. BOX N-8188, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000002"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000003", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, LYFORD CAY, P.O. BOX N-7785, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000003"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000004", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "P.O. BOX N-3708 BAHAMAS FINANCIAL CENTRE, P.O. BOX N-3708 SHIRLEY & CHARLOTTE STS, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000004"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000005", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, 3RD FLOOR, LYFORD CAY, P.O. BOX N-3024, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000005"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000006", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "303 SHIRLEY STREET, P.O. BOX N-492, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000006"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000007", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "OCEAN CENTRE, MONTAGU FORESHORE, P.O. BOX SS-19084 EAST BAY STREET, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000007"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000008", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "PROVIDENCE HOUSE, EAST WING EAST HILL ST, P.O. BOX CB-12399, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000008"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000009", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "BAYSIDE EXECUTIVE PARK, WEST BAY & BLAKE, P.O. BOX N-4875, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000009"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000010", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "GROUND FLOOR, GOODMAN'S BAY CORPORATE CE, P.O. BOX N 3933, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000010"}
{"SOCIAL_HANDLE": "shuddersv", "DATE_OF_BIRTH": "16/7/1974", "ADDR_STATE": "NC", "ADDR_POSTAL_CODE": "257609", "ENTITY_TYPE": "TEST", "GENDER": "F", "srccode": "MDMPER", "RECORD_ID": "151110080", "DSRC_ACTION": "A", "ADDR_CITY": "Raleigh", "DRIVERS_LICENSE_NUMBER": "95", "PHONE_NUMBER": "984-881-8384", "NAME_LAST": "OBERMOELLER", "entityid": "151110080", "ADDR_LINE1": "3802 eBllevue RD", "DATA_SOURCE": "TEST"}
{"SOCIAL_HANDLE": "battlesa", "ADDR_STATE": "LA", "ADDR_POSTAL_CODE": "70706", "NAME_FIRST": "DEVIN", "ENTITY_TYPE": "TEST", "GENDER": "M", "srccode": "MDMPER", "CC_ACCOUNT_NUMBER": "5018608175414044187", "RECORD_ID": "151267101", "DSRC_ACTION": "A", "ADDR_CITY": "Denham Springs", "DRIVERS_LICENSE_NUMBER": "614557601", "PHONE_NUMBER": "318-398-0649", "NAME_LAST": "LOVELL", "entityid": "151267101", "ADDR_LINE1": "8487 Ashley ", "DATA_SOURCE": "TEST"}
`
var testBadData string = `{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000001", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
{"DATA_SOURCE": "ICIJ", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
{"RECORD_ID": "24000001", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "ANNEX FREDERICK & SHIRLEY STS, P.O. BOX N-4805, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000001"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000002", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "SUITE E-2,UNION COURT BUILDING, P.O. BOX N-8188, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000002"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000003", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, LYFORD CAY, P.O. BOX N-7785, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000003"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000004", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "P.O. BOX N-3708 BAHAMAS FINANCIAL CENTRE, P.O. BOX N-3708 SHIRLEY & CHARLOTTE STS, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000004"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000005", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, 3RD FLOOR, LYFORD CAY, P.O. BOX N-3024, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000005"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000005B" "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "LYFORD CAY HOUSE, 3RD FLOOR, LYFORD CAY, P.O. BOX N-3024, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000005"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000006", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "303 SHIRLEY STREET, P.O. BOX N-492, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000006"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000007", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "OCEAN CENTRE, MONTAGU FORESHORE, P.O. BOX SS-19084 EAST BAY STREET, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000007"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000008", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "PROVIDENCE HOUSE, EAST WING EAST HILL ST, P.O. BOX CB-12399, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000008"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000009", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "BAYSIDE EXECUTIVE PARK, WEST BAY & BLAKE, P.O. BOX N-4875, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000009"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000010", "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "GROUND FLOOR, GOODMAN'S BAY CORPORATE CE, P.O. BOX N 3933, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000010"}
{"DATA_SOURCE": "ICIJ", "RECORD_ID": "24000010B" "ENTITY_TYPE": "ADDRESS", "RECORD_TYPE": "ADDRESS", "icij_source": "BAHAMAS", "icij_type": "ADDRESS", "COUNTRIES": [{"COUNTRY_OF_ASSOCIATION": "BHS"}], "ADDR_FULL": "GROUND FLOOR, GOODMAN'S BAY CORPORATE CE, P.O. BOX N 3933, NASSAU, BAHAMAS", "REL_ANCHOR_DOMAIN": "ICIJ_ID", "REL_ANCHOR_KEY": "24000010"}
{"SOCIAL_HANDLE": "shuddersv", "DATE_OF_BIRTH": "16/7/1974", "ADDR_STATE": "NC", "ADDR_POSTAL_CODE": "257609", "ENTITY_TYPE": "TEST", "GENDER": "F", "srccode": "MDMPER", "RECORD_ID": "151110080", "DSRC_ACTION": "A", "ADDR_CITY": "Raleigh", "DRIVERS_LICENSE_NUMBER": "95", "PHONE_NUMBER": "984-881-8384", "NAME_LAST": "OBERMOELLER", "entityid": "151110080", "ADDR_LINE1": "3802 eBllevue RD", "DATA_SOURCE": "TEST"}
{"SOCIAL_HANDLE": "battlesa", "ADDR_STATE": "LA", "ADDR_POSTAL_CODE": "70706", "NAME_FIRST": "DEVIN", "ENTITY_TYPE": "TEST", "GENDER": "M", "srccode": "MDMPER", "CC_ACCOUNT_NUMBER": "5018608175414044187", "RECORD_ID": "151267101", "DSRC_ACTION": "A", "ADDR_CITY": "Denham Springs", "DRIVERS_LICENSE_NUMBER": "614557601", "PHONE_NUMBER": "318-398-0649", "NAME_LAST": "LOVELL", "entityid": "151267101", "ADDR_LINE1": "8487 Ashley ", "DATA_SOURCE": "TEST"}
`
