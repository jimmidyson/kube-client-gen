package immutables

import (
	"go/types"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func javaPackage(rootPackage, pkgPath string) string {
	// lastDotIndex := strings.LastIndex(pkgPath, ".")
	// lastSlashIndex := strings.LastIndex(pkgPath, "/")
	// if 0 < lastDotIndex && lastSlashIndex < lastDotIndex {
	// 	pkgPath = pkgPath[:lastDotIndex+1] + "Abstract" + pkgPath[lastDotIndex+1:]
	// }

	var javaSubPackage string
	if strings.HasPrefix(pkgPath, "github.com/openshift/origin/pkg/") {
		goApiPackage := strings.TrimPrefix(pkgPath, "github.com/openshift/origin/pkg/")
		javaSubPackage = "openshift.types." + strings.Replace(strings.Replace(goApiPackage, "/api", "", -1), "/", ".", -1)
	} else {
		goApiPackage := strings.TrimPrefix(pkgPath, "k8s.io/kubernetes/pkg/")
		javaSubPackage = "kubernetes.types." + strings.Replace(goApiPackage, "/", ".", -1)
	}

	return rootPackage + "." + javaSubPackage
}

func javaPackageToDir(rootDir, javaPackage string) string {
	return filepath.Join(
		rootDir,
		strings.Replace(javaPackage, ".", string(filepath.Separator), -1),
	)
}

func javaType(rootPackage string, typ types.Type, typeName string) (string, error) {
	switch fldT := typ.Underlying().(type) {
	case *types.Slice:
		elemType, err := javaType(rootPackage, fldT.Elem(), fldT.Elem().String())
		if err != nil {
			return "", err
		}
		return "java.util.List<" + elemType + ">", nil
	case *types.Map:
		keyType, err := javaType(rootPackage, fldT.Key(), fldT.Key().String())
		if err != nil {
			return "", err
		}
		elemType, err := javaType(rootPackage, fldT.Elem(), fldT.Elem().String())
		if err != nil {
			return "", err
		}
		return "java.util.Map<" + keyType + ", " + elemType + ">", nil
	case *types.Struct:
		switch typeName {
		case "k8s.io/kubernetes/pkg/runtime.RawExtension":
			return "io.fabric8.kubernetes.types.api.HasMetadata", nil
		case "k8s.io/kubernetes/pkg/api/unversioned.Time":
			return "java.util.Date", nil
		default:
			return javaPackage(rootPackage, typeName), nil
		}
	case *types.Pointer:
		return javaType(rootPackage, fldT.Elem(), fldT.Elem().String())
	case *types.Basic:
		return javaTypeBasic(fldT.Kind()), nil
	default:
		return "", errors.Errorf("unknown field type %s", fldT.String())
	}
}

func javaTypeBasic(kind types.BasicKind) string {
	switch kind {
	case types.Bool:
		return "Boolean"
	case types.Int, types.Int8, types.Int16, types.Int32,
		types.Uint, types.Uint8, types.Uint16, types.Uint32:
		return "Integer"
	case types.Int64, types.Uint64:
		return "Long"
	case types.String:
		return "String"
	case types.Float32, types.Float64:
		return "Float"
	default:
		return ""
	}
}
