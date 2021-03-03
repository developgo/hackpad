// +build js

package main

import (
	"errors"
	"os"
	"path/filepath"
	"syscall/js"

	"github.com/johnstarich/go-wasm/internal/interop"
	"github.com/johnstarich/go-wasm/internal/process"
	"github.com/johnstarich/go-wasm/internal/promise"
	"github.com/johnstarich/go-wasm/log"
)

func installFunc(this js.Value, args []js.Value) interface{} {
	resolve, reject, prom := promise.New()
	go func() {
		err := install(args)
		if err != nil {
			reject(interop.WrapAsJSError(err, "Failed to install binary"))
			return
		}
		resolve(nil)
	}()
	return prom
}

func install(args []js.Value) error {
	if len(args) != 1 {
		return errors.New("Expected command name to install")
	}
	command := args[0].String()
	command = filepath.Base(command) // ensure no path chars are present

	if err := os.MkdirAll("/bin", 0644); err != nil {
		return err
	}

	body, err := httpGetFetch("wasm/" + command + ".wasm")
	if err != nil {
		return err
	}
	fs := process.Current().Files()
	fd, err := fs.Open("/bin/"+command, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0750)
	if err != nil {
		return err
	}
	defer fs.Close(fd)
	if _, err := fs.Write(fd, body, 0, body.Len(), nil); err != nil {
		return err
	}
	log.Print("Install completed: ", command)
	return nil
}
