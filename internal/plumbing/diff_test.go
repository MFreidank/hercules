package plumbing_test

import (
	"testing"
	"unicode/utf8"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/hercules.v4/internal"
	"gopkg.in/src-d/hercules.v4/internal/core"
	items "gopkg.in/src-d/hercules.v4/internal/plumbing"
	"gopkg.in/src-d/hercules.v4/internal/test"
	"gopkg.in/src-d/hercules.v4/internal/test/fixtures"
)

func TestFileDiffMeta(t *testing.T) {
	fd := fixtures.FileDiff()
	assert.Equal(t, fd.Name(), "FileDiff")
	assert.Equal(t, len(fd.Provides()), 1)
	assert.Equal(t, fd.Provides()[0], items.DependencyFileDiff)
	assert.Equal(t, len(fd.Requires()), 2)
	assert.Equal(t, fd.Requires()[0], items.DependencyTreeChanges)
	assert.Equal(t, fd.Requires()[1], items.DependencyBlobCache)
	assert.Len(t, fd.ListConfigurationOptions(), 1)
	assert.Equal(t, fd.ListConfigurationOptions()[0].Name, items.ConfigFileDiffDisableCleanup)
	facts := map[string]interface{}{}
	facts[items.ConfigFileDiffDisableCleanup] = true
	fd.Configure(facts)
	assert.True(t, fd.CleanupDisabled)
}

func TestFileDiffRegistration(t *testing.T) {
	summoned := core.Registry.Summon((&items.FileDiff{}).Name())
	assert.Len(t, summoned, 1)
	assert.Equal(t, summoned[0].Name(), "FileDiff")
	summoned = core.Registry.Summon((&items.FileDiff{}).Provides()[0])
	assert.True(t, len(summoned) >= 1)
	matched := false
	for _, tp := range summoned {
		matched = matched || tp.Name() == "FileDiff"
	}
	assert.True(t, matched)
}

func TestFileDiffConsume(t *testing.T) {
	fd := fixtures.FileDiff()
	deps := map[string]interface{}{}
	cache := map[plumbing.Hash]*object.Blob{}
	hash := plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe")
	cache[hash], _ = test.Repository.BlobObject(hash)
	hash = plumbing.NewHash("334cde09da4afcb74f8d2b3e6fd6cce61228b485")
	cache[hash], _ = test.Repository.BlobObject(hash)
	hash = plumbing.NewHash("dc248ba2b22048cc730c571a748e8ffcf7085ab9")
	cache[hash], _ = test.Repository.BlobObject(hash)
	deps[items.DependencyBlobCache] = cache
	changes := make(object.Changes, 3)
	treeFrom, _ := test.Repository.TreeObject(plumbing.NewHash(
		"a1eb2ea76eb7f9bfbde9b243861474421000eb96"))
	treeTo, _ := test.Repository.TreeObject(plumbing.NewHash(
		"994eac1cd07235bb9815e547a75c84265dea00f5"))
	changes[0] = &object.Change{From: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeFrom,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("dc248ba2b22048cc730c571a748e8ffcf7085ab9"),
		},
	}, To: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeTo,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("334cde09da4afcb74f8d2b3e6fd6cce61228b485"),
		},
	}}
	changes[1] = &object.Change{From: object.ChangeEntry{}, To: object.ChangeEntry{
		Name: ".travis.yml",
		Tree: treeTo,
		TreeEntry: object.TreeEntry{
			Name: ".travis.yml",
			Mode: 0100644,
			Hash: plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"),
		},
	},
	}
	changes[2] = &object.Change{From: object.ChangeEntry{
		Name: "rbtree.go",
		Tree: treeFrom,
		TreeEntry: object.TreeEntry{
			Name: "rbtree.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("14c3fa5a1cca103032f10379467a3a2f210e5f94"),
		},
	}, To: object.ChangeEntry{},
	}
	deps[items.DependencyTreeChanges] = changes
	res, err := fd.Consume(deps)
	assert.Nil(t, err)
	diffs := res[items.DependencyFileDiff].(map[string]items.FileDiffData)
	assert.Equal(t, len(diffs), 1)
	diff := diffs["analyser.go"]
	assert.Equal(t, diff.OldLinesOfCode, 307)
	assert.Equal(t, diff.NewLinesOfCode, 309)
	deletions := 0
	insertions := 0
	for _, edit := range diff.Diffs {
		switch edit.Type {
		case diffmatchpatch.DiffEqual:
			continue
		case diffmatchpatch.DiffInsert:
			insertions += utf8.RuneCountInString(edit.Text)
		case diffmatchpatch.DiffDelete:
			deletions += utf8.RuneCountInString(edit.Text)
		}
	}
	assert.Equal(t, deletions, 13)
	assert.Equal(t, insertions, 15)
}

