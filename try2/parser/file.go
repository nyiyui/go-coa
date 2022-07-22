package parser

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func (e *Env) LoadPathOnly(path string) (re *Nodes, err error) {
	var data []byte
	resp, err := (&http.Client{Timeout: 1 * time.Second}).Get(path)
	if err != nil {
		if strings.Contains(err.Error(), "unsupported protocol scheme") {
			goto LocalFS
		}
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err2 := Body.Close()
		if err2 != nil {
			err = err2
		}
	}(resp.Body)
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
LocalFS:
	if data == nil {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read local: %w", err)
		}
	}
	root := Nodes{}
	err = Parser.ParseBytes(path, data, &root)
	if err != nil {
		return nil, err
	}
	err = check(&root)
	if err != nil {
		return nil, err
	}
	return &root, nil
}
func (e *Env) LoadPath(path string) (re Evaler, err error) {
	root, err := e.LoadPathOnly(path)
	return root.Eval(e, 0)
}
