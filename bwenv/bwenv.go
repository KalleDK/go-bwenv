package bwenv

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Folder struct {
	ID   string
	Name string
}

type Field struct {
	Name  string
	Value string
}

type Item struct {
	ID     string
	Name   string
	Fields []Field
}

type BitWarden struct {
	args []string
	path string
	env  []string
}

type Config struct {
	Path string
	Key  string
}

func (c Config) New() BitWarden {

	path := c.Path
	if path == "" {
		path = "bw"
	}

	if filepath.Base(path) == path {
		if lp, err := exec.LookPath(path); err != nil {
			log.Fatal(err)
		} else {
			path = lp
		}
	}

	return BitWarden{
		args: []string{c.Path},
		path: path,
		env:  append(os.Environ(), `BW_SESSION=`+c.Key),
	}
}

func (bw BitWarden) run(input io.Reader, args ...string) ([]byte, error) {
	cmd := exec.Cmd{
		Path:  bw.path,
		Args:  append(append(bw.args, args...), "--nointeraction"),
		Env:   bw.env,
		Stdin: input,
	}

	out, err := cmd.Output()
	if err != nil {
		return out, err
	}

	return out, nil
}

func (bw BitWarden) Sync() {
	bw.run(nil, "sync")
}

func (bw BitWarden) Encode() {
	r := strings.NewReader(`[{"object":"folder","id":"6e697460-8961-4d3d-b6e7-adb7009dd04c","name":"ServerProd"}]`)
	bw.run(r, "encode")
}

func (bw BitWarden) GetFolders(name string) (folder []Folder, err error) {
	out, err := bw.run(nil, "list", "folders", "--search", name)
	if err != nil {
		return
	}

	if err := json.Unmarshal(out, &folder); err != nil {
		return nil, err
	}

	return folder, nil
}

func (bw BitWarden) GetFolder(name string) (folder Folder, err error) {
	folders, err := bw.GetFolders(name)
	if err != nil {
		return folder, err
	}
	if len(folders) > 1 {
		return folder, errors.New("not a unique folder")
	}
	return folders[0], nil
}

func (bw BitWarden) GetItems(folder_id, name string) (items []Item, err error) {
	out, err := bw.run(nil, "list", "items", "--folderid", folder_id, "--search", name)
	if err != nil {
		return
	}

	if err := json.Unmarshal(out, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (bw BitWarden) GetItem(folder_id, name string) (item Item, err error) {
	items, err := bw.GetItems(folder_id, name)
	if err != nil {
		return item, err
	}
	if len(items) > 1 {
		return item, errors.New("not a unique item")
	}
	return items[0], nil
}

func (bw BitWarden) Exists(folder_id, name string) bool {
	if _, err := bw.GetItem(folder_id, name); err != nil {
		return false
	}

	return true
}

type EnvConfig struct {
	Config Config
	Folder string
}

type BitwardenEnv struct {
	cli      BitWarden
	folderID string
}

func (ec EnvConfig) New() BitwardenEnv {
	cli := ec.Config.New()
	folder, err := cli.GetFolder(ec.Folder)
	if err != nil {
		log.Fatal(err)
	}

	return BitwardenEnv{
		cli:      cli,
		folderID: folder.ID,
	}
}

func (bw BitwardenEnv) GetItem(name string) (item Item, err error) {
	return bw.cli.GetItem(bw.folderID, name)
}

func (bw BitwardenEnv) GetEnv(name string) (fields []Field, err error) {

	item, err := bw.cli.GetItem(bw.folderID, name)
	if err != nil {
		return nil, err
	}

	return item.Fields, nil
}
