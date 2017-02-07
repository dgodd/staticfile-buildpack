package main_test

import (
	"os"
	"path/filepath"

	main "."
	bp "github.com/cloudfoundry/libbuildpack"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type MockManifest struct {
	DefaultVersionFunc  func(depName string) (bp.Dependency, error)
	FetchDependencyFunc func(dep bp.Dependency, outputFile string) error
}

func (mock *MockManifest) DefaultVersion(depName string) (bp.Dependency, error) {
	return mock.DefaultVersionFunc(depName)
}
func (mock *MockManifest) FetchDependency(dep bp.Dependency, outputFile string) error {
	return mock.FetchDependencyFunc(dep, outputFile)
}

var _ = Describe("Compile", func() {
	var buildDir, cacheDir string
	BeforeEach(func() {
		var err error
		buildDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
		cacheDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		os.RemoveAll(buildDir)
		os.RemoveAll(cacheDir)
	})

	Context("Empty Staticfile", func() {
		var mockManifest *MockManifest
		BeforeEach(func() {
			err := ioutil.WriteFile(filepath.Join(buildDir, "Staticfile"), []byte(""), 0666)
			Expect(err).ToNot(HaveOccurred())

			mockManifest = &MockManifest{
				DefaultVersionFunc: func(depName string) (bp.Dependency, error) {
					return bp.Dependency{}, nil
				},
				FetchDependencyFunc: func(dep bp.Dependency, outputFile string) error {
					return nil
				},
			}
		})

		It("places templated nginx conf in <build_dir>/nginx/conf/nginx.conf", func() {
			err := main.Compile(buildDir, cacheDir, mockManifest, ".")
			Expect(err).ToNot(HaveOccurred())

			file, err := ioutil.ReadFile(filepath.Join(buildDir, "nginx", "conf", "nginx.conf"))
			Expect(err).ToNot(HaveOccurred())

			Expect(string(file)).To(ContainSubstring("root ../public"))
		})

		It("copies mime.type to <build_dir>/nginx/conf/mime.types", func() {
			err := main.Compile(buildDir, cacheDir, mockManifest, ".")
			Expect(err).ToNot(HaveOccurred())

			file, err := ioutil.ReadFile(filepath.Join(buildDir, "nginx", "conf", "mime.types"))
			Expect(err).ToNot(HaveOccurred())

			Expect(string(file)).To(ContainSubstring("text/html html htm shtml"))
		})
	})
})
