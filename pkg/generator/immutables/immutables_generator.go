package immutables

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/jimmidyson/kube-client-gen/pkg/generator"
	"github.com/jimmidyson/kube-client-gen/pkg/loader"
)

const immutableTemplateText = `package {{.JavaPackage}};
import org.immutables.value.Value;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonUnwrapped;
{{if .Doc}}{{comment .Doc ""}}{{end}}
@Value.Immutable
abstract class Abstract{{.ClassName}} {{if .HasMetadata}}implements io.fabric8.kubernetes.types.api.HasMetadata {{end}}{{"{"}}{{$className := .ClassName}}{{$goPackage := .GoPackage}}{{range .Fields}}
{{if .Doc}}{{comment .Doc "  "}}{{end}}{{if eq .Name ""}}
  @JsonUnwrapped{{else}}
  @JsonProperty("{{.Name}}"){{end}}{{if typeName .Type | ne "TypeMeta"}}{{if eq .Type "java.util.Date"}}
  @com.fasterxml.jackson.databind.annotation.JsonDeserialize(using = io.fabric8.kubernetes.types.api.RFC3339DateDeserializer.class)
  @com.fasterxml.jackson.annotation.JsonFormat(shape = com.fasterxml.jackson.annotation.JsonFormat.Shape.STRING, pattern = io.fabric8.kubernetes.types.api.RFC3339DateDeserializer.RFC3339_FORMAT, timezone="UTC"){{end}}
  public abstract {{.Type}} {{if eq .Type "Boolean"}}is{{else}}get{{end}}{{if .Name}}{{upperFirst .Name | sanitize}}{{else}}{{typeName .Type | upperFirst | sanitize}}{{end}}();{{else}}
  @Value.Derived
  public {{.Type}} get{{typeName .Type}}() {
    return {{.Type}}.of("{{$className}}", "{{apiVersion $goPackage}}");
  }
  @Value.Derived
  public String getApiVersion() {
    return getTypeMeta().getApiVersion();
  }
  @Value.Derived
  public String getKind() {
    return getTypeMeta().getKind();
  }{{end}}
{{end}}
}
`

var startOfLineRegexp = regexp.MustCompile(`(?m:^)`)

var immutableTemplate = template.Must(template.New("immutable").
	Funcs(
		template.FuncMap{
			"comment": func(doc string, indent string) string {
				return indent + "/*\n" + startOfLineRegexp.ReplaceAllString(doc, indent+" * ") + "\n" + indent + " */"
			},
			"typeName": func(s string) string {
				lastDotIndex := strings.LastIndex(s, ".")
				if lastDotIndex >= 0 {
					return s[lastDotIndex+1:]
				}
				return s
			},
			"packageName": func(s string) string {
				lastDotIndex := strings.LastIndex(s, ".")
				if lastDotIndex >= 0 {
					return s[:lastDotIndex]
				}
				return s
			},
			"upperFirst": func(s string) string {
				if s == "" {
					return ""
				}
				r, n := utf8.DecodeRuneInString(s)
				return string(unicode.ToUpper(r)) + s[n:]
			},
			"apiVersion": func(s string) string {
				apiVersion := path.Base(s)
				apiGroup := path.Base(strings.TrimSuffix(s, apiVersion))
				if apiGroup != "api" {
					apiVersion = apiGroup + "/" + apiVersion
				}
				return apiVersion
			},
			"sanitize": func(s string) string {
				res := ""
				splitRes := strings.Split(s, ".")
				for i, spl := range splitRes {
					if i > 0 {
						if len(spl) > 0 {
							r, n := utf8.DecodeRuneInString(spl)
							spl = string(unicode.ToUpper(r)) + spl[n:]
						}
					}
					res += spl
				}
				return res
			},
		},
	).
	Parse(immutableTemplateText))

const styleClassName = "ImmutablesStyle"

func New(c Config) generator.Generator {
	c.Logger.Debug("creating generator", "type", "immutables")
	return &immutablesGenerator{
		config: c,
	}
}

type Config struct {
	generator.Config

	JavaRootPackage string
}

