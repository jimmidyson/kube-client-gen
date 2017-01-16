package immutables

import (
	"go/types"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func javaPackage(rootPackage, openshiftRootPackage, pkgPath string) (string, string, string) {
	pkgPath = strings.TrimPrefix(pkgPath, "github.com/openshift/origin/vendor/")

	if strings.HasPrefix(pkgPath, "github.com/openshift/origin/pkg/") {
		goAPIPackage := strings.TrimPrefix(pkgPath, "github.com/openshift/origin/pkg/")
		strippedPackage := strings.Replace(goAPIPackage, "/api", "", -1)
		splitPkg := strings.Split(strippedPackage, "/")
		return openshiftRootPackage + "." + strings.Replace(strippedPackage, "/", ".", -1), strings.Join([]string{splitPkg[len(splitPkg)-2], splitPkg[len(splitPkg)-1]}, "-"), "openshift"
	}
	if strings.HasPrefix(pkgPath, "k8s.io/kubernetes") {
		goAPIPackage := strings.TrimPrefix(strings.TrimPrefix(pkgPath, "k8s.io/kubernetes/pkg/"), "k8s.io/kubernetes/federation/")
		splitPkg := strings.Split(goAPIPackage, "/")
		if len(splitPkg) >= 2 {
			return rootPackage + "." + strings.Replace(goAPIPackage, "/", ".", -1), strings.Join([]string{splitPkg[len(splitPkg)-2], splitPkg[len(splitPkg)-1]}, "-"), "kubernetes"
		}
	}
	return "", "", ""
}

func javaPackageToDir(rootDir, moduleName, javaPackage string) string {
	return filepath.Join(
		rootDir,
		moduleName, "src", "main", "java",
		strings.Replace(javaPackage, ".", string(filepath.Separator), -1),
	)
}

func javaType(rootPackage, openshiftRootPackage string, typ types.Type, typeName string) (string, error) {
	typeName = strings.TrimPrefix(typeName, "github.com/openshift/origin/vendor/")
	switch fldT := typ.Underlying().(type) {
	case *types.Slice:
		elemType, err := javaType(rootPackage, openshiftRootPackage, fldT.Elem(), fldT.Elem().String())
		if err != nil {
			return "", err
		}
		return "java.util.List<" + elemType + ">", nil
	case *types.Map:
		keyType, err := javaType(rootPackage, openshiftRootPackage, fldT.Key(), fldT.Key().String())
		if err != nil {
			return "", err
		}
		elemType, err := javaType(rootPackage, openshiftRootPackage, fldT.Elem(), fldT.Elem().String())
		if err != nil {
			return "", err
		}
		return "java.util.Map<" + keyType + ", " + elemType + ">", nil
	case *types.Struct:
		switch typeName {
		case "k8s.io/kubernetes/pkg/runtime.RawExtension":
			return "io.fabric8.kubernetes.types.api.v1.HasMetadata", nil
		case "k8s.io/kubernetes/pkg/api/unversioned.Time":
			return "java.util.Date", nil
		case "k8s.io/kubernetes/pkg/util/intstr.IntOrString":
			return "io.fabric8.kubernetes.types.common.IntOrString", nil
		default:
			javaPkg, _, _ := javaPackage(rootPackage, openshiftRootPackage, typeName)
			return javaPkg, nil
		}
	case *types.Pointer:
		return javaType(rootPackage, openshiftRootPackage, fldT.Elem(), fldT.Elem().String())
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
