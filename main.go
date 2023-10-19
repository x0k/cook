package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func makePrinter(lines string) RecipeLinesPrinter {
	if strings.HasPrefix(lines, "#!") {
		return CmdRecipeLinesPrinter
	} else {
		return StdRecipeLinesPrinter
	}
}

var fileNames = []string{"recipes", "Recipes", "recipe", "Recipe"}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("No recipe name provided")
	}
	var file *os.File
	var err error
	for _, fileName := range fileNames {
		file, err = os.Open(fileName)
		if err == nil {
			defer file.Close()
			break
		}
	}
	if err != nil {
		log.Fatal("Recipe file not found, allowed file names: ", strings.Join(fileNames, ", "))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	recipeName := os.Args[1]
	collector := NewRecipeLinesCollector(recipeName)
	isRecipeFound, err := collector.CollectLines(scanner)
	if err != nil {
		log.Fatal("Error during collection recipe lines ", err)
	}
	if !isRecipeFound {
		log.Fatalf("Recipe \"%s\" not found ", recipeName)
	}
	lines := collector.GetLines()
	if len(lines) < 1 {
		log.Fatal("Recipe file is empty ")
	}
	printer := makePrinter(lines)
	if err != nil {
		log.Fatal("Error during creating printer ", err)
	}
	err = printer.Print(lines)
	if err != nil {
		log.Fatal("Error during printing ", err)
	}
}
