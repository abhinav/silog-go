package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/doc/comment"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type example struct {
	Name       string
	Doc        *comment.Doc
	Code       string
	Screenshot string
}

func main() {
	log := log.New(os.Stderr, "", 0)
	if err := run(log); err != nil {
		log.Fatal(err)
	}
}

func run(log *log.Logger) error {
	dir := flag.String("dir", "examples", "directory containing example subdirectories")
	out := flag.String("out", "examples.md", "output markdown file")
	flag.Parse()

	entries, err := os.ReadDir(*dir)
	if err != nil {
		return fmt.Errorf("reading examples directory: %w", err)
	}

	if err := os.MkdirAll("./img/examples", 0o755); err != nil {
		return fmt.Errorf("creating img/examples directory: %w", err)
	}

	var examples []*example
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		ex, err := loadExample(log, *dir, entry.Name())
		if err != nil {
			return fmt.Errorf("loading example %s: %w", entry.Name(), err)
		}
		examples = append(examples, ex)
	}

	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Name < examples[j].Name
	})

	f, err := os.Create(*out)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	if err := generateMarkdown(f, examples); err != nil {
		return fmt.Errorf("generating markdown: %w", err)
	}

	return nil
}

func loadExample(log *log.Logger, dir, exampleDir string) (*example, error) {
	mainPath := filepath.Join(dir, exampleDir, "main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return nil, fmt.Errorf("reading main.go: %w", err)
	}

	// Parse the file to get doc comment
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, mainPath, content, parser.ParseComments|parser.PackageClauseOnly)
	if err != nil {
		return nil, fmt.Errorf("parsing main.go: %w", err)
	}

	var docComment *comment.Doc
	if file.Doc != nil {
		text := file.Doc.Text()
		var p comment.Parser
		docComment = p.Parse(text)
	}

	// Extract code blocks
	lines := strings.Split(string(content), "\n")
	var codeParts []string
	var currentBlock []string
	inBlock := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "// <EXAMPLE>" {
			inBlock = true
			currentBlock = nil
		} else if strings.TrimSpace(line) == "// </EXAMPLE>" {
			if inBlock {
				codeParts = append(codeParts, dedent(currentBlock))
				currentBlock = nil
			}
			inBlock = false
		} else if inBlock {
			currentBlock = append(currentBlock, line)
		}
	}
	code := strings.Join(codeParts, "\n\n")

	// Run freeze command
	cmd := exec.Command("freeze", "--config", "freeze.json",
		"--execute", "go run ./examples/"+exampleDir,
		"--output", filepath.Join("img", "examples", exampleDir+".svg"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("running freeze: %w", err)
	}

	return &example{
		Name:       exampleDir,
		Doc:        docComment,
		Code:       code,
		Screenshot: filepath.Join("img", "examples", exampleDir+".svg"),
	}, nil
}

func dedent(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	// Find minimum indentation
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := 0
		for _, r := range line {
			if r == ' ' || r == '\t' {
				indent++
			} else {
				break
			}
		}
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// Remove common indentation
	var result []string
	for _, line := range lines {
		if len(line) >= minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func generateMarkdown(w io.Writer, examples []*example) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	fmt.Fprintln(bw, "# Examples")
	fmt.Fprintln(bw)

	commentPrinter := &comment.Printer{
		HeadingLevel: 2,
		HeadingID: func(*comment.Heading) string {
			return "" // GitHub-default headings
		},
	}

	for _, ex := range examples {
		if ex.Doc != nil {
			md := commentPrinter.Markdown(ex.Doc)
			fmt.Fprintln(bw, string(md))
		}

		if ex.Code != "" {
			fmt.Fprintln(bw)
			fmt.Fprintln(bw, "```go")
			fmt.Fprintln(bw, ex.Code)
			fmt.Fprintln(bw, "```")
			fmt.Fprintln(bw)
		}

		fmt.Fprintf(bw, "![screenshot of %s](%s)\n\n", ex.Name, ex.Screenshot)
	}

	return nil
}