type immutablesGenerator struct {
	config Config
}

var _ generator.Generator = &immutablesGenerator{}

func (g *immutablesGenerator) Generate(pkgs []loader.Package) error {
	g.config.Logger.Debug("generating")

	stylesPackage := g.config.JavaRootPackage + ".kubernetes.types.api"
	pkgDir := javaPackageToDir(g.config.OutputDirectory, stylesPackage)
	if err := os.MkdirAll(pkgDir, 0755); err != nil && os.IsExist(err) {
		return errors.Wrapf(err, "failed to create directory %s", pkgDir)
	}

	stylesClass := stylesPackage + "." + styleClassName

	for _, pkg := range pkgs {
		javaPkg := javaPackage(g.config.JavaRootPackage, pkg.Path)
		pkgDir := javaPackageToDir(g.config.OutputDirectory, javaPkg)
		g.config.Logger.Debug("generating for package", "package", pkg.Path, "javaPackage", javaPkg, "dir", pkgDir)

		if err := os.MkdirAll(pkgDir, 0755); err != nil && os.IsExist(err) {
			return errors.Wrapf(err, "failed to create directory %s", pkgDir)
		}

		if err := g.writePackageJava(pkgDir, javaPkg, stylesClass, pkg.Doc); err != nil {
			return errors.Wrap(err, "failed to write package-info.java file")
		}

		for _, typ := range pkg.Types {
			if typ.Name == "Time" {
				continue
			}

			fp := filepath.Join(pkgDir, "Abstract"+typ.Name+".java")

			if !g.config.Force {
				_, err := os.Stat(fp)
				if err == nil {
					return errors.Errorf("target file %s already exists", fp)
				}
				if !os.IsNotExist(err) {
					return errors.Errorf("failed to check if target file %s exists: %v", fp, err)
				}
			}

			f, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return errors.Wrapf(err, "failed to open file %s to write", fp)
			}

			if err := g.write(javaPkg, typ, f); err != nil {
				return errors.Wrapf(err, "failed to write file %s", fp)
			}
		}
	}

	return nil
}

func (g *immutablesGenerator) write(pkg string, typ loader.Type, f *os.File) error {
	defer func() {
		_ = f.Close()
	}()

	type field struct {
		Type string
		Name string
		Doc  string
	}

	type data struct {
		JavaPackage string
		GoPackage   string
		ClassName   string
		HasMetadata bool
		Doc         string
		Fields      []field
	}

	fields := make([]field, 0, len(typ.Fields))

	hasMetadata := false
	for _, fld := range typ.Fields {
		javaType, err := javaType(g.config.JavaRootPackage, fld.Type, fld.TypeName)
		if err != nil {
			return errors.Wrapf(err, "unhandled field type %s for field %s.%s.%s", pkg, typ.Name, fld.Type.String())
		}

		if fld.JSONProperty == "metadata" && fld.Type.String() == "k8s.io/kubernetes/pkg/api/v1.ObjectMeta" {
			hasMetadata = true
		}

		if fld.Type.String() == "k8s.io/kubernetes/pkg/api/unversioned.Time" {
			javaType = "java.util.Date"
		}

		fields = append(fields, field{javaType, fld.JSONProperty, fld.Doc})
	}

	return immutableTemplate.Execute(f, data{
		JavaPackage: pkg,
		GoPackage:   typ.Package,
		ClassName:   typ.Name,
		HasMetadata: hasMetadata,
		Doc:         typ.Doc,
		Fields:      fields,
	})
}

func (g *immutablesGenerator) writePackageJava(pkgDir, javaPackage, styleClass, doc string) error {
	pkgDoc := doc
	if len(pkgDoc) > 0 {
		pkgDoc = startOfLineRegexp.ReplaceAllString(pkgDoc, "// ") + "\n"
	}
	contents := []byte(fmt.Sprintf("%s@%s\npackage %s;\n", pkgDoc, styleClass, javaPackage))
	return ioutil.WriteFile(filepath.Join(pkgDir, "package-info.java"), contents, 0644)
}
