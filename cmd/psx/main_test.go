package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestSplitMemoryCards(t *testing.T) {
	tables := map[string]struct {
		input  []string
		output map[string][]string
	}{
		"empty": {
			input: []string{
				filepath.Join("..", "..", "testdata", "blank.mcd"),
			},
		},
		"good": {
			input: []string{
				filepath.Join("..", "..", "testdata", "MemoryCard2-1.mcd"),
				filepath.Join("..", "..", "testdata", "MemoryCard2-2.mcd"),
				filepath.Join("..", "..", "testdata", "MemoryCard2-3.mcd"),
				filepath.Join("..", "..", "testdata", "MemoryCard2-4.mcd"),
			},
			output: map[string][]string{
				"d57756b1739262daa57ffe4885baa27d76966edbaefa458d7abeb184b5fa62c5": []string{"SCES-00582", "SCES-00582-1.mcd"},
				"f99bbc63ff9d537368508ebcc1d4bec380a2271fadfa2290be66fef76dbf67ad": []string{"SCES-00967", "SCES-00967-1.mcd"},
				"567a7331f402648e36036e881237a56db9b43e1d5ebe975ded40cf3e9e7a2790": []string{"SCES-00984", "SCES-00984-1.mcd"},
				"0c932ca72ba02ff8d4e6898d3d44cdf8140db6b5585c18e61726ef9372d5160f": []string{"SCES-01237", "SCES-01237-1.mcd"},
				"d375da2bfdc98271a7640b1bac1c8b3264c213d9dc4663dbf280fdc005adf1b0": []string{"SCES-02380", "SCES-02380-1.mcd"},
				"dfe854e73025ee6541c83bc5424ab5d2f1e1b1104092df8c3230eea21294d5cd": []string{"SCES-02380", "SCES-02380-2.mcd"},
				"65dc9d74ee5978b8e8f2cf39354b72a8dd4c84d347d7d55e9a96b70c2ac597e4": []string{"SLES-00016", "SLES-00016-1.mcd"},
				"c84272396cd775c0ebccda7fdfafa35829832d38ade8ef12ba5614f66be99bb3": []string{"SLES-00024", "SLES-00024-1.mcd"},
				"897753f84d0667bee7a216e71e6758d669dc661d65ab1270601c3c95dde3fde9": []string{"SLES-00327", "SLES-00327-1.mcd"},
				"67d5f67fa301c3463011da3f4cb899ac4033ec3bfd34137dbd7c7329ad066258": []string{"SLES-00477", "SLES-00477-1.mcd"},
				"2e9e7bc3b05d3012afc51b5ec8b3525aadda3a38af802af8959bee44f4edcbc2": []string{"SLES-00524", "SLES-00524-1.mcd"},
				"aae51b3bc67ca355dd4b9b88e6a47c3036fb70a98da8636d9cff7c249be0102b": []string{"SLES-01051", "SLES-01051-1.mcd"},
				"b866a8d1fb2b941854c7e379049381001ea4315ffdf18732697d7d84cd6041bc": []string{"SLES-01370", "SLES-01370-1.mcd"},
				"e669b80b77dcb1f5d280622d5f4174a8afd69755d6c1bbe9679c8b6470bb617f": []string{"SLES-01374", "SLES-01374-1.mcd"},
				"41cb854d1cedc6bcca34656a7b4550ab01fe19dfd146e953aaeeb199a8950af5": []string{"SLES-01893", "SLES-01893-1.mcd"},
				"41711bb06ecc10f7802e633b7fa019f4f415cf88e0ad89e171c9baf1724d4884": []string{"SLES-02055", "SLES-02055-1.mcd"},
				"30ca8e451ca43c00897984be251a6392989343abd76779becfd1e42138d58b89": []string{"SLES-02158", "SLES-02158-1.mcd"},
				"aca51d85691a64fac2312323f02c5b3dca15503b4fbd22b054b8b3ce9893ef40": []string{"SLES-02886", "SLES-02886-1.mcd"},
				"95a9a3802e74c930f48a8edcd2e5d552c09fd9bb9383ff0963f2129fdede09bd": []string{"SLES-02906", "SLES-02906-1.mcd"},
				"698ac7fda15e0292bdbf5a9fa29cf2322e9ccae9ec663104961dbf8ae44882d4": []string{"SLES-02908", "SLES-02908-1.mcd"},
				"b4e8eee61c6aa6a0e750f2f93f3662e9bcde48f7133d6c124a1c25674fa25ae4": []string{"SLUS-00859", "SLUS-00859-1.mcd"},
			},
		},
		"issue9": {
			input: []string{
				filepath.Join("..", "..", "testdata", "9", "MemoryCard2-1.mcd"),
			},
			output: map[string][]string{
				"373421c84c68f18ab5751ec41ef805db223e85f13d3bb6caee3eab55cdc2d2b7": []string{"SCUS-94163", "SCUS-94163-1.mcd"},
				"363b5c49aace090f121af467a468582191defb32b79ba938358f4af93b609913": []string{"SCUS-94244", "SCUS-94244-1.mcd"},
				"fe6bd1c6551f3053d3054ccd226e1274920b8a762d462019c69bd84fd380b5f9": []string{"SCUS-94426", "SCUS-94426-1.mcd"},
				"e436758c040da5bcda3258c9b9d9c444fb58af613b31a65957ccf2b609b0c074": []string{"SCUS-94900", "SCUS-94900-1.mcd"},
				"458f99a5560958bf8742e3a8f9c2b6ec40943038d2ce47b27fcd82594b2d0b17": []string{"SLUS-00032", "SLUS-00032-1.mcd"},
				"a014ec20c65ff3a989217fc0659e4110202296c1f29c4192e7178d390d25ed79": []string{"SLUS-00398", "SLUS-00398-1.mcd"},
				"4a60baed3b9628ea795c4e5349f3d559054791fe2d912048d4d77d6080dc2258": []string{"SLUS-00439", "SLUS-00439-1.mcd"},
				"d30cd43e7eb22b22ff27108398a4deff68d8e1bcf26a161ca21836f65cd0792b": []string{"SLUS-00620", "SLUS-00620-1.mcd"},
				"9340558615f5f01850b2783425a3b48399dd7ddb8802ac01d9d3d6173087b102": []string{"SLUS-00839", "SLUS-00839-1.mcd"},
				"2f7476eb606dc2d359af8cb5d61f3f19ac067e1f4f3a5326734e645f6fcbf7c9": []string{"SLUS-00840", "SLUS-00840-1.mcd"},
				"e26dbe44a416b86b91ef2dca4b0a2b23f41b182b95cd88b27f14409697187434": []string{"SLUS-00892", "SLUS-00892-1.mcd"},
				"b3280891276617f68ae41fda349fd7b76aff843f76e06e38967af65c5b790161": []string{"SLUS-01251", "SLUS-01251-1.mcd"},
				"f8f4fbbb4afda799257e8e05d6a3304dc5a03ef5d733bc4e9e2c2c8726821c9e": []string{"SLUS-01541", "SLUS-01541-1.mcd"},
			},
		},
	}

	for name, table := range tables {
		t.Run(name, func(t *testing.T) {
			oldFs := fs
			defer func() { fs = oldFs }()
			fs = afero.NewCopyOnWriteFs(afero.NewReadOnlyFs(afero.NewOsFs()), afero.NewMemMapFs())

			dir, err := afero.TempDir(fs, "", "psx")
			if err != nil {
				t.Fatal(err)
			}

			if err := splitMemoryCards(dir, table.input); err != nil {
				t.Fatal(err)
			}

			files, dirs := make(map[string]struct{}), make(map[string]struct{})

			h := sha256.New()

			for checksum, path := range table.output {
				file := filepath.Join(append([]string{dir}, path...)...)
				files[file], dirs[filepath.Dir(file)] = struct{}{}, struct{}{}

				f, err := fs.Open(file)
				if err != nil {
					t.Fatal(err)
				}

				h.Reset()
				if _, err := io.Copy(h, f); err != nil {
					t.Fatal(err)
				}

				f.Close()

				assert.Equal(t, checksum, fmt.Sprintf("%0*x", h.Size()<<1, h.Sum(nil)))
			}

			if err := afero.Walk(fs, dir, func(path string, info os.FileInfo, err error) error {
				if path == dir {
					return nil
				}

				switch {
				case info.Mode().IsDir():
					if _, ok := dirs[path]; !ok {
						t.Errorf("directory %s should not exist", path)
					}
				case info.Mode().IsRegular():
					if _, ok := files[path]; !ok {
						t.Errorf("regular file %s should not exist", path)
					}
				default:
					t.Errorf("file %s should not exist", path)
				}

				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
