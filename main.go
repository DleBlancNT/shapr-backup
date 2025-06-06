package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"

	_ "modernc.org/sqlite"
)

type Globals struct {
	ProjectRoot string   `env:"PROJECT_ROOT" type:"" default:"$LOCALAPPDATA\\Packages\\Shapr3D.Shapr3D_dvv5p1vgwv6mp"`
	Target      string   `env:"EXPORT_DIR" short:"d" help:"Export directory." default:"."`
	AddRevision bool     `short:"r" help:"Add revision ID to filename."`
	AddDirs     bool     `short:"s" help:"Make folders for export."`
	Apps        []string `arg:"1" optional:"1" help:"Drawings..."`
}

var global Globals

type ShaprMetadata struct {
	RemoteID         string `json:"remoteID"`
	RevisionID       int    `json:"revisionID"`
	LocalChangeCount int    `json:"localChangeCount"`
}

func init() {
	godotenv.Load()
	kong.Parse(&global)
}

func JSON(v any) string {
	b, _ := json.MarshalIndent(v, ``, `  `)
	return string(b)
}

func sanitize(name string) string {
	var out string

	for _, c := range name {
		switch c {
		case '/', '\\', ':':
			c = '_'
		}
		out += string(c)
	}

	return out
}

func mkzipname(project, folder string, rev int, index int) string {
	var name string
	folder = sanitize(folder)
	project = sanitize(project)
	if global.AddDirs && folder != `` {
		folder += string(os.PathSeparator)
	} else if folder != `` {
		folder += `_`
	}
	name = fmt.Sprintf("%s%s", folder, project)
	if global.AddRevision && rev > 0 {
		name += fmt.Sprintf(" [rev-%d]", rev)
	}
	if index > 0 {
		name += fmt.Sprintf(" (%d)", index)
	}
	return filepath.Join(global.Target, name+`.shapr`)
}

func main() {

	var projects []os.DirEntry
	var err error

	dir := os.ExpandEnv(global.ProjectRoot)

	datadb := filepath.Join(dir, `LocalState`, `storage`, `projectStorage.db`)

	var sqldb *sql.DB

	if sqldb, err = sql.Open("sqlite", datadb); err != nil {
		panic(err)
	}

	project_dir := filepath.Join(dir, `LocalState`, `projects`)

	if projects, err = os.ReadDir(project_dir); err != nil {
		panic(err)
	}

	for e := range projects {
		var rows *sql.Rows

		name := projects[e].Name()
		workspace := filepath.Join(project_dir, name, `project`, `workspace`)

		if _, err = os.Stat(workspace); err != nil {
			continue
		}

		if rows, err = sqldb.Query(`select ifnull(title,projectid),ifnull(folderpath,""),ifnull(revisionid,0) from projects where projectid=?`, name); err != nil {
			panic(err)
		}

		for rows.Next() {
			var title, folder, zipname string
			var revisionid int

			if err = rows.Scan(&title, &folder, &revisionid); err != nil {
				fmt.Printf("Failed for projectID = %s\n", name)
				panic(err)
			}

			index := 0
			zipname = mkzipname(title, folder, revisionid, index)

			if title != `` {

				match := len(global.Apps) == 0
				for _, wildcard := range global.Apps {
					if match, _ = filepath.Match(strings.ToLower(wildcard), strings.ToLower(title)); match {
						break
					}
				}
				if !match {
					continue
				}

				for {
					if _, err := os.Stat(zipname); err != nil {
						break
					} else {
						fmt.Printf("%s exists.\n", zipname)
						index++
						zipname = mkzipname(title, folder, revisionid, index)
					}
				}
				fmt.Printf("Exporting %s\n", zipname)

				var infile, zipfile *os.File
				var w io.Writer
				if global.AddDirs && folder != `` {
					os.MkdirAll(folder, 0755)
				}
				if zipfile, err = os.Create(zipname); err != nil {
					panic(err)
				}
				zipw := zip.NewWriter(zipfile)

				// Empty file
				if w, err = zipw.Create(".export_log"); err != nil {
					panic(err)
				}

				// Metadata file
				metadata := ShaprMetadata{
					RemoteID:         name,
					RevisionID:       revisionid,
					LocalChangeCount: 0,
				}

				if w, err = zipw.Create(".metadata"); err != nil {
					panic(err)
				}

				buf, _ := json.Marshal(&metadata)
				if _, err = io.Copy(w, bytes.NewReader(buf)); err != nil {
					panic(err)
				}

				if infile, err = os.Open(workspace); err != nil {
					panic(err)
				}

				// Workspace File
				if w, err = zipw.Create(`workspace`); err != nil {
					panic(err)
				}

				if _, err = io.Copy(w, infile); err != nil {
					panic(err)
				}

				infile.Close()
				zipw.Close()
			}

		}
		rows.Close()
	}

	_ = datadb
}
