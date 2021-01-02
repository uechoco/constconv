package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	typeNames    = flag.String("type", "", "comma-separated list of type names; must be set")
	templateFile = flag.String("template", "", "template file path which is relative from the srcdir; must be set")
	templateData = flag.String("data", "", "semicolon-separated list of extra template data. each data coupled by equal sign; e.g. \"typename=Foo;prefix=Bar\"")
	outputFile   = flag.String("output", "", "output file name; default srcdir/<snake-cased-type>_constconv.go")
	buildTags    = flag.String("tags", "", "comma-separated list of build tags to apply")
)

// Usage is a replacement usage function for the flag package.
func Usage() {
	_, _ = fmt.Fprintf(os.Stderr, `Usage of constconv:
	constconv -type T -template F [optional flags] [directory]
	constconv -type T -template F [optional flags] files... # Must be a single package
`) // TODO: print go.dev package URL
	flag.PrintDefaults()
}

func main() {
	log.SetPrefix("constconv: ")
	flag.Usage = Usage
	flag.Parse()

	config := preprocess()

	gen := NewGenerator(config)
	if err := gen.LoadTemplate(); err != nil {
		log.Fatalf("loading template output: %s", err)
	}

	parser := NewParser(config)
	if err := parser.Parse(); err != nil {
		log.Fatalf("parsing output: %s", err)
	}

	src, err := gen.Generate(parser)
	if err != nil {
		log.Fatalf("generating output: %s", err)
	}

	if err := ioutil.WriteFile(config.outputFile, src, 0644); err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

type Config struct {
	types        []string          // The list of type names. e.g.: ["UserStatus", "os.FileMode"]
	tags         []string          // The list of build tags.
	extraData    map[string]string // The map of extra template data
	execArgsStr  string            // The argument string executed
	dirOrFiles   []string          // The arguments
	baseDir      string            // The base directory
	templateFile string            // The template file name.
	outputFile   string            // The output file name.
}

// preprocess processes flags and initializes its related. It fatal-exit if error occurred.
func preprocess() *Config {
	config := &Config{}

	if len(*typeNames) == 0 || len(*templateFile) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	config.types = strings.Split(*typeNames, ",")

	if len(*buildTags) > 0 {
		config.tags = strings.Split(*buildTags, ",")
	}

	// convert "typename=Foo;prefix=Bar" into {"typename":"Foo","prefix":"Bar"}
	config.extraData = make(map[string]string)
	if len(*templateData) > 0 {
		kvPairs := strings.Split(*templateData, ";")
		for _, kvPair := range kvPairs {
			kv := strings.SplitN(kvPair, "=", 2)
			if len(kv) != 2 {
				log.Fatalf("invalid extraData: %s", kvPair)
			}
			config.extraData[kv[0]] = kv[1]
		}
	}

	// accept either one directory or a list of files.
	config.dirOrFiles = flag.Args()
	if len(config.dirOrFiles) == 0 {
		config.dirOrFiles = []string{"."}
	}

	dir, dirSpecified, err := detectDirectory(config.dirOrFiles)
	if err != nil {
		log.Fatal(err)
	} else if dirSpecified && len(config.tags) != 0 {
		log.Fatal("-tags option applies only to directories, not when files are specified")
	}
	config.baseDir = dir

	config.outputFile = *outputFile
	if config.outputFile == "" {
		underscoreTypeName := strings.Replace(config.types[0], ".", "_", 1)
		baseName := fmt.Sprintf("%s_constconv.go", underscoreTypeName)
		config.outputFile = filepath.Join(config.baseDir, strings.ToLower(baseName))
	}

	config.templateFile = filepath.Join(config.baseDir, *templateFile)

	config.execArgsStr = strings.Join(os.Args[1:], " ")

	return config
}

// detectDirectory detects the base directory by arguments.
func detectDirectory(dirOrFiles []string) (dir string, dirSpecified bool, err error) {
	if len(dirOrFiles) == 1 {
		if isDir, err := isDirectory(dirOrFiles[0]); err != nil {
			return "", false, err
		} else if isDir {
			dir = dirOrFiles[0]
			return dir, true, nil
		}
	}
	dir = filepath.Dir(dirOrFiles[0])
	return dir, false, nil
}

// Helper

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) (bool, error) {
	info, err := os.Stat(name)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
