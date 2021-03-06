package mediafs

import (
	"bytes"
	"errors"
	"github.com/majiru/simpleblog/pkg/basicfs"
	"github.com/majiru/simpleblog/pkg/page"
	"github.com/majiru/simpleblog/pkg/webfs"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const dirTemplName = "dir.tmpl"

type mediafs struct {
	root  string
	templ string
}

//NewMediafs creates a new mediafs webfs
func NewMediafs(root string) webfs.Webfs {
	contentDir := filepath.Join(root, "media")
	templ := filepath.Join(root, dirTemplName)
	if _, err := os.Stat(templ); err != nil {
		templ = filepath.Join(root, "..", dirTemplName)
	}
	os.Mkdir(contentDir, 0755)
	return &mediafs{contentDir, templ}
}

func (mfs *mediafs) Stat(request string) (os.FileInfo, error) {
	path := filepath.Join(mfs.root, request)
	return os.Stat(path)
}

func (mfs *mediafs) Read(request string) (io.ReadSeeker, error) {
	path := filepath.Join(mfs.root, request)
	fi, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return os.Open(path)
	}

	out, err := mfs.openDir(request)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (mfs *mediafs) openDir(path string) (io.ReadSeeker, error) {
	files, dirs, err := basicfs.List(filepath.Join(mfs.root, path))
	if err != nil {
		return nil, err
	}

	files = append(files, dirs...)
	p, _ := page.NewPage("File Browser", "/")
	p.Sidebar = make(map[string][]page.Page)
	p.Sidebar["root"] = []page.Page{}
	for _, f := range files {
		listing, _ := page.NewPage(f, filepath.Join(path, f))
		p.Sidebar["root"] = append(p.Sidebar["root"], *listing)
	}

	var out bytes.Buffer
	t, err := template.ParseFiles(mfs.templ)
	if err != nil {
		return nil, errors.New("Template not found")
	}
	if err := t.Execute(&out, p); err != nil {
		return nil, errors.New("Error processing template")
	}
	return strings.NewReader(out.String()), nil
}
