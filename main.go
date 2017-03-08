package main

import (
	"encoding/json"
	"github.com/nicolasgomollon/peterplanner/parsers"
	"github.com/nicolasgomollon/peterplanner/types"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

const RegistrarPath = "/var/www/registrar/"

type File struct {
	Name string
	Term string
}

func Departments() []string {
	departments := make([]string, 0)
	files, _ := ioutil.ReadDir(RegistrarPath)
	for _, f := range files {
		if f.IsDir() {
			departments = append(departments, f.Name())
		}
	}
	return departments
}

func SchedulesFor(deptDir string, termsMap *map[string]bool) []File {
	deptPath := RegistrarPath + deptDir + "/"
	schedules := make([]File, 0)
	r, _ := regexp.Compile(`soc_(\d{4}-\d{2})\.txt`)
	files, _ := ioutil.ReadDir(deptPath)
	for _, f := range files {
		if !f.IsDir() {
			filename := f.Name()
			if strings.HasPrefix(filename, "soc_") {
				term := r.FindStringSubmatch(filename)[1]
				(*termsMap)[term] = true
				schedules = append(schedules, File{Name: filename, Term: term})
			}
		}
	}
	return schedules
}

func ParseCatalogue(deptDir string, courses *map[string]types.Course) {
	deptPath := RegistrarPath + deptDir + "/"
	filepath := deptPath + "catalogue.html"
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return
	}
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	responseHTML := string(b)
	parsers.ParseCatalogue(responseHTML, courses)
}

func ParsePrerequisites(deptDir string, courses *map[string]types.Course) {
	deptPath := RegistrarPath + deptDir + "/"
	filepath := deptPath + "prereqs.html"
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return
	}
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	responseHTML := string(b)
	parsers.ParsePrerequisites(responseHTML, courses)
}

func ParseWebSOC(deptDir string, courses *map[string]types.Course, termsMap *map[string]bool) {
	deptPath := RegistrarPath + deptDir + "/"
	files := SchedulesFor(deptDir, termsMap)
	for _, file := range files {
		b, err := ioutil.ReadFile(deptPath + file.Name)
		if err != nil {
			panic(err)
		}
		responseTXT := string(b)
		parsers.ParseWebSOC(file.Term, responseTXT, courses)
	}
}

func main() {
	termsMap := make(map[string]bool, 0)
	catalogue := types.Catalogue{}
	courses := make(map[string]types.Course, 0)
	for _, deptDir := range Departments() {
		ParseCatalogue(deptDir, &courses)
		ParsePrerequisites(deptDir, &courses)
		ParseWebSOC(deptDir, &courses, &termsMap)
	}
	catalogue.Courses = courses
	
	terms := make([]string, len(termsMap))
	i := 0
	for t := range termsMap {
		terms[i] = t
		i++
	}
	sort.Sort(sort.Reverse(sort.StringSlice(terms)))
	catalogue.Terms = terms
	
	exportJSON, err := json.Marshal(catalogue)
	if err != nil {
		panic(err)
	}
	
	filepath := RegistrarPath + "catalogue.json"
	err = ioutil.WriteFile(filepath, exportJSON, 0644)
	if err != nil {
		panic(err)
	}
}
