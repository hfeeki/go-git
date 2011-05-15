package git

import (
	"exec"
	"fmt"
	"os"
	"strings"
	"testing"
)

var (
	repo    *Repo
	revwalk *RevWalk
	path    string
	author  string
	ref     = &Reference{}
)

func init() {
	path = "./tmp"
	author = "foot@bar.com"
}

// Repo
func TestInitBare(t *testing.T) {
	repo = new(Repo)
	if err := repo.Init(path, BARE); err != nil {
		t.Fatal("Error:", err)
	}
}

func TestOpenBare(t *testing.T) {
	defer os.RemoveAll(path)
	defer repo.Free()
	err := repo.Open(path)
	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestInitNotBare(t *testing.T) {
	repo = new(Repo)
	if err := repo.Init(path, NOTBARE); err != nil {
		t.Fatal("Error:", err)
	}
}

func TestOpenNotBare(t *testing.T) {
	err := repo.Open(path + "/.git")
	if err != nil {
		t.Fatal("Error:", err)
	}
}

//FIXME: figure out how to seed an intial HEAD
func TestSeed(t *testing.T) {
	var (
		cmd *exec.Cmd
	)

	tmpfile := "README"

	f, err := os.Create(path + "/" + tmpfile)
	_, err = f.WriteString("foo\n")
	f.Close()
	if err != nil {
		println(err.String())
		os.Exit(1)
	}
	cmd, err = run("git add " + tmpfile)
	cmd.Close()
	cmd, err = run("git commit -m test")
	cmd.Close()

	if err != nil {
		t.Fatal("Error:", err)
	}
}

// Index
func TestIndexAdd(t *testing.T) {
	index := new(Index)
	defer index.Free()
	err := index.Open(repo)
	check(t, err)
	tmpfile := "README"
	f, err := os.OpenFile(path+"/"+tmpfile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	_, err = f.WriteString("foo\n")
	f.Close()
	err = index.Add(tmpfile)
	check(t, err)
	err = index.Write()
	check(t, err)
}

// Commit
func TestCommit(t *testing.T) {
	TestIndexAdd(t)
	index := new(Index)
	defer index.Free()
	err := index.Open(repo)
	check(t, err)
	tree, err := TreeFromIndex(repo, index)
	check(t, err)
	head, _ := GetHeadString(repo)
	parent, err := NewOidString(head)
	check(t, err)
	s := NewSignature("Foo Bar", "foo@bar.com")
	err = CommitCreate(repo, tree, parent, s, s, "some stuff here")
	check(t, err)
}

func TestManyCommits(t *testing.T) {
	for i := 0; i < 29; i++ {
		TestCommit(t)
	}
}

// RevWalk
func TestNewRevWalk(t *testing.T) {
	var err os.Error
	revwalk, err = NewRevWalk(repo)
	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestRevWalkNext(t *testing.T) {
	head, _ := GetHeadString(repo)
	o, _ := NewOidString(head)
	revwalk.Push(o)
	if err := revwalk.Next(o); err != nil {
		t.Fatal("Error:", err)
	}
	if o.String() != head {
		t.Errorf("oid string should match %v is %v", head, o.String())
	}
}

func TestRevWalk(t *testing.T) {
	r := new(Repo)
	err := r.Open("./tmp/.git")
	check(t, err)
	o := NewOid()
	for {
		if err := revwalk.Next(o); err != nil {
			break
		}
		c := new(Commit)
		c.Lookup(repo, o)
		fmt.Printf("%v %v %v %v\n", o.String(), c.Author(), c.Email(), c.Msg())
	}
}

// Oid
func TestNewOid(t *testing.T) {
	head, err := GetHeadString(repo)
	check(t, err)
	if _, err := NewOidString(head); err != nil {
		t.Error(err)
	}
}

// Singature
func TestSignature(t *testing.T) {
	NewSignature("foo", "bar")
}

// Tree
func TestTreeFromIndex(t *testing.T) {
	index := new(Index)
	defer index.Free()
	err := index.Open(repo)
	check(t, err)
	_, err = TreeFromIndex(repo, index)
	check(t, err)
}

// Important: this must be called after all of the Test functions
func TestFinal(t *testing.T) {
	if revwalk != nil {
		revwalk.Free()
	}
	if repo != nil {
		repo.Free()
	}
}

// private helper functions
func run(s string) (cmd *exec.Cmd, err os.Error) {
	wd := "./tmp/"
	args := strings.Split(s, " ", -1)
	bin, err := exec.LookPath(args[0])

	cmd, err = exec.Run(bin, args, os.Environ(), wd, exec.DevNull, exec.Pipe, exec.PassThrough)

	return
}

func check(t *testing.T, err os.Error) {
	if err != nil {
		t.Error(err)
	}
}
