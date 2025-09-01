package main

import (
	"fmt"
	"os"

	"github.com/abetd/sqlgen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sqlgen <inDir>")
	}
	inDir := os.Args[1]
	gen := sqlgen.NewCodeGen(inDir)

	if err := gen.CodeGen(); err != nil {
		fmt.Println(err)
	}
}
