package gosrc_test

import (
	"fmt"
	"go/build"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xaionaro-go/gosrc"
)

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func ExampleOpenDirectoryByPkgPath() {
	dir, err := gosrc.OpenDirectoryByPkgPath(&build.Default, "github.com/xaionaro-go/gosrc", false, false, false, nil)
	assertNoError(err)

	if len(dir.Packages) != 1 {
		panic("expected one package")
	}
	pkg := dir.Packages[0]

	for _, file := range pkg.Files {
		for _, s := range file.Structs() {
			fields, err := s.Fields()
			assertNoError(err)
			fmt.Println("amount of fields:", len(fields), ";\t amount of methods:", len(s.Methods()), ";  \tstruct name:", s.Name())
		}
	}
}

func TestOpenDirectoryByPkgPath(t *testing.T) {
	dir, err := gosrc.OpenDirectoryByPkgPath(&build.Default, "github.com/xaionaro-go/gosrc", false, false, false, nil)
	require.NoError(t, err)
	require.Len(t, dir.Packages, 1)

	dir, err = gosrc.OpenDirectoryByPkgPath(&build.Default, ".", false, false, false, nil)
	require.NoError(t, err)
	require.Len(t, dir.Packages, 1)

	dir, err = gosrc.OpenDirectoryByPkgPath(&build.Default, ".", true, true, false, nil)
	require.NoError(t, err)
	require.Len(t, dir.Packages, 2)

	dir, err = gosrc.OpenDirectoryByPkgPath(&build.Default, ".", true, false, false, nil)
	require.NoError(t, err)
	require.Len(t, dir.Packages, 1)

	dir, err = gosrc.OpenDirectoryByPkgPath(&build.Default, ".", false, true, false, nil)
	require.NoError(t, err)
	require.Len(t, dir.Packages, 1)

	dir, err = gosrc.OpenDirectoryByPkgPath(&build.Default, ".", true, true, true, nil)
	require.NoError(t, err)
	require.Len(t, dir.Packages, 2)
}
