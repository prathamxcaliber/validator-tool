package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/folders", getFolders)
	http.HandleFunc("/files", getFiles)
	http.HandleFunc("/process", handleProcess)
	http.HandleFunc("/compare", serveCompare)

	log.Println("Server running at http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func serveCompare(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/compare.html"))
	tmpl.Execute(w, nil)
}

func getFolders(w http.ResponseWriter, r *http.Request) {
	var folders []string
	base := "EHR"

	entries, err := os.ReadDir(base)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
		}
	}

	json.NewEncoder(w).Encode(folders)
}

func getFiles(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")
	var files []string
	base := filepath.Join("EHR", folder)

	entries, err := ioutil.ReadDir(base)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	json.NewEncoder(w).Encode(files)
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Folder string `json:"folder"`
		Mapper string `json:"mapper"`
		Input  string `json:"input"`
		Custom string `json:"custom"`
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Load the mapper template
	var tmplContent []byte
	var err error
	if req.Custom != "" {
		tmplContent = []byte(req.Custom)
	} else {
		mapperPath := filepath.Join("EHR", req.Folder, req.Mapper)
		tmplContent, err = os.ReadFile(mapperPath)
		if err != nil {
			http.Error(w, "Failed to read mapper file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Register functions
    funcMap := template.FuncMap{
		"sub1": func(x int) int { return x - 1 },
        "formatDate": func(date string) string {
            parsed, err := time.Parse("01/02/2006", date)
            if err != nil {
                return date
            }
            return parsed.Format("2006-01-02")
        },
        "isArray": func(v interface{}) bool {
            _, ok := v.([]interface{})
            return ok
        },
        "getDisplayNameForRaceCode": func(code string) string {
            switch code {
            case "1006-6":
                return "White"
            case "1424-1":
                return "Asian"
            default:
                return "Unknown"
            }
        },
        "getDisplayNameForMaritalCode": func(code string) string {
            switch code {
            case "M":
                return "Married"
            case "S":
                return "Single"
            case "D":
                return "Divorced"
            case "W":
                return "Widowed"
            case "U":
                return "Unknown"
            default:
                return "Other"
            }
        },
        "last": func(index int, arr interface{}) bool {
            switch slice := arr.(type) {
            case []interface{}:
                return index == len(slice)-1
            default:
                return true
            }
        },
    }
    

	// Parse the template
	tmpl, err := template.New("mapper").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the input JSON
	var inputArray []map[string]interface{}
	if err := json.Unmarshal([]byte(req.Input), &inputArray); err != nil {
		http.Error(w, "Invalid input JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(inputArray) == 0 {
		http.Error(w, "Empty input array", http.StatusBadRequest)
		return
	}

	// Execute the template with the first item
	inputData := inputArray[0]
	var output strings.Builder
	if err := tmpl.Execute(&output, inputData); err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(output.String()))
}
