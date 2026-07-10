package module

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/parvez3019/go-swagger3/parser/model"
	log "github.com/sirupsen/logrus"
)

type Parser interface {
	Parse() error
}

type parser struct {
	model.Utils
}

func NewParser(utils model.Utils) Parser {
	return &parser{
		Utils: utils,
	}
}

// Parse parse sub-package
func (p *parser) Parse() error {
	log.Info("Parsing Modules ...")
	walker := func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			if strings.HasPrefix(strings.Trim(strings.TrimPrefix(path, p.ModulePath), "/"), ".git") {
				return nil
			}
			if p.isExcluded(path) {
				return filepath.SkipDir
			}
			fns, err := filepath.Glob(filepath.Join(path, "*.go"))
			if len(fns) == 0 || err != nil {
				return nil
			}
			// p.debug(path)
			name := filepath.Join(p.ModuleName, strings.TrimPrefix(path, p.ModulePath))
			name = filepath.ToSlash(name)
			p.KnownPkgs = append(p.KnownPkgs, model.Pkg{
				Name: name,
				Path: path,
			})
			p.KnownNamePkg[name] = &p.KnownPkgs[len(p.KnownPkgs)-1]
			p.KnownPathPkg[path] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		}
		return nil
	}
	return filepath.Walk(p.ModulePath, walker)
}

func (p *parser) isExcluded(path string) bool {
	if len(p.ExcludePaths) == 0 {
		return false
	}
	rel := strings.Trim(strings.TrimPrefix(path, p.ModulePath), string(filepath.Separator))
	relSlash := filepath.ToSlash(rel)
	pathSlash := filepath.ToSlash(path)
	for _, exclude := range p.ExcludePaths {
		exclude = filepath.ToSlash(strings.TrimSpace(exclude))
		if exclude == "" {
			continue
		}
		if relSlash == exclude || strings.HasPrefix(relSlash, exclude+"/") ||
			pathSlash == exclude || strings.HasPrefix(pathSlash, exclude+"/") ||
			strings.Contains(relSlash, exclude) {
			return true
		}
	}
	return false
}
