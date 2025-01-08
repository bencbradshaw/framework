package framework

import (
	"flag"
	"fmt"
	"os"
)

func FramworkCli() {
	if len(os.Args) < 2 {
		fmt.Println("Expected 'create-entity' or 'create-route' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create-entity":
		createEntityCmd := flag.NewFlagSet("create-entity", flag.ExitOnError)
		entityName := createEntityCmd.String("name", "", "Name of the entity")
		createEntityCmd.Parse(os.Args[2:])
		if *entityName == "" {
			fmt.Println("Please provide a name for the entity using -name flag")
			os.Exit(1)
		}
		fmt.Printf("Creating a new entity: %s\n", *entityName)
		// Add your entity creation logic here

	case "create-route":
		createRouteCmd := flag.NewFlagSet("create-route", flag.ExitOnError)
		routePath := createRouteCmd.String("path", "", "Path of the route")
		createRouteCmd.Parse(os.Args[2:])
		if *routePath == "" {
			fmt.Println("Please provide a path for the route using -path flag")
			os.Exit(1)
		}
		fmt.Printf("Creating a new route: %s\n", *routePath)
		// Add your route creation logic here

	default:
		fmt.Println("Expected 'create-entity' or 'create-route' subcommands")
		os.Exit(1)
	}
}
