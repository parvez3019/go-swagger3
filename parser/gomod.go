package parser

import (
	"github.com/mikunalpha/go-module"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func (p *parser) parseGoMod() error {
	b, err := ioutil.ReadFile(p.GoModFilePath)
	if err != nil {
		return err
	}
	goMod, err := module.Parse(b)
	if err != nil {
		return err
	}
	for i := range goMod.Requires {
		pathRunes := []rune{}
		for _, v := range goMod.Requires[i].Path {
			if !unicode.IsUpper(v) {
				pathRunes = append(pathRunes, v)
				continue
			}
			pathRunes = append(pathRunes, '!')
			pathRunes = append(pathRunes, unicode.ToLower(v))
		}
		pkgName := goMod.Requires[i].Path
		pkgPath := filepath.Join(p.GoModCachePath, string(pathRunes)+"@"+goMod.Requires[i].Version)
		pkgName = filepath.ToSlash(pkgName)
		p.KnownPkgs = append(p.KnownPkgs, pkg{
			Name: pkgName,
			Path: pkgPath,
		})
		p.KnownNamePkg[pkgName] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		p.KnownPathPkg[pkgPath] = &p.KnownPkgs[len(p.KnownPkgs)-1]

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
				name := filepath.Join(pkgName, strings.TrimPrefix(path, pkgPath))
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
		filepath.Walk(pkgPath, walker)
	}
	if p.Debug {
		for i := range p.KnownPkgs {
			p.debug(p.KnownPkgs[i].Name, "->", p.KnownPkgs[i].Path)
		}
	}
	return nil
}
