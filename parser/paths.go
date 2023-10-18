package parser

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/hanyue2020/go-swagger3/parser/utils"
)

func (p *parser) verifyAndSetPaths() error {
	// check modulePath is exist
	if err := p.verifyModulePath(); err != nil {
		return err
	}

	// check go.mod file is exist
	if err := p.setGoModFilePath(); err != nil {
		return err
	}

	// check mainFilePath is exist
	if err := p.verifyMainFilePath(); err != nil {
		return err
	}

	// get module name from go.mod file
	if err := p.getModuleNameFromGoModFile(); err != nil {
		return err
	}

	// check go module cache path is exist ($GOPATH/pkg/mod)
	if err := p.setGoModCachePath(); err != nil {
		return err
	}

	// check handlerPath is exist
	if err := p.setHandlerPath(); err != nil {
		return err
	}
	return nil
}

func (p *parser) setHandlerPath() error {
	var err error
	if p.HandlerPath != "" {
		p.HandlerPath, err = filepath.Abs(p.HandlerPath)
		if err != nil {
			return err
		}
		_, err := os.Stat(p.HandlerPath)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			return fmt.Errorf("cannot get information of %s: %s", p.HandlerPath, err)
		}
	}
	p.Debugf("handler path: %s", p.HandlerPath)
	return nil
}

func (p *parser) setGoModCachePath() error {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		current, err := user.Current()
		if err != nil {
			return fmt.Errorf("cannot get current user: %s", err)
		}
		goPath = filepath.Join(current.HomeDir, "go")
	}
	goModCachePath := filepath.Join(goPath, "pkg", "mod")
	goModCacheInfo, err := os.Stat(goModCachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("cannot get information of %s: %s", goModCachePath, err)
	}
	if !goModCacheInfo.IsDir() {
		return fmt.Errorf("%s should be a directory", goModCachePath)
	}
	p.GoModCachePath = goModCachePath
	p.Debugf("go module cache path: %s", p.GoModCachePath)
	return nil
}

func (p *parser) getModuleNameFromGoModFile() error {
	moduleName := utils.GetModuleNameFromGoMod(p.GoModFilePath)
	if moduleName == "" {
		return fmt.Errorf("cannot get module name from %s", p.GoModFilePath)
	}
	p.ModuleName = moduleName
	p.Debugf("module name: %s", p.ModuleName)
	return nil
}

func (p *parser) verifyMainFilePath() error {
	if p.MainFilePath == "" {
		fns, err := filepath.Glob(filepath.Join(p.ModulePath, "*.go"))
		if err != nil {
			return err
		}
		for _, fn := range fns {
			if utils.IsMainFile(fn) {
				p.MainFilePath = fn
				break
			}
		}
	} else {
		mainFileInfo, err := os.Stat(p.MainFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			return fmt.Errorf("cannot get information of %s: %s", p.MainFilePath, err)
		}
		if mainFileInfo.IsDir() {
			return fmt.Errorf("mainFilePath should not be a directory")
		}
	}
	p.Debugf("main file path: %s", p.MainFilePath)
	return nil
}

func (p *parser) setGoModFilePath() error {
	goModFilePath := filepath.Join(p.ModulePath, "go.mod")
	goModFileInfo, err := os.Stat(goModFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("cannot get information of %s: %s", goModFilePath, err)
	}
	if goModFileInfo.IsDir() {
		return fmt.Errorf("%s should be a file", goModFilePath)
	}
	p.GoModFilePath = goModFilePath
	p.Debugf("go.mod file path: %s", p.GoModFilePath)
	return err
}

func (p *parser) verifyModulePath() error {
	var err error
	p.ModulePath, err = filepath.Abs(p.ModulePath)
	if err != nil {
		return err
	}
	moduleInfo, err := os.Stat(p.ModulePath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("cannot get information of %s: %s", p.ModulePath, err)
	}
	if !moduleInfo.IsDir() {
		return fmt.Errorf("modulePath should be a directory")
	}
	p.Debugf("module path: %s", p.ModulePath)
	return err
}
