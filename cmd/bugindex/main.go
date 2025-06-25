package main

import (
    "bufio"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strings"
)

type Bug struct {
    ID       string
    Title    string
    Severity string
    Status   string
    Path     string
}

var (
    idRe       = regexp.MustCompile(`(?i)^id:\s*(.+)`)
    titleRe    = regexp.MustCompile(`(?i)^title:\s*(.+)`)
    severityRe = regexp.MustCompile(`(?i)^severity:\s*(.+)`)
    statusRe   = regexp.MustCompile(`(?i)^status:\s*(.+)`)
)

func main() {
    bugsDir := []string{"docs/bugs", "docs/bugs-phase-02"}
    var bugs []Bug

    for _, dir := range bugsDir {
        filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
            if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
                return nil
            }
            bug, ok := parseBugFile(path)
            if ok {
                bugs = append(bugs, bug)
            }
            return nil
        })
    }

    // Sort by ID
    sort.Slice(bugs, func(i, j int) bool { return bugs[i].ID < bugs[j].ID })

    // Counters
    counts := map[string]int{"open": 0, "fixed": 0, "in progress": 0}
    for _, b := range bugs {
        key := strings.ToLower(b.Status)
        counts[key]++
    }

    // Print summary header
    fmt.Printf("# Bug Index (auto-generated)\n\n")
    fmt.Printf("| ID | Title | Severity | Status | File |\n")
    fmt.Printf("|----|-------|----------|--------|------|\n")
    for _, b := range bugs {
        fmt.Printf("| %s | %s | %s | %s | %s |\n", b.ID, b.Title, b.Severity, b.Status, b.Path)
    }
    fmt.Printf("\n**Totals**: Open=%d, In-Progress=%d, Fixed=%d (Total=%d)\n",
        counts["open"], counts["in progress"], counts["fixed"], len(bugs))
}

func parseBugFile(path string) (Bug, bool) {
    f, err := os.Open(path)
    if err != nil {
        return Bug{}, false
    }
    defer f.Close()

    var bug Bug
    bug.Path = path
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if bug.ID == "" {
            if m := idRe.FindStringSubmatch(line); len(m) == 2 {
                bug.ID = m[1]
                continue
            }
        }
        if bug.Title == "" {
            if m := titleRe.FindStringSubmatch(line); len(m) == 2 {
                bug.Title = m[1]
                continue
            }
        }
        if bug.Severity == "" {
            if m := severityRe.FindStringSubmatch(line); len(m) == 2 {
                bug.Severity = m[1]
                continue
            }
        }
        if bug.Status == "" {
            if m := statusRe.FindStringSubmatch(line); len(m) == 2 {
                bug.Status = m[1]
                continue
            }
        }
        if bug.ID != "" && bug.Title != "" && bug.Severity != "" && bug.Status != "" {
            break
        }
    }
    ok := bug.ID != "" && bug.Title != "" && bug.Status != ""
    return bug, ok
} 