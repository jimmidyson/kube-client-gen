package loader_test

import (
	"go/types"

	"github.com/inconshreveable/log15"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/jimmidyson/kube-client-gen/pkg/loader"
)

func typeFromPackage(pkg Package, typeName, fieldName string) types.Type {
	for _, t := range pkg.Types {
		if t.Name == typeName {
			for _, f := range t.Fields {
				if f.Name == fieldName {
					return f.Type
				}
			}
		}
	}
	return nil
}

var _ = Describe("Loader", func() {
	var logger log15.Logger

	BeforeSuite(func() {
		logger = log15.New()
		logger.SetHandler(log15.DiscardHandler())
	})

	It("errors for unknown packages", func() {
		loader := New([]string{"github.com/jimmidyson/kube-client-gen/pkg/loader/testdata/unknown"}, logger)
		_, err := loader.Load()
		Expect(err).To(HaveOccurred())
	})

	It("parses single packages", func() {
		loader := New([]string{"github.com/jimmidyson/kube-client-gen/pkg/loader/testdata/pkg1"}, logger)
		pkgs, err := loader.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(pkgs).To(HaveLen(1))
		Expect(pkgs).To(Equal([]Package{
			{
				Path: "github.com/jimmidyson/kube-client-gen/pkg/loader/testdata/pkg1",
				Types: []Type{
					{
						Name:    "Type1",
						Package: "github.com/jimmidyson/kube-client-gen/pkg/loader/testdata/pkg1",
						Fields: []Field{
							{Name: "Field1", Doc: "Some doc.", Anonymous: false, JSONRequired: true, JSONProperty: "Field1", Type: types.Typ[types.Int], TypeName: "int"},
							{Name: "Field2", Doc: "", Anonymous: false, JSONRequired: true, JSONProperty: "f2", Type: types.Typ[types.String], TypeName: "string"},
							{Name: "Field4", Doc: "Even more doc.", Anonymous: false, JSONRequired: false, JSONProperty: "", Type: types.NewSlice(types.Typ[types.String]), TypeName: "[]string"},
							{Name: "Field5", Doc: "And some\nmore doc.", Anonymous: false, JSONRequired: false, JSONProperty: "f5", Type: types.NewMap(types.Typ[types.String], types.Typ[types.Bool]), TypeName: "map[string]bool"},
							{Name: "Type5", Doc: "", Anonymous: true, JSONRequired: false, JSONProperty: "", Type: typeFromPackage(pkgs[0], "Type1", "Type5"), TypeName: "github.com/jimmidyson/kube-client-gen/pkg/loader/testdata/pkg1.Type5"},
							{Name: "Type5s", Doc: "", JSONRequired: false, JSONProperty: "t5s", Type: typeFromPackage(pkgs[0], "Type1", "Type5s"), TypeName: "[]github.com/jimmidyson/kube-client-gen/pkg/loader/testdata/pkg1.Type5"},
						},
						Doc:            "Type1 is a normal type\nwith a single field and a description.",
						GenerateClient: true,
						Namespaced:     true,
					},
					{
						Name:    "Type5",
						Package: "github.com/jimmidyson/kube-client-gen/pkg/loader/testdata/pkg1",
						Fields: []Field{
							{Name: "Type5Field", Doc: "Something.", Anonymous: false, JSONRequired: true, JSONProperty: "t5", Type: types.Typ[types.Uint32], TypeName: "uint32"},
							{Name: "Type5Field2", Doc: "Something else.", Anonymous: false, JSONRequired: true, JSONProperty: "t6", Type: types.NewSlice(types.Typ[types.Uint32]), TypeName: "[]uint32"},
						},
						Doc:            "",
						GenerateClient: true,
						Namespaced:     false,
					},
				},
			},
		}))
	})
})
