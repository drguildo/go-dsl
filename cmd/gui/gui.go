// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build gui

package gui

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"github.com/webview/webview"

	"3e8.eu/go/dsl"
	jsgraphs "3e8.eu/go/dsl/graphs/javascript"

	"3e8.eu/go/dsl/cmd/web/common"
)

const Enabled = true

//go:embed res
var resources embed.FS

var (
	c    *common.Client
	w    webview.WebView
	stop chan bool
)

func Run(config dsl.Config) {
	c = common.NewClient(config)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		stopWebView()
	}()

	startWebView()

	go func() {
		<-time.After(30 * time.Second)
		os.Exit(1)
	}()

	c.Close()

	os.Exit(0)
}

func getMainDataURI() string {
	style, _ := resources.ReadFile("res/style.css")
	styleDSL, _ := common.Files.ReadFile("res/dsl.css")

	data := map[string]interface{}{
		"Style":    template.CSS(style),
		"StyleDSL": template.CSS(styleDSL),
		"Script":   template.JS(jsgraphs.Script()),
	}

	buf := new(bytes.Buffer)
	tpl := template.Must(template.ParseFS(resources, "res/main.html"))
	tpl.Execute(buf, data)

	return "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func receive(stop chan bool) {
	receiver := make(chan common.StateChange, 10)
	c.RegisterReceiver(receiver)

	defer func() {
		c.UnregisterReceiver(receiver)
	}()

	for {
		select {

		case change := <-receiver:
			updateState(common.GetStateMessage(change))

		case <-stop:
			return

		}
	}
}

func updateState(msg common.Message) {
	w.Dispatch(func() {
		w.Eval("if (window.updateState) updateState(" + string(msg.JSON()) + ")")
	})
}

func showMessage(msg string) {
	msgJSON, _ := json.Marshal(msg)

	w.Dispatch(func() {
		w.Eval("showMessage(" + string(msgJSON) + ")")
	})
}

func initialized() {
	isInitialized = true

	change := c.State()
	updateState(common.GetStateMessage(change))
}

func writeArchive(state common.StateChange) (path string, err error) {
	filenameBase := state.Time.Format("dsl_20060102_150405")
	filename := filenameBase + ".zip"

	var paths = []string{
		filepath.Join(xdg.UserDirs.Download, filename),
		filepath.Join(xdg.Home, filename),
	}

	var f *os.File

	for i := range paths {
		path = paths[i]
		f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)

		if i < len(paths)-1 && errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return
		} else {
			break
		}
	}

	defer f.Close()

	err = common.WriteArchive(f, filenameBase, state)
	return
}

func save() {
	change := c.State()
	if change.State != common.StateReady {
		return
	}

	path, err := writeArchive(change)

	path, _ = filepath.Abs(path)

	if err == nil {
		showMessage(fmt.Sprintf("Saved to %s.", path))
	} else if errors.Is(err, fs.ErrExist) {
		showMessage(fmt.Sprintf("File %s already exists.", path))
	} else if err != nil {
		showMessage("Saving failed!")
		fmt.Println("failed to save archive:", err)
	}
}

func setPassword(data string) {
	err := c.SetPassword(data)
	if err != nil {
		fmt.Println("unexpected call to setPassword")
	}
}

func setPassphrase(data string) {
	err := c.SetPassphrase(data)
	if err != nil {
		fmt.Println("unexpected call to setPassphrase")
	}
}

func startWebView() {
	w = webview.New(false)
	defer w.Destroy()

	w.SetTitle("xDSL stats")

	w.SetSize(620, 300, webview.HintMin)
	w.SetSize(620, 600, webview.HintNone)

	stop = make(chan bool, 1)
	go receive(stop)

	w.Bind("goInitialized", initialized)
	w.Bind("goSave", save)
	w.Bind("goSetPassword", setPassword)
	w.Bind("goSetPassphrase", setPassphrase)

	script, _ := resources.ReadFile("res/script.js")
	w.Init(string(script))

	w.Navigate(getMainDataURI())

	w.Run()
}

func stopWebView() {
	if w != nil {
		stop <- true
		w.Terminate()
	}
}