func TestFileDiffConsumeInvalidBlob(t *testing.T) {
	fd := fixtures.FileDiff()
	deps := map[string]interface{}{}
	cache := map[plumbing.Hash]*object.Blob{}
	hash := plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe")
	cache[hash], _ = test.Repository.BlobObject(hash)
	hash = plumbing.NewHash("334cde09da4afcb74f8d2b3e6fd6cce61228b485")
	cache[hash], _ = test.Repository.BlobObject(hash)
	hash = plumbing.NewHash("dc248ba2b22048cc730c571a748e8ffcf7085ab9")
	cache[hash], _ = test.Repository.BlobObject(hash)
	deps[items.DependencyBlobCache] = cache
	changes := make(object.Changes, 1)
	treeFrom, _ := test.Repository.TreeObject(plumbing.NewHash(
		"a1eb2ea76eb7f9bfbde9b243861474421000eb96"))
	treeTo, _ := test.Repository.TreeObject(plumbing.NewHash(
		"994eac1cd07235bb9815e547a75c84265dea00f5"))
	changes[0] = &object.Change{From: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeFrom,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("ffffffffffffffffffffffffffffffffffffffff"),
		},
	}, To: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeTo,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("334cde09da4afcb74f8d2b3e6fd6cce61228b485"),
		},
	}}
	deps[items.DependencyTreeChanges] = changes
	res, err := fd.Consume(deps)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	changes[0] = &object.Change{From: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeFrom,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("dc248ba2b22048cc730c571a748e8ffcf7085ab9"),
		},
	}, To: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeTo,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("ffffffffffffffffffffffffffffffffffffffff"),
		},
	}}
	res, err = fd.Consume(deps)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestCountLines(t *testing.T) {
	blob, _ := test.Repository.BlobObject(
		plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"))
	lines, err := items.CountLines(blob)
	assert.Equal(t, lines, 12)
	assert.Nil(t, err)
	lines, err = items.CountLines(nil)
	assert.Equal(t, lines, -1)
	assert.NotNil(t, err)
	blob, _ = internal.CreateDummyBlob(plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"), true)
	lines, err = items.CountLines(blob)
	assert.Equal(t, lines, -1)
	assert.NotNil(t, err)
	// test_data/blob
	blob, err = test.Repository.BlobObject(
		plumbing.NewHash("c86626638e0bc8cf47ca49bb1525b40e9737ee64"))
	assert.Nil(t, err)
	lines, err = items.CountLines(blob)
	assert.Equal(t, lines, -1)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "binary")
}

func TestBlobToString(t *testing.T) {
	blob, _ := test.Repository.BlobObject(
		plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"))
	str, err := items.BlobToString(blob)
	assert.Nil(t, err)
	assert.Equal(t, str, `language: go

go:
  - 1.7

go_import_path: gopkg.in/src-d/hercules.v1
`+"  "+`
script:
  - go test -v -cpu=1,2 ./...

notifications:
  email: false
`)
	str, err = items.BlobToString(nil)
	assert.Equal(t, str, "")
	assert.NotNil(t, err)
	blob, _ = internal.CreateDummyBlob(plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"), true)
	str, err = items.BlobToString(blob)
	assert.Equal(t, str, "")
	assert.NotNil(t, err)
}

func TestFileDiffDarkMagic(t *testing.T) {
	fd := fixtures.FileDiff()
	deps := map[string]interface{}{}
	cache := map[plumbing.Hash]*object.Blob{}
	hash := plumbing.NewHash("448eb3f312849b0ca766063d06b09481c987b309")
	cache[hash], _ = test.Repository.BlobObject(hash) // 1.java
	hash = plumbing.NewHash("3312c92f3e8bdfbbdb30bccb6acd1b85bc338dfc")
	cache[hash], _ = test.Repository.BlobObject(hash) // 2.java
	deps[items.DependencyBlobCache] = cache
	changes := make(object.Changes, 1)
	treeFrom, _ := test.Repository.TreeObject(plumbing.NewHash(
		"f02289bfe843388a1bb3c7dea210374082dd86b9"))
	treeTo, _ := test.Repository.TreeObject(plumbing.NewHash(
		"eca91acf1fd828f20dcb653a061d8c97d965bc6c"))
	changes[0] = &object.Change{From: object.ChangeEntry{
		Name: "test.java",
		Tree: treeFrom,
		TreeEntry: object.TreeEntry{
			Name: "test.java",
			Mode: 0100644,
			Hash: plumbing.NewHash("448eb3f312849b0ca766063d06b09481c987b309"),
		},
	}, To: object.ChangeEntry{
		Name: "test.java",
		Tree: treeTo,
		TreeEntry: object.TreeEntry{
			Name: "test.java",
			Mode: 0100644,
			Hash: plumbing.NewHash("3312c92f3e8bdfbbdb30bccb6acd1b85bc338dfc"),
		},
	}}
	deps[items.DependencyTreeChanges] = changes
	res, err := fd.Consume(deps)
	assert.Nil(t, err)
	magicDiffs := res[items.DependencyFileDiff].(map[string]items.FileDiffData)["test.java"]
	fd.CleanupDisabled = true
	res, err = fd.Consume(deps)
	assert.Nil(t, err)
	plainDiffs := res[items.DependencyFileDiff].(map[string]items.FileDiffData)["test.java"]
	assert.NotEqual(t, magicDiffs.Diffs, plainDiffs.Diffs)
	assert.Equal(t, magicDiffs.OldLinesOfCode, plainDiffs.OldLinesOfCode)
	assert.Equal(t, magicDiffs.NewLinesOfCode, plainDiffs.NewLinesOfCode)
}

func TestFileDiffFork(t *testing.T) {
	fd1 := fixtures.FileDiff()
	clones := fd1.Fork(1)
	assert.Len(t, clones, 1)
	fd2 := clones[0].(*items.FileDiff)
	assert.True(t, fd1 == fd2)
	fd1.Merge([]core.PipelineItem{fd2})
}