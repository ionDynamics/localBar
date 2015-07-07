package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/streamrail/concurrent-map"
	"go.iondynamics.net/iDhelper/crypto"

	"go.iondynamics.net/localBar/core"
)

var secret = flag.String("secret", "insecure", "")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	listenAddr := flag.String("listen", ":1842", "")
	flag.Parse()

	ln, err := net.Listen("tcp", *listenAddr)
	eh(err)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)
	binByt, err := reader.ReadBytes('\n')
	eh(err)
	blob := &core.Blob{}
	err = json.Unmarshal([]byte(crypto.Decrypt(*secret, string(binByt))), blob)
	eh(err)

	configs := make(map[string]core.Config)

	cfgByt, err := ioutil.ReadFile("config.json")
	eh(err)

	err = json.Unmarshal(cfgByt, &configs)
	eh(err)

	cfg, ok := configs[blob.Name]
	if !ok {
		cfg, ok = configs["default"]
		if !ok {
			panic("no config")
		}
	}

	replacer := core.Replacer{
		Map: map[string]string{
			"<<NAME>>":       blob.Name,
			"<<BINARYNAME>>": cfg.BinaryName,
		},
	}

	cfg.BinaryName = replacer.Replace(cfg.BinaryName)
	replacer.Map["<<BINARYNAME>>"] = cfg.BinaryName

	cfg.Workspace = filepath.Clean(replacer.Replace(cfg.Workspace))
	replacer.Map["<<WORKSPACE>>"] = cfg.Workspace

	cfg.RunCommand = replacer.Replace(cfg.RunCommand)
	replacer.Map["<<RUNCOMMAND>>"] = cfg.RunCommand

	replacedAssets := make(map[string][]byte)
	for key, val := range cfg.Assets {
		replacedAssets[replacer.Replace(key)] = val
	}
	cfg.Assets = replacedAssets

	PrepareWorkspace(cfg, blob)
	go Run(blob.Name, cfg)

}

func eh(err error) {
	if err != nil {
		panic(err)
	}
}

func PrepareWorkspace(cfg core.Config, blob *core.Blob) {

	Stop(blob.Name)
	os.MkdirAll(cfg.Workspace, 0600)
	binPath := filepath.Clean(cfg.Workspace + string(filepath.Separator) + cfg.BinaryName)
	fmt.Println(os.Remove(binPath))

	for key, val := range cfg.Assets {
		ioutil.WriteFile(filepath.Clean(cfg.Workspace+string(filepath.Separator)+key), val, 0600)
	}
	ioutil.WriteFile(binPath, blob.Binary, 0600)
}

var cancelChannels = cmap.New()

func Stop(name string) {
	cancel, ok := cancelChannels.Get(name)
	if ok {
		cancel.(chan bool) <- true
		//@TODO: Wait for real process shutdown
		time.Sleep(1 * time.Second)
	}
}

func Run(name string, cfg core.Config) {
	var cancel chan bool
	cif, ok := cancelChannels.Get(name)
	if !ok {
		cancel = make(chan bool)
		cancelChannels.Set(name, cancel)
	} else {
		cancel = cif.(chan bool)
	}

	cmd := exec.Command(cfg.RunCommand, cfg.RunArgs...)
	cmd.Dir = cfg.Workspace
	done := make(chan bool)

	go func(c *exec.Cmd) {

		logfile := filepath.Clean(cfg.Workspace + string(filepath.Separator) + "log.txt")

		file, err := os.OpenFile(logfile, os.O_APPEND, 0600)
		if err != nil {
			file, err = os.Create(logfile)
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		defer file.Close()

		c.Stdout = file
		c.Stderr = file

		err = c.Start()
		if err != nil {
			fmt.Fprintln(file, "\n", err)
		}
		err = c.Wait()
		done <- true
		if err != nil {
			fmt.Fprintln(file, "\n", err)
		}
	}(cmd)

	select {

	case <-cancel:
		err := cmd.Process.Kill()
		<-done
		if err != nil {
			fmt.Println(err)
		}
		cancelChannels.Remove(name)

	case <-done:
		cancelChannels.Remove(name)
	}

}
