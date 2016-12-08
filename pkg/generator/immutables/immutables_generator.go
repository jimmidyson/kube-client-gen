package immutables

import (
	"fmt"
	"io"
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

	"encoding/xml"

	"github.com/jimmidyson/kube-client-gen/pkg/generator"
	"github.com/jimmidyson/kube-client-gen/pkg/loader"
)

const immutableTemplateText = `package {{.JavaPackage}};
{{if .Doc}}
{{comment .Doc ""}}{{end}}
@org.immutables.value.Value.Immutable
{{$fieldsLen := len .Fields}}{{if len .Fields}}@com.fasterxml.jackson.annotation.JsonInclude(value=com.fasterxml.jackson.annotation.JsonInclude.Include.NON_EMPTY, content=com.fasterxml.jackson.annotation.JsonInclude.Include.NON_NULL)
@com.fasterxml.jackson.annotation.JsonPropertyOrder({
{{range $i, $f := .Fields}}{{if eq 0 $i}} {{end}}{{if lt 0 (len $f.Name)}} "{{$f.Name}}"{{if isNotLastField $i $fieldsLen}},{{end}}{{end}}{{end}}
}){{end}}
@com.fasterxml.jackson.databind.annotation.JsonSerialize(as = Immutable{{.ClassName}}.class)
@com.fasterxml.jackson.databind.annotation.JsonDeserialize(as = Immutable{{.ClassName}}.class)
public abstract class {{.ClassName}} implements {{if .HasMetadata}}io.fabric8.kubernetes.types.api.v1.HasMetadata, {{end}}io.fabric8.kubernetes.types.common.WithValidation {{"{"}}{{$className := .ClassName}}{{$goPackage := .GoPackage}}{{range .Fields}}
{{if .Doc}}
{{comment .Doc "  "}}{{end}}{{if eq .Name ""}}
  @com.fasterxml.jackson.annotation.JsonUnwrapped{{else}}
  @com.fasterxml.jackson.annotation.JsonProperty("{{.Name}}"){{end}}{{if typeName .Type | ne "TypeMeta"}}{{if eq .Type "java.util.Date"}}
  @com.fasterxml.jackson.databind.annotation.JsonDeserialize(using = io.fabric8.kubernetes.types.common.RFC3339DateDeserializer.class)
  @com.fasterxml.jackson.annotation.JsonFormat(shape = com.fasterxml.jackson.annotation.JsonFormat.Shape.STRING, pattern = io.fabric8.kubernetes.types.common.RFC3339DateDeserializer.RFC3339_FORMAT, timezone="UTC"){{end}}
  {{$optional := isOptional $className (typeName .Type) .Optional $fieldsLen}}{{validationConstraints $className .Name $optional}}public abstract {{if $optional}}java.util.Optional<{{end}}{{.Type}}{{if $optional}}>{{end}} {{if eq .Type "Boolean"}}is{{else}}get{{end}}{{if .Name}}{{upperFirst .Name | sanitize}}{{else}}{{typeName .Type | upperFirst | sanitize}}{{end}}();{{else}}
  @org.immutables.value.Value.Derived
  public {{.Type}} get{{typeName .Type}}() {
    return new {{.Type}}.Builder().kind("{{$className}}").apiVersion("{{apiVersion $goPackage}}").build();
  }

	@com.fasterxml.jackson.annotation.JsonIgnore
  @org.immutables.value.Value.Derived
  public String getApiVersion() {
    return getTypeMeta().getApiVersion();
  }

	@com.fasterxml.jackson.annotation.JsonIgnore
  @org.immutables.value.Value.Derived
  public String getKind() {
    return getTypeMeta().getKind();
  }{{end}}{{end}}

	public static class Builder extends Immutable{{.ClassName}}.Builder {}

}
`

const modulePomTemplateText = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>

  <parent>
    <groupId>{{.GroupID}}</groupId>
    <artifactId>{{.ParentArtifactID}}</artifactId>
    <version>{{.Version}}</version>
  </parent>

  <artifactId>{{.ArtifactID}}</artifactId>{{if lt 0 (len .Dependencies)}}{{$parent := .}}

  <dependencies>{{range .Dependencies}}
    <dependency>
      <groupId>{{$parent.GroupID}}</groupId>
      <artifactId>{{.}}</artifactId>
      <version>${project.version}</version>
    </dependency>{{end}}
  </dependencies>{{end}}

