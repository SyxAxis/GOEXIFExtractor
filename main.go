//
// This tool scans folder trees for JPGs, dumps the EXIF information from images into a CSV.
//
// Example command-line:
//
//   EXIFExtractor -filepath <source-file-path> -csvfile <csvfile.csv>
//
//
//  WARNING : This is the first proper GOlang util I've written! It's very basic and probably very buggy!
//
//
//	Standard EXIF fields
// 	ImageDescription,Make,Model,Software,DateTime,Artist,Copyright,ExposureTime,FNumber,ISOSpeedRatings,DateTimeOriginal,DateTimeDigitized,FocalLength,CameraOwnerName,BodySerialNumber,LensModel,GPSLatitudeRef,GPSLatitude,GPSLongitudeRef,GPSLongitude
//
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"io/ioutil"

	"github.com/dsoprea/go-exif"
	log "github.com/dsoprea/go-logging"
)

var (
	filepathArg     = ""
	csvFileResults  = ""
	userEXIFFields  = ""
	printLoggingArg = false

	constEXIFFields = []string{
		"ImageDescription",
		"Make",
		"Model",
		"Software",
		"DateTime",
		"Artist",
		"Copyright",
		"ExposureTime",
		"FNumber",
		"ISOSpeedRatings",
		"DateTimeOriginal",
		"DateTimeDigitized",
		"FocalLength",
		"CameraOwnerName",
		"BodySerialNumber",
		"LensModel",
		"GPSLatitudeRef",
		"GPSLatitude",
		"GPSLongitudeRef",
		"GPSLongitude",
	}
)

