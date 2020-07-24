package fakesysfs

import (
	"io/ioutil"
	"os"
)

type Tree interface {
	Add(name string, attrs map[string]string) Tree
	Name() string
	Items() []Tree
	SetAttrs() error
	Create() error
}

type Creator interface {
	Create(Tree) error
}

func NewTree(name string, attrs map[string]string) Tree {
	return &tree{
		name:  name,
		attrs: attrs,
		items: []Tree{},
	}
}

type tree struct {
	name  string
	attrs map[string]string
	items []Tree
}

func (t *tree) Add(name string, attrs map[string]string) Tree {
	n := NewTree(name, attrs)
	t.items = append(t.items, n)
	return n
}

func (t *tree) Items() []Tree {
	return t.items
}

func (t *tree) Name() string {
	return t.name
}

func (t *tree) Create() error {
	return newCreator().Create(t)
}

func (t *tree) SetAttrs() error {
	if t.attrs == nil {
		return nil
	}
	var err error
	for name, content := range t.attrs {
		err = ioutil.WriteFile(name, []byte(content), 0644)
		if err != nil {
			break
		}
	}
	return err
}

type creator struct{}

func newCreator() Creator {
	return &creator{}
}

func (c *creator) Create(t Tree) error {
	if err := t.SetAttrs(); err != nil {
		return err
	}
	return c.createItems(t.Items())
}

func (c *creator) createItems(t []Tree) error {
	var err error
	for _, st := range t {
		err = os.Mkdir(st.Name(), 0755)
		if err != nil {
			break
		}
		err = c.createItem(st)
	}
	return err
}

func (c *creator) createItem(st Tree) error {
	os.Chdir(st.Name())
	defer os.Chdir("..")
	return st.Create()
}

type FakeSysfs struct {
	base string
	root Tree
}

func NewFakeSysfs(base string) (*FakeSysfs, error) {
	return &FakeSysfs{
		base: base,
		// DO NOT USE NEITHER "." or "" HERE!!
		root: NewTree("_", nil),
	}, nil
}

func (fs *FakeSysfs) AddTree(entries ...string) Tree {
	pos := fs.root
	for _, entry := range entries {
		pos = pos.Add(entry, nil)
	}
	return pos
}

func (fs *FakeSysfs) Base() string {
	return fs.base
}

func (fs *FakeSysfs) Root() Tree {
	return fs.root
}

func (fs *FakeSysfs) Setup() error {
	oldWd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(oldWd)
	os.Chdir(fs.base)
	return fs.root.Create()
}

func (fs *FakeSysfs) Teardown() error {
	return os.RemoveAll(fs.base)
}