</project>
`

var startOfLineRegexp = regexp.MustCompile(`(?m:^)`)

var immutableTemplate = template.Must(template.New("immutable").
	Funcs(
		template.FuncMap{
			"isNotLastField": func(currentIndex, numFields int) bool {
				return currentIndex < (numFields - 1)
			},
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
			"isOptional": func(className, fieldType string, optional bool, numFields int) bool {
				return className != "TypeMeta" && fieldType != "ObjectMeta" && optional && numFields > 1
			},
			"validationConstraints": func(className, fieldName string, optional bool) string {
				switch className {
				case "ObjectMeta":
					switch fieldName {
					case "namespace":
						return `@javax.validation.Valid
	@javax.validation.constraints.Size(max = 253)
	@javax.validation.constraints.Pattern(regexp = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$")
  `
					}
				case "EnvVar":
					switch fieldName {
					case "name":
						return `@javax.validation.Valid
	@javax.validation.constraints.Pattern(regexp = "^[A-Za-z_][A-Za-z0-9_]*$")
  `
					}
				case "Container", "Volume", "ContainePort", "ContainerStatus", "ServicePort", "EndpointPort":
					switch fieldName {
					case "name":
						return `@javax.validation.Valid
	@javax.validation.constraints.Size(max = 63)
	@javax.validation.constraints.Pattern(regexp = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")
  `
					}
				}
				return `@javax.validation.Valid
	`
			},
		},
	).
	Parse(immutableTemplateText))

var modulePomTemplate = template.Must(template.New("modulePOM").Parse(modulePomTemplateText))

func New(c Config) generator.Generator {
	c.Logger.Debug("creating generator", "type", "immutables")
	return &immutablesGenerator{
		config: c,
	}
}

type Config struct {
	generator.Config

	JavaRootPackage          string
	JavaRootOpenShiftPackage string
	StyleClass               string
}

type immutablesGenerator struct {
	config Config
}

var _ generator.Generator = &immutablesGenerator{}

func (g *immutablesGenerator) Generate(pkgs []loader.Package) error {
	g.config.Logger.Debug("generating")

	p, err := parsePOM(filepath.Join(g.config.OutputDirectory, "pom.xml"))
	if err != nil {
		return errors.Wrap(err, "failed to parse parent POM")
	}

	for _, pkg := range pkgs {
		dependencies := []string{"common"}

		depMap := map[string]struct{}{}
		javaPkg, moduleName, platform := javaPackage(g.config.JavaRootPackage, g.config.JavaRootOpenShiftPackage, pkg.Path)
		moduleName = platform + "-" + moduleName
		pkgDir := javaPackageToDir(g.config.OutputDirectory, moduleName, javaPkg)
		g.config.Logger.Debug("generating for package", "package", pkg.Path, "javaPackage", javaPkg, "dir", pkgDir)

		if err := os.MkdirAll(pkgDir, 0700); err != nil && os.IsExist(err) {
			return errors.Wrapf(err, "failed to create directory %s", pkgDir)
		}

		if err := g.writePackageJava(pkgDir, javaPkg, g.config.StyleClass, pkg.Doc); err != nil {
			return errors.Wrap(err, "failed to write package-info.java file")
		}

		for _, typ := range pkg.Types {
			if typ.Name == "Time" {
				continue
			}

			fp := filepath.Join(pkgDir, typ.Name+".java")

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

			for _, fld := range typ.Fields {
				_, moduleDep, platform := javaPackage(g.config.JavaRootPackage, g.config.JavaRootOpenShiftPackage, fld.TypeName)
				moduleDep = strings.Split(moduleDep, ".")[0]
				if len(moduleDep) > 0 && platform+"-"+moduleDep != moduleName && moduleDep != "util-intstr" {
					moduleDep = platform + "-" + moduleDep
					depMap[moduleDep] = struct{}{}
				}
			}
		}

		for k := range depMap {
			dependencies = append(dependencies, k)
		}

		if err := g.writeModulePOM(filepath.Join(g.config.OutputDirectory, moduleName), p.GroupID, moduleName, p.ArtifactID, p.Version, dependencies); err != nil {
			return errors.Wrap(err, "failed to write module POM file")
		}
	}

	return nil
}

type field struct {
	Type     string
	Name     string
	Doc      string
	Optional bool
}

type data struct {
	JavaPackage string
	GoPackage   string
	ClassName   string
	HasMetadata bool
	Doc         string
	Fields      []field
}

func (g *immutablesGenerator) write(pkg string, typ loader.Type, f io.WriteCloser) error {
	defer func() {
		_ = f.Close()
	}()

	fields := make([]field, 0, len(typ.Fields))

	hasMetadata := false
	hasTypemeta := false
	for _, fld := range typ.Fields {
		javaType, err := javaType(g.config.JavaRootPackage, g.config.JavaRootOpenShiftPackage, fld.Type, fld.TypeName)
		if err != nil {
			return errors.Wrapf(err, "unhandled field type %s for field %s.%s.%s", pkg, typ.Name, fld.Type.String())
		}

		if fld.JSONProperty == "metadata" && fld.Type.String() == "k8s.io/kubernetes/pkg/api/v1.ObjectMeta" {
			hasMetadata = true
		}

		if fld.Type.String() == "k8s.io/kubernetes/pkg/api/unversioned.TypeMeta" {
			hasTypemeta = true
		}

		fields = append(fields, field{javaType, fld.JSONProperty, fld.Doc, !fld.JSONRequired})
	}

	return immutableTemplate.Execute(f, data{
		JavaPackage: pkg,
		GoPackage:   typ.Package,
		ClassName:   typ.Name,
		HasMetadata: hasMetadata && hasTypemeta,
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

func (g *immutablesGenerator) writeModulePOM(moduleDir, groupID, artifactID, parentArtifactID, version string, dependencies []string) error {
	type params struct {
		GroupID          string
		ArtifactID       string
		ParentArtifactID string
		Version          string
		Dependencies     []string
	}

	f, err := os.Create(filepath.Join(moduleDir, "pom.xml"))
	if err != nil {
		return errors.Wrap(err, "failed to create module POM file")
	}
	defer func() { _ = f.Close() }() // #nosec

	return modulePomTemplate.Execute(f, params{
		GroupID:          groupID,
		ArtifactID:       artifactID,
		ParentArtifactID: parentArtifactID,
		Version:          version,
		Dependencies:     dependencies,
	})
}

type pom struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

func parsePOM(pomPath string) (*pom, error) {
	f, err := os.Open(pomPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open POM at %s", pomPath)
	}
	var p pom
	if err := xml.NewDecoder(f).Decode(&p); err != nil {
		return nil, errors.Wrapf(err, "unable to parse POM at %s", pomPath)
	}
	return &p, nil
}
