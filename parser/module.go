package parser

import (
	"github.com/parvez3019/go-swagger3/logger"
	"os"
	"path/filepath"
	"strings"
)

type ModuleParser interface {
	ParseModule() error
}

type moduleParser struct {
	Path
	*PkgAndSpecs
	*logger.Logger
}

func NewModuleParser(path Path, specs *PkgAndSpecs, logger *logger.Logger) ModuleParser {
	return &moduleParser{
		Path:        path,
		PkgAndSpecs: specs,
		Logger:      logger,
	}
}

func (p *moduleParser) ParseModule() error {
	walker := func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			if strings.HasPrefix(strings.Trim(strings.TrimPrefix(path, p.ModulePath), "/"), ".git") {
				return nil
			}
			fns, err := filepath.Glob(filepath.Join(path, "*.go"))
			if len(fns) == 0 || err != nil {
				return nil
			}
			// p.debug(path)
			name := filepath.Join(p.ModuleName, strings.TrimPrefix(path, p.ModulePath))
			name = filepath.ToSlash(name)
			p.KnownPkgs = append(p.KnownPkgs, pkg{
				Name: name,
				Path: path,
			})
			p.KnownNamePkg[name] = &p.KnownPkgs[len(p.KnownPkgs)-1]
			p.KnownPathPkg[path] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		}
		return nil
	}
	filepath.Walk(p.ModulePath, walker)
	return nil
}
