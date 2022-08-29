package main

import (
	"flag"
	"fmt"
	"os"

	"blocksui-server-node/config"
	"blocksui-server-node/server"
)

var CMDS = map[string]string{
	"node": "Runs the CRCLS node.",
	"help": "Prints the help context.",
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help":
			fmt.Println("")
			fmt.Println("Usage: crcls command [OPTIONS]")
			fmt.Println("")

			for cmd, description := range CMDS {
				fmt.Printf("  %s\t%s\n", cmd, description)
			}
			fmt.Println("")
		case "node":
			mode := flag.String("m", "production", "-m development")
			port := flag.String("p", ":80", "-p :8080")

			// Register with the Node Manager

			fmt.Println("Starting the CRCLS Node")
			config := config.New(*mode, *port)
			server.Start(config)
		default:
			fmt.Println("")
			fmt.Println("")
			fmt.Println("Usage: crcls command [OPTIONS]")
			fmt.Println("")
			fmt.Println("For list of commands please use: crcls help")
			fmt.Println("")
		}
	} else {
		fmt.Println("")
		fmt.Println("Missing command.")
		fmt.Println("")
		fmt.Println("Usage: crcls command [OPTIONS]")
		fmt.Println("")
		fmt.Println("For list of commands please use: crcls help")
		fmt.Println("")
		os.Exit(1)
	}

	os.Exit(0)
}
