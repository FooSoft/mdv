package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/FooSoft/goldsmith"
	"github.com/FooSoft/goldsmith-components/devserver"
	"github.com/FooSoft/goldsmith-components/plugins/livejs"
	"github.com/FooSoft/goldsmith-components/plugins/markdown"
	"github.com/toqueteos/webbrowser"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type builder struct {
	port     int
	browsing bool
}

func (self *builder) Build(contentDir, buildDir, cacheDir string) {
	log.Print("building...")

	gm := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Typographer),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)

	errs := goldsmith.Begin(contentDir).
		Clean(true).
		Chain(markdown.NewWithGoldmark(gm)).
		Chain(livejs.New()).
		End(buildDir)

	for _, err := range errs {
		log.Print(err)
	}

	if !self.browsing {
		webbrowser.Open(fmt.Sprintf("http://127.0.0.1:%d", self.port))
		self.browsing = true
	}
}

func main() {
	port := flag.Int("port", 8080, "port")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("unexpected number of arguments")
	}

	requestPath := flag.Arg(0)
	buildDir, err := ioutil.TempDir("", "mvd-*")
	if err != nil {
		log.Fatal(err)
	}

	info, err := os.Stat(requestPath)
	if err != nil {
		log.Fatal(err)
	}

	contentDir := requestPath
	if !info.IsDir() {
		contentDir = filepath.Dir(requestPath)
	}

	go func() {
		b := &builder{port: *port}
		devserver.DevServe(b, *port, contentDir, buildDir, "")
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		<-sigs
		log.Println("terminating...")
		break
	}
	if err := os.RemoveAll(buildDir); err != nil {
		log.Fatal(err)
	}
}
