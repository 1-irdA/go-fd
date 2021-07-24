package finder

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fatih/color"
)

type find struct {
	path     string
	searched string
	nbFiles  uint32
	opts     Options
	wg       sync.WaitGroup
	elapsed  time.Duration
}

func New(path string, searched string, options Options) *find {
	checkParamsAndOpts(path, searched, options)
	return &find{path: path, searched: searched, opts: options}
}

func (f *find) Find() {
	defer f.details()
	f.checkRegIfNeed()
	f.launch()
}

func (f *find) launch() {
	start := time.Now()
	_, err := os.Lstat(f.path)
	if err != nil {
		printErr("Invalid path")
	}
	f.wg.Add(1)
	f.worker(f.path)
	f.wg.Wait()
	f.elapsed = time.Since(start)
}

func (f *find) worker(dirPath string) {
	defer f.wg.Done()
	file, oErr := os.Open(dirPath)

	if oErr != nil {
		abs, _ := filepath.Abs(dirPath)
		printErr("cannot access to " + abs)
	}

	defer file.Close()

	dirs, rErr := file.ReadDir(-1)

	if rErr != nil {
		abs, _ := filepath.Abs(dirPath)
		printErr("cannot access to " + abs)
	}

	for _, entry := range dirs {
		atomic.AddUint32(&f.nbFiles, 1)
		if entry.IsDir() && !isHidden(filepath.Join(dirPath, entry.Name())) {
			if f.correspond(entry) {
				f.printPath(dirPath, entry, color.New(color.FgBlue))
			}
			f.wg.Add(1)
			go f.worker(path.Join(dirPath, entry.Name()))
		} else if f.correspond(entry) {
			f.printPath(dirPath, entry, color.New(color.FgGreen))
		}
	}
}

func (f *find) details() {
	if f.opts.Bench {
		color.Yellow("Files browsed %d, search duration : %v\n", f.nbFiles, f.elapsed)
	}
}

func (f *find) checkRegIfNeed() {
	if f.opts.Regex {
		if _, err := regexp.Compile(f.searched); err != nil {
			printErr("Invalid regex : " + f.searched)
		}
	}
}

func (f *find) correspond(entry fs.DirEntry) (result bool) {
	if f.opts.File && f.opts.Dir {
		result = f.checkName(entry)
	} else if f.opts.File && !entry.IsDir() {
		result = f.checkName(entry)
	} else if f.opts.Dir && entry.IsDir() {
		result = f.checkName(entry)
	}
	return
}

func (f *find) checkName(entry fs.DirEntry) (result bool) {
	if f.opts.Regex {
		match, err := regexp.MatchString(f.searched, entry.Name())
		if err != nil {
			printErr("Regex error")
		}
		result = match
	} else if strings.Contains(entry.Name(), f.searched) {
		result = true
	}
	return
}

func (f *find) printPath(dirPath string, entry fs.DirEntry, c *color.Color) {
	if f.opts.Absolute {
		c.Println(path.Join(dirPath, entry.Name()))
	} else if path, err := filepath.Rel(f.path, path.Join(dirPath, entry.Name())); err == nil {
		c.Println(path)
	}
}

func checkParamsAndOpts(path, searched string, opts Options) {
	if path == "" {
		printErr("Needs location to search")
	} else if searched == "" {
		printErr("Needs value to search")
	} else if !opts.Dir && !opts.File {
		printErr("Needs to search files [-f], folders [-d] or both")
	}
}

func isHidden(path string) bool {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}

	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false
	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}

func printErr(msg string) {
	color.Red(msg)
	os.Exit(1)
}