// IfdEntry JSON def struct
type IfdEntry struct {
	IfdPath     string                `json:"ifd_path"`
	FqIfdPath   string                `json:"fq_ifd_path"`
	IfdIndex    int                   `json:"ifd_index"`
	TagID       uint16                `json:"tag_id"`
	TagName     string                `json:"tag_name"`
	TagTypeID   exif.TagTypePrimitive `json:"tag_type_id"`
	TagTypeName string                `json:"tag_type_name"`
	UnitCount   uint32                `json:"unit_count"`
	Value       interface{}           `json:"value"`
	ValueString string                `json:"value_string"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// ================================================================================
//
// ================================================================================
func getFileExif(filepathArg string) []IfdEntry {

	f, err := os.Open(filepathArg)
	log.PanicIf(err)

	data, err := ioutil.ReadAll(f)
	log.PanicIf(err)

	rawExif, err := exif.SearchAndExtractExif(data)
	log.PanicIf(err)

	// Run the parse.

	im := exif.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()

	entries := make([]IfdEntry, 0)
	visitor := func(fqIfdPath string, ifdIndex int, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (err error) {
		defer func() {
			if state := recover(); state != nil {
				err = log.Wrap(state.(error))
				log.Panic(err)
			}
		}()

		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)
		log.PanicIf(err)

		it, err := ti.Get(ifdPath, tagId)
		if err != nil {
			if log.Is(err, exif.ErrTagNotFound) {
				// fmt.Printf("WARNING: Unknown tag: [%s] (%04x)\n", ifdPath, tagId)
				return nil
			}
			log.Panic(err)
		}

		valueString := ""
		var value interface{}
		if tagType.Type() == exif.TypeUndefined {
			var err error
			value, err = valueContext.Undefined()
			if err != nil {
				if err == exif.ErrUnhandledUnknownTypedTag {
					value = nil
				} else {
					log.Panic(err)
				}
			}

			valueString = fmt.Sprintf("%v", value)
		} else {
			valueString, err = valueContext.FormatFirst()
			log.PanicIf(err)

			value = valueString
		}

		entry := IfdEntry{
			IfdPath:     ifdPath,
			FqIfdPath:   fqIfdPath,
			IfdIndex:    ifdIndex,
			TagID:       tagId,
			TagName:     it.Name,
			TagTypeID:   tagType.Type(),
			TagTypeName: tagType.Name(),
			UnitCount:   valueContext.UnitCount(),
			Value:       value,
			ValueString: valueString,
		}

		entries = append(entries, entry)

		return nil
	}

	_, err = exif.Visit(exif.IfdStandard, im, ti, rawExif, visitor)
	log.PanicIf(err)

	return entries

}

// ================================================================================
// Check the EXIF field you want against the fields you got from the current file
// ================================================================================
func crossCheckEXIFArrayToRequest(fileEXIFData []IfdEntry, wipEXIFFieldList []string) string {

	var csvEXIFData []string

	for _, EXIFFieldName := range wipEXIFFieldList {

		fndFieldVal := ""
		for _, fileEXIFDataRec := range fileEXIFData {
			// fileEXIFDataRec.TagName
			// fileEXIFDataRec.ValueString

			if EXIFFieldName == fileEXIFDataRec.TagName {
				// Odd data kludges!!

				// anything with a trailing "/1" means whole number not fraction
				if strings.HasSuffix(fileEXIFDataRec.ValueString, "/1") {
					fndFieldVal = strings.TrimSuffix(fileEXIFDataRec.ValueString, "/1")
				} else {
					fndFieldVal = fileEXIFDataRec.ValueString
				}

				switch fileEXIFDataRec.TagName {
				case "FocalLength":
					fndFieldVal = fndFieldVal + " mm"
				case "ExposureTime":

					if strings.Contains(fndFieldVal, "/") && !strings.Contains(fndFieldVal, "1/") {
						flen01, err := strconv.ParseFloat(strings.Split(fndFieldVal, "/")[0], 64)
						if err != nil {
							log.Panic("Cannot convert the source fstop left operator.")
						}
						flen02, err := strconv.ParseFloat(strings.Split(fndFieldVal, "/")[1], 64)
						if err != nil {
							log.Panic("Cannot convert the source fstop right operator.")
						}
						fndFieldVal = fmt.Sprintf("%.1f", flen01/flen02)
					}
					fndFieldVal = fndFieldVal + " sec"

				case "FNumber":
					// if there's a slash it's a ratio in fracitonal form and we need it in a decimal without trailing zeros
					if strings.Contains(fndFieldVal, "/") {
						flen01, err := strconv.ParseFloat(strings.Split(fndFieldVal, "/")[0], 64)
						if err != nil {
							log.Panic("Cannot convert the source fstop left operator.")
						}
						flen02, err := strconv.ParseFloat(strings.Split(fndFieldVal, "/")[1], 64)
						if err != nil {
							log.Panic("Cannot convert the source fstop right operator.")
						}
						fndFieldVal = fmt.Sprintf("%.1f", flen01/flen02)
					}
					fndFieldVal = "f/" + fndFieldVal
				}
			}
		}
		csvEXIFData = append(csvEXIFData, fndFieldVal)
	}

	csvEXIFDataLine := ""
	for _, exifChkData := range csvEXIFData {
		csvEXIFDataLine = csvEXIFDataLine + "," + "\"" + exifChkData + "\""
	}

	return csvEXIFDataLine
}

// ================================================================================
// get a list of JPG files, it will walk the entire tree from the root you give it
// ================================================================================
func getListFiles(filepathArg string) []string {

	var files []string

	err := filepath.Walk(filepathArg, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".jpg" || filepath.Ext(path) == ".JPG" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return files
}

// ================================================================================
// main()
// ================================================================================
func main() {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintErrorf(err, "Program error.")
			os.Exit(1)
		}
	}()

	flag.StringVar(&filepathArg, "filepath", "", "Root folder to scan for JPG image files.")
	flag.StringVar(&csvFileResults, "csvfile", "", "CSV output file to hold results.")
	flag.StringVar(&userEXIFFields, "exif-fields", "", "User selected EXIF fields.")
	flag.BoolVar(&printLoggingArg, "verbose", false, "Print logging.")

	flag.Parse()

	if filepathArg == "" {
		fmt.Printf("Please use [-filepath] option to specify a path to scan for JPG files.\n")
		os.Exit(1)
	}

	if csvFileResults == "" {
		fmt.Printf("Please use [-csvfile] option to specify an output file.\n")
		os.Exit(1)
	}

	var wipEXIFFieldList []string
	if userEXIFFields == "" {
		wipEXIFFieldList = constEXIFFields
	} else {
		wipEXIFFieldList = strings.Split(userEXIFFields, ",")
	}

	if printLoggingArg == true {
		cla := log.NewConsoleLogAdapter()
		log.AddAdapter("console", cla)
	}

	// -------------------------------------------------------------
	// create a new CSV file
	// -------------------------------------------------------------
	outCSVFile, err := os.Create(csvFileResults)
	check(err)

	// -------------------------------------------------------------
	// add CSV file header line
	// -------------------------------------------------------------
	csvHeaderField := ""
	for _, wipEXIFFieldName := range wipEXIFFieldList {
		if csvHeaderField == "" {
			csvHeaderField = "\"" + wipEXIFFieldName + "\""
		} else {
			csvHeaderField = csvHeaderField + ",\"" + wipEXIFFieldName + "\""
		}
	}
	outFileInfo := []byte("\"Filename\",\"FolderPath\"," + csvHeaderField + "\n")
	_, err = outCSVFile.Write(outFileInfo)
	check(err)

	// -------------------------------------------------------------
	// scan the source folder for jpg files
	//   get a string array of filepaths
	// -------------------------------------------------------------
	files := getListFiles(filepathArg)

	// -------------------------------------------------------------
	// loop through the string array of files
	// -------------------------------------------------------------
	for _, jpgFile := range files {

		fmt.Printf("Getting info for [%s]...\n", jpgFile)

		// -------------------------------------------------------------
		// get the EXIF data for the current image file
		// -------------------------------------------------------------
		entries := getFileExif(jpgFile)

		// -------------------------------------------------------------
		// split the path and filename
		// -------------------------------------------------------------
		jpgFilePath := filepath.Dir(jpgFile)
		jpgFileName := filepath.Base(jpgFile)

		// -------------------------------------------------------------
		// write these into the CSV output file
		// -------------------------------------------------------------
		outFileInfo := []byte("\"" + jpgFileName + "\"" + "," + "\"" + jpgFilePath + "\"")
		_, err := outCSVFile.Write(outFileInfo)
		check(err)

		// -------------------------------------------------------------
		// check the EXIF data from the image file
		//   against the core list of EXIF fields we want
		// -------------------------------------------------------------
		csvEXIFDataLine := crossCheckEXIFArrayToRequest(entries, wipEXIFFieldList)

		// -------------------------------------------------------------
		// the crosscheck will have passed back a string of CSV data
		// -------------------------------------------------------------
		outFileInfo = []byte(csvEXIFDataLine)
		_, err = outCSVFile.Write(outFileInfo)
		check(err)

		// new line in output file
		outFileInfo = []byte("\n")
		_, err = outCSVFile.Write(outFileInfo)
		check(err)

	}

}
