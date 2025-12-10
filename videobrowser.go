package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	baseDir  = "indir" // base directory
	maxDepth = 4       // max depth from baseDir
	prgname  = "videobrowser"
	version  = "v0.1"
	httpip   = "localhost"
	httpport = "8900"
)

type Entry struct {
	Name string
	Path string
	Type string // "file" or "dir"
}

type PageData struct {
	CurrentPath string
	Entries     []Entry
	FileContent string
	IsFileView  bool
	ParentPath  string
	IsImage     bool
	IsVideo     bool
}

var tmpl = template.Must(template.ParseFiles("template.html"))

// ---------------- Safe path join ----------------
func safeJoin(base, rel string) (string, error) {
	cleaned := filepath.Clean(filepath.Join(base, rel))
	absBase, _ := filepath.Abs(base)
	absCleaned, _ := filepath.Abs(cleaned)

	if !strings.HasPrefix(absCleaned, absBase) {
		return "", os.ErrPermission
	}
	return cleaned, nil
}

// ---------------- Parent path (never above root) ----------------
func parentPath(relPath string) string {
	cleaned := filepath.Clean(relPath)
	if cleaned == "." || cleaned == "" {
		return "." // already at root
	}
	parent := filepath.Dir(cleaned)
	if parent == "." {
		return "." // cannot go above root
	}
	return parent
}

// ---------------- List directory entries with max depth ----------------
func listDir(relPath string, depth int) ([]Entry, error) {
	if depth > maxDepth {
		return []Entry{}, nil
	}

	full, err := safeJoin(baseDir, relPath)
	if err != nil {
		return nil, err
	}

	items, err := os.ReadDir(full)
	if err != nil {
		return nil, err
	}

	var result []Entry
	for _, it := range items {
		entry := Entry{
			Name: it.Name(),
			Path: filepath.Join(relPath, it.Name()),
		}
		if it.IsDir() {
			entry.Type = "dir"
		} else {
			entry.Type = "file"
		}
		result = append(result, entry)
	}
	return result, nil
}

// ---------------- HTTP handler ----------------
func handler(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	if rel == "" {
		rel = "." // root
	}

	fullPath, err := safeJoin(baseDir, rel)
	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	data := PageData{CurrentPath: rel, ParentPath: parentPath(rel)}

	if info.IsDir() {
		entries, err := listDir(rel, 1)
		if err != nil {
			http.Error(w, "Error reading directory", http.StatusInternalServerError)
			return
		}

		// Sort directories first
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Type == entries[j].Type {
				return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
			}
			return entries[i].Type == "dir"
		})

		data.Entries = entries
		data.IsFileView = false
	} else {
		data.IsFileView = true
		ext := strings.ToLower(filepath.Ext(fullPath))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif", ".bmp":
			data.IsImage = true
		case ".mp4", ".webm", ".ogg":
			data.IsVideo = true
		case ".zip":
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(fullPath))
			http.ServeFile(w, r, fullPath)
			return
		default:
			content, err := os.ReadFile(fullPath)
			if err != nil {
				http.Error(w, "Could not read file", http.StatusInternalServerError)
				return
			}
			data.FileContent = string(content)
		}
	}

	// Serve media files directly
	if info.Mode().IsRegular() && (data.IsImage || data.IsVideo) {
		http.ServeFile(w, r, fullPath)
		return
	}

	tmpl.Execute(w, data)
}

// ---------------- Main ----------------
func main() {
	// ensure baseDir exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		log.Fatalf("Base directory %s does not exist", baseDir)
	}

	log.Printf("%s %s\n",prgname,version)
	log.Printf("Listening on http://%s:%s\n",httpip,httpport)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+httpport, nil))
}

