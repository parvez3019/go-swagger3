package gomod

import (
	"os"
	"path/filepath"
	"sort"
	"unicode"

	log "github.com/sirupsen/logrus"

	"github.com/parvez3019/go-swagger3/parser/model"
	"golang.org/x/mod/modfile"
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

// Parse builds an index from go.mod that maps every required module's import path
// to its directory in the module cache.
//
// The original implementation eagerly walked the full directory tree of every
// `require` entry, registering every sub-package. Since Go 1.24's `tool` directive a
// module's go.mod also lists every transitive dependency of its tools as an indirect
// require (hundreds of modules the documented service never references). Walking them
// all (and parsing their ASTs in the API phase) dominated runtime by minutes.
//
// Instead we record only path -> cache-directory here (no walking) and let the schema
// parser locate and index a dependency package the first time an annotation actually
// references a type from it. Every module stays resolvable, but only referenced ones
// are ever touched.
func (p *parser) Parse() error {
	log.Info("Parsing GoMod Info ...")
	b, err := os.ReadFile(p.GoModFilePath)
	if err != nil {
		return err
	}
	file, err := modfile.Parse(p.GoModFilePath, b, nil)
	if err != nil {
		return err
	}

	mods := make([]model.DepModule, 0, len(file.Require))
	for i := range file.Require {
		mods = append(mods, model.DepModule{
			ImportPath: file.Require[i].Mod.Path,
			CacheDir:   moduleCacheDir(p.GoModCachePath, file.Require[i].Mod.Path, file.Require[i].Mod.Version),
		})
	}
	// Longest import path first so prefix resolution picks the most specific module
	// (e.g. github.com/foo/bar/v2 before github.com/foo/bar).
	sort.Slice(mods, func(i, j int) bool { return len(mods[i].ImportPath) > len(mods[j].ImportPath) })
	p.DepModules = mods

	if p.RunInDebugMode {
		p.Debugf("go.mod module index: %d modules (resolved to cache dirs on demand)", len(p.DepModules))
		for i := range p.DepModules {
			p.Debugf("module %s -> %s", p.DepModules[i].ImportPath, p.DepModules[i].CacheDir)
		}
	}

	return nil
}

// moduleCacheDir reproduces the module-cache path encoding: an uppercase letter in
// the module path is escaped as "!" followed by its lowercase form.
func moduleCacheDir(cacheRoot, pkgName, version string) string {
	pathRunes := []rune{}
	for _, v := range pkgName {
		if !unicode.IsUpper(v) {
			pathRunes = append(pathRunes, v)
			continue
		}
		pathRunes = append(pathRunes, '!')
		pathRunes = append(pathRunes, unicode.ToLower(v))
	}
	return filepath.Join(cacheRoot, string(pathRunes)+"@"+version)
}
