// +build linux darwin

package snapshot_test

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sclevine/spec"

	"github.com/buildpacks/lifecycle/snapshot"
	h "github.com/buildpacks/lifecycle/testhelpers"
)

func TestKanikoSnapshotter(t *testing.T) {
	spec.Run(t, "Test Image", testKanikoSnapshotter)
}

func testKanikoSnapshotter(t *testing.T, when spec.G, it spec.S) {
	var (
		snapshotter *snapshot.KanikoSnapshotter
		tmpDir      string
	)

	it.Before(func() {
		// Using the default tmp dir causes kaniko to go haywire for some reason
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Error: %s\n", err)
		}
		tmpDir, err = ioutil.TempDir(filepath.Join(cwd, "..", "tmp"), "kaniko")
		if err != nil {
			t.Fatalf("Error: %s\n", err)
		}

		createTestFileWithContent(t, filepath.Join(tmpDir, "cnb", "privatefile"), "hello\n")
		createTestFileWithContent(t, filepath.Join(tmpDir, "layers", "privatefile"), "hello\n")
		createTestFileWithContent(t, filepath.Join(tmpDir, "file-to-change"), "hello\n")
		createTestFileWithContent(t, filepath.Join(tmpDir, "file-not-to-change"), "hello\n")
		createTestFileWithContent(t, filepath.Join(tmpDir, "file-to-delete"), "hello\n")
		createTestFileWithContent(t, filepath.Join(tmpDir, "bin", "file-not-to-change"), "hello\n")

		snapshotter = &snapshot.KanikoSnapshotter{
			RootDir:                    tmpDir,
			DetectFilesystemIgnoreList: false,
		}
		snapshotter.IgnoredPaths = []string{
			filepath.Join(snapshotter.RootDir, "platform"),
			filepath.Join(snapshotter.RootDir, "layers"),
			filepath.Join(snapshotter.RootDir, "cnb"),
		}
	})

	it.After(func() {
		os.RemoveAll(tmpDir)
	})

	when("files are added and modified", func() {
		var (
			snapshotFile     string
			expectedNumFiles int
		)

		it.Before(func() {
			h.AssertNil(t, snapshotter.Init())

			// this file should exist in a deleted state in the snapshot
			os.Remove(filepath.Join(snapshotter.RootDir, "file-to-delete"))

			// these files should exist in the snapshot
			createTestFileWithContent(t, filepath.Join(snapshotter.RootDir, "newfile"), "hello\n")
			createTestFileWithContent(t, filepath.Join(snapshotter.RootDir, "my-space", "newfile-in-dir"), "hello\n")

			// these files should have updated content in the snapshot
			createTestFileWithContent(t, filepath.Join(snapshotter.RootDir, "file-to-change"), "hola\n")

			// these files should not exist in the snapshot
			createTestFileWithContent(t, filepath.Join(snapshotter.RootDir, "layers", "file-to-ignore"), "hello\n")
			createTestFileWithContent(t, filepath.Join(snapshotter.RootDir, "cnb", "file-to-ignore"), "hello\n")
			createTestFileWithContent(t, filepath.Join(snapshotter.RootDir, "platform", "file-to-ignore"), "hello\n")

			expectedNumFiles = 4

			tmpFile, err := ioutil.TempFile("", "snapshot")
			if err != nil {
				t.Fatalf("Error: %s\n", err)
			}

			snapshotFile = tmpFile.Name()

			err = snapshotter.TakeSnapshot(snapshotFile)
			if err != nil {
				t.Fatalf("Error: %s\n", err)
			}
		})

		it("includes the expected files in the snapshot", func() {
			data, err := os.Open(snapshotFile)
			if err != nil {
				t.Fatalf("Error: %s\n", err)
			}
			defer data.Close()

			tr := tar.NewReader(data)
			actualNumFiles := 0
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break // End of archive
				}

				if err != nil {
					t.Fatalf("Error: %s\n", err)
				}

				if hdr.FileInfo().IsDir() {
					continue
				}

				actualNumFiles++
				switch hdr.Name {
				case strings.Trim(filepath.Join(snapshotter.RootDir, ".wh.file-to-delete"), "/"):
					// ensure we only count the whiteout file
					continue
				case "newfile":
				case "my-space/newfile-in-dir":
					// ensure the contents is as expected
					assertSnapshotFile(t, tr, "hello\n")
				case "file-to-change":
					// ensure the contents is updated
					assertSnapshotFile(t, tr, "hola\n")
				default:
					t.Fatalf("Unexpected file in snapshot: %s\n", hdr.Name)
				}
			}
			// make sure we have the right number of files in our snapshot
			h.AssertEq(t, actualNumFiles, expectedNumFiles)
		})
	})
}

func mkdir(t *testing.T, dirs ...string) {
	t.Helper()
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0777); err != nil {
			t.Fatalf("Error: %s\n", err)
		}
	}
}

func createTestFileWithContent(t *testing.T, path string, content string) {
	mkdir(t, filepath.Dir(path))

	data := []byte(content)

	if err := ioutil.WriteFile(path, data, 0777); err != nil {
		t.Fatalf("Error: %s", err)
	}
}

func assertSnapshotFile(t *testing.T, tr *tar.Reader, content string) {
	var b bytes.Buffer
	if _, err := io.Copy(&b, tr); err != nil {
		t.Fatalf("Unexpected info:\n%s\n", err)
	}

	if s := cmp.Diff(b.String(), content); s != "" {
		t.Fatalf("Unexpected info:\n%s\n", s)
	}
}
