package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/ahmedalhulaibi/substance-structgen/gostruct"

	"github.com/ahmedalhulaibi/substance"
	"github.com/ahmedalhulaibi/substance/substancegen"
)

func main() {
	helpText := `Usage: 
		gengostruct -db="dbtype" -cnstr="connection:string@locahost:9999 -file="outputFile.go" -fileperobj=false
			or
		gengostruct -jsonsrc="path-to-substance-objects.json -file="outputFile.go""
			or
		gengostruct -file="outputFile.go"	<-- Defaults to -jsonsrc=substance-objects.json
			or
		gengostruct							<-- Default -jsonsrc=substance-objects.json and creates file per obj
`
	dbtype := flag.String("db", "", "Database driver name.\nSupported databases types:\n\t- mysql\n\t- postgres \n\t- sqlite3\n")
	connString := flag.String("cnstr", "", "Connection string to connect to database.")
	jsonSourceFilePath := flag.String("jsonsrc", "substance-objects.json", "JSON substance-objects.json file describing the database objects. This can be used as an alternative to providing connection info.")
	outputSrcFilePath := flag.String("file", "gostruct.go", "File to output source code. If blank outputs to stdout.")
	filePerObjFlag := flag.Bool("fileperobj", true, "Creates a file for each obj/table. Default is true, if false else outputs to gostruct.go (can be overridden using 'file' option)")
	flag.Parse()

	var objects map[string]substancegen.GenObjectType

	if jsonSourceFilePath != nil {
		jsonFile, err := os.Open(*jsonSourceFilePath)
		if err != nil {
			fmt.Printf(helpText)
			log.Panicf(err.Error())
		}
		log.Printf("Opened %s successfully", *jsonSourceFilePath)
		byteVal, _ := ioutil.ReadAll(jsonFile)
		log.Printf("Read %s successfully", *jsonSourceFilePath)
		json.Unmarshal(byteVal, &objects)
		log.Printf("Unmarshalled %s successfully", *jsonSourceFilePath)
	} else if dbtype != nil && connString != nil {
		results, err := substance.DescribeDatabase(*dbtype, *connString)
		if err != nil {
			fmt.Printf(helpText)
			log.Panicf(err.Error())
		}
		if len(results) > 0 {
			log.Println("Database: ", results[0].DatabaseName)
		}
		var tables []string
		for _, result := range results {
			log.Printf("Table: %s\n", result.TableName)
			tables = append(tables, result.TableName)
		}
		log.Println("=====================")

		objects = substancegen.GetObjectTypesFunc(*dbtype, *connString, tables)
	}
	if objects != nil {
		if *filePerObjFlag {
			for key, genobj := range objects {
				objectsTmp := make(map[string]substancegen.GenObjectType, 0)
				objectsTmp[key] = genobj
				var outputBuff bytes.Buffer
				outputBuff.WriteString("package model\n\n")
				gostruct.GenObjectTypeToStructFunc(objectsTmp, &outputBuff)
				err := ioutil.WriteFile(strings.Join([]string{key, ".go"}, ""), outputBuff.Bytes(), 0664)
				if err != nil {
					fmt.Printf(helpText)
					fmt.Printf(outputBuff.String())
				}
			}
		} else {
			log.Println("printing objects")
			log.Println(objects)
			var outputBuff bytes.Buffer
			outputBuff.WriteString("package model\n\n")
			gostruct.GenObjectTypeToStructFunc(objects, &outputBuff)
			err := ioutil.WriteFile(*outputSrcFilePath, outputBuff.Bytes(), 0664)
			if err != nil {
				fmt.Printf(helpText)
				fmt.Printf(outputBuff.String())
			}
		}
	}
}
