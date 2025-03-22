package server

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
)

type templateFS interface {
	fs.FS
	fs.ReadFileFS
}

type Template struct {
	templates map[string]*template.Template
}

func newTemplate(fsys templateFS, root string) *Template {
	templates := make(map[string]*template.Template)
	for name, paths := range extractPaths(fsys, root) {
		templates[name] = template.Must(template.New(name).Funcs(sprig.FuncMap()).ParseFS(fsys, paths...))
	}
	return &Template{templates}
}

func extractPaths(fsys templateFS, root string) map[string][]string {
	r := regexp.MustCompile(`{{.*?template\s+"(?P<dependency>[\w-.]+)".*?}}`)
	dependencyIndex := r.SubexpIndex("dependency")
	dependencyTree := map[string][]string{}

	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		var content []byte
		content, err = fsys.ReadFile(path)
		if err != nil {
			return err
		}

		name := filepath.Base(path)
		var dependencies []string
		for _, match := range r.FindAllStringSubmatch(string(content), -1) {
			dependencies = append(dependencies, match[dependencyIndex])
		}

		dependencyTree[name] = dependencies
		return nil
	})
	if err != nil {
		panic(err.Error())
	}

	resolvedDependencyTree := map[string][]string{}
	var resolve func(string) []string
	resolve = func(name string) []string {
		dependencies := []string{name}
		for _, dependency := range dependencyTree[name] {
			subDependencies, ok := resolvedDependencyTree[dependency]
			if !ok {
				subDependencies = resolve(dependency)
			}
			dependencies = append(dependencies, subDependencies...)
		}

		resolvedDependencyTree[name] = dependencies
		return dependencies
	}

	dependencyPaths := make(map[string][]string)

	for name, _ := range dependencyTree {
		dependencies := slices.Compact(resolve(name))
		paths := make([]string, len(dependencies))
		for i, file := range dependencies {
			paths[i] = filepath.Join(root, file)
		}
		dependencyPaths[name] = paths
	}

	return dependencyPaths

}

func (t *Template) Render(writer io.Writer, name string, data any, c echo.Context) error {
	templates, ok := t.templates[name]
	if !ok {
		return fmt.Errorf("unknown template %s", name)
	}
	return templates.ExecuteTemplate(writer, name, data)
}
