package scenarios

import (
	"bytes"
	"log"
	"os/exec"
)

//go:generate go-bindata -ignore (go|rb)$ -pkg scenarios .

// Load executes the sql script at `path` on postgres database at `url`
func Load(url string, path string) {
	sql, err := Asset(path)

	if err != nil {
		log.Panic(err)
	}

	reader := bytes.NewReader(sql)
	cmd := exec.Command("psql", url)
	cmd.Stdin = reader

	err = cmd.Run()

	if err != nil {
		log.Panic(err)
	}

}
