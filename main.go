package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

var exportedJsonPath string
var originalJsonPath string

func init() {
	flag.StringVar(&exportedJsonPath, "e", "", "Specify the exported jsonfile")
	flag.StringVar(&originalJsonPath, "o", "", "Specify the original jsonfile")
}

func main() {
	flag.Parse()

	if exportedJsonPath == "" || originalJsonPath == "" {
		PrintHelp()
		os.Exit(1)
	}

	var exportedExpressions []string

	exported, err := ioutil.ReadFile(exportedJsonPath)
	if err != nil {
		log.Printf("Could not open exported json file \n%v", err)
	}
	var exportedJsonContent map[string]interface{}
	err = json.Unmarshal(exported, &exportedJsonContent)
	if err != nil {
		log.Printf("Could not unmarshal to interface\n%v", err)
	}
	for _, panel := range exportedJsonContent["panels"].([]interface{}) {
		resultMap := panel.(map[string]interface{})
		for k, _ := range resultMap {
			if k == "targets" {
				targets := resultMap[k].([]interface{})
				for _, target := range targets {
					nestedResultMap := target.(map[string]interface{})
					for e, _ := range nestedResultMap {
						if e == "expr" {
							exportedExpressions = append(exportedExpressions, nestedResultMap[e].(string))
						}
					}
				}
			}
		}
	}

	original, err := ioutil.ReadFile(originalJsonPath)
	if err != nil {
		log.Printf("Could not open original json file\n%v", err)
	}
	err, jsonFromYaml := GetJsonFromYaml(original)
	if err != nil {
		log.Printf("Could not get json from original yaml\n%v", err)
	}

	var originalJsonContent map[string]interface{}

	err = json.Unmarshal(jsonFromYaml, &originalJsonContent)
	if err != nil {
		log.Printf("Could not unmarshal to interface|\n%v", err)
	}

	for _, panel := range originalJsonContent["panels"].([]interface{}) {
		resultMap := panel.(map[string]interface{})
		for k, _ := range resultMap {
			if k == "targets" {
				targets := resultMap[k].([]interface{})
				for _, target := range targets {
					jsonInTarget := target.(map[string]interface{})
					go func () {
						jsonInTarget["expr"] = <-returnExpressions(exportedExpressions)
					}()
				}
			}
		}
	}
	marshalled, err := json.MarshalIndent(originalJsonContent, "", "    ")
	if err != nil {
		log.Printf("Could not marshal exportedJsoncontent\n%v", err)
	}
	fmt.Println(string(marshalled))

}

func returnExpressions(expressionList []string) <-chan string {
	ch := make(chan string)
	for _, expression := range expressionList {
		ch <- expression
	}
	return ch
}

func PrintHelp() {
	fmt.Println("You have to specify -e for exported json")
	fmt.Println("You have to specify -o for output json")
	fmt.Println(os.Args[0], "-e <exported-json-path> -o <original-yaml-containing-json-path>")
}

func GetJsonFromYaml(data []byte) (error, []byte) {
	originalYamlContent := make(map[interface{}]interface{})
	err := yaml.Unmarshal(data, &originalYamlContent)
	if err != nil {
		return err, nil
	}
	spec := originalYamlContent["spec"].(map[interface{}]interface{})
	return nil, []byte(spec["json"].(string))
}