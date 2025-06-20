package templater

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/azhai/allgo/fsutil"
	"github.com/iancoleman/strcase"
)

var defaultFuncMap = template.FuncMap{
	"Lower":      strings.ToLower,
	"Upper":      strings.ToUpper,
	"Camelize":   strcase.ToCamel,
	"Underscore": strcase.ToSnake,
}

// Factory 模板工厂，可设置模板目录用于扫描.tmpl文件
type Factory struct {
	inclSubFiles bool
	discoverDir  string
	presetTmpls  map[string]*template.Template
	SharedFuncs  template.FuncMap
}

// NewFactory 创建工厂，可选的模板目录
func NewFactory(dir string, inclSubs bool) *Factory {
	if dir != "" {
		dir, _ = filepath.Abs(dir)
	}
	return &Factory{
		inclSubFiles: inclSubs,
		discoverDir:  dir,
		presetTmpls:  make(map[string]*template.Template),
		SharedFuncs:  defaultFuncMap,
	}
}

// UpdateFuncs 增加共享函数
func (f *Factory) UpdateFuncs(funcs template.FuncMap) *Factory {
	for name, fun := range funcs {
		f.SharedFuncs[name] = fun
	}
	return f
}

// AddSubs 加载子模板
func (f *Factory) AddSubs(tmpl *template.Template) *template.Template {
	if f.inclSubFiles == false || tmpl == nil {
		return tmpl
	}
	files, err := filepath.Glob(filepath.Join(f.discoverDir, "sub_*.tmpl"))
	if err == nil && len(files) > 0 {
		tmpl, _ = tmpl.ParseFiles(files...)
	}
	return tmpl
}

// GetTemplate 根据名称获取模板对象，先找已注册模板，再找模板目录下文件
func (f *Factory) GetTemplate(name string, funcs template.FuncMap) *template.Template {
	if tmpl, ok := f.presetTmpls[name]; ok && len(funcs) == 0 {
		return f.AddSubs(tmpl)
	}
	if f.discoverDir == "" {
		return nil
	}
	filename := filepath.Join(f.discoverDir, name+".tmpl")
	if size, _ := fsutil.FileSize(filename); size > 0 {
		tmpl := f.RegisterFile(filename, funcs)
		return f.AddSubs(tmpl)
	}
	return nil
}

// Register 注册模板内容
func (f *Factory) Register(name, content string, funcs template.FuncMap) *template.Template {
	// Funcs adds the elements of the argument map to the template's function map.
	// It must be called before the template is parsed.
	tmpl := template.New(name).Funcs(f.SharedFuncs)
	if len(funcs) > 0 {
		tmpl = tmpl.Funcs(funcs)
	}
	var err error // 有可能所需的funcs不齐全，导致解析失败
	if tmpl, err = tmpl.Parse(content); err == nil {
		f.presetTmpls[name] = tmpl
	}
	return tmpl
}

// RegisterFile 注册模板文件
func (f *Factory) RegisterFile(filename string, funcs template.FuncMap) *template.Template {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}
	name := filepath.Base(filename)
	if extname := filepath.Ext(name); extname != "" {
		name = name[:len(name)-len(extname)]
	}
	return f.Register(name, string(fileData), funcs)
}

// Render 渲染指定模板
func (f *Factory) Render(name string, data any) ([]byte, error) {
	var tmpl *template.Template
	if tmpl = f.GetTemplate(name, nil); tmpl == nil {
		return nil, fmt.Errorf("cannot find the template named %s", name)
	}
	return RenderTemplate(tmpl, data)
}

// RenderTemplate 使用数据渲染模板
func RenderTemplate(tmpl *template.Template, data any) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
