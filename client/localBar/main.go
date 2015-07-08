package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"go.iondynamics.net/iDhelper/crypto"

	"go.iondynamics.net/localBar/core"
)

func main() {
	sourcePtr := flag.String("source", os.Args[0], "")
	goosPtr := flag.String("goos", os.Getenv("GOOS"), "")
	goarchPtr := flag.String("goarch", os.Getenv("GOARCH"), "")
	namePtr := flag.String("name", "localbar_build", "")
	serverPtr := flag.String("server", "localhost:1842", "")
	secretPtr := flag.String("secret", "insecure", "")

	flag.Parse()

	dir, err := filepath.Abs(filepath.Dir(*sourcePtr))
	if err != nil {
		panic(err)
	}

	path, err := core.GoBuild(dir, *goosPtr, *goarchPtr)
	if err != nil {
		panic(err)
	}

	binByt, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	blob := core.Blob{
		Name:   *namePtr,
		Binary: binByt,
	}
	blobByt, err := json.Marshal(blob)
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", *serverPtr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Fprintln(conn, crypto.Encrypt(*secretPtr, string(blobByt)))

	reader := bufio.NewReader(conn)
	answer, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(answer)
	}

	fmt.Println(os.Remove(path))
}
