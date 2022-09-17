package main

import (
	"blocksui-node/config"
	"blocksui-node/contracts"
	"blocksui-node/server"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"
)

var (
	// Main Flags
	mainFlags = flag.NewFlagSet("main", flag.ContinueOnError)
	env       = mainFlags.String("e", "production", "-e development")

	// Balance Flags
	balanceFlags     = flag.NewFlagSet("balance", flag.ExitOnError)
	showStakeBalance = balanceFlags.Bool("stake", false, "--stake - Show staking balance")

	// Node Flags
	nodeFlags = flag.NewFlagSet("node", flag.ExitOnError)
	port      = nodeFlags.String("p", ":80", "-p :8080")
)

var CMDS = map[string]string{
	"balance":    "Returns the node's ether balance. Use --stake to get your staking balance.",
	"init":       "Initialize the CLI.",
	"node":       "Runs the CRCLS node.",
	"register":   "Register this node with the network.",
	"unregister": "Unregister this node with the network.",
	"help":       "Prints the help context.",
}

func initialize(c *config.Config) error {
	if isInitialized(c.HomeDir) {
		fmt.Println("The CLI is already initialized.")
		os.Exit(1)
	}

	// Make the config directory in the user Home directory
	if err := os.MkdirAll(filepath.Join(c.HomeDir, ".crcls"), 0755); err != nil {
		return err
	}

	if _, err := GenerateAccount(c.HomeDir); err != nil {
		return err
	}

	return nil
}

func isInitialized(homeDir string) bool {
	_, derr := os.Stat(path.Join(homeDir, ".crcls"))
	if os.IsNotExist(derr) {
		return false
	} else {
		_, kerr := os.Stat(path.Join(homeDir, ".crcls/keyfile"))
		if os.IsNotExist(kerr) {
			return false
		}
	}

	return true
}

func ensureInit(homeDir string) {
	if !isInitialized(homeDir) {
		fmt.Println("The CLI is not initialized yet. Make sure to run `bui init` first.")
		os.Exit(1)
	}
}

func main() {
	mainFlags.Parse(os.Args[1:])
	c := config.New(*env)

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
		case "balance":
			balanceFlags.Parse(os.Args[2:])

			ensureInit(c.HomeDir)
			account, err := LoadAccount(c)
			if err != nil {
				fmt.Printf("[Load Accounts] %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Address: %s\n", account.Address)

			var balance *big.Int

			if *showStakeBalance {
				if err := contracts.LoadContracts(c); err != nil {
					fmt.Printf("[Load Contracts] %v\n", err)
					os.Exit(1)
				}

				balance, err = contracts.StakeBalance(account.Address)
				if err != nil {
					fmt.Printf("[Stake Balances] %v\n", err)
					os.Exit(1)
				}
			} else {
				balance, err = account.Balance()
				if err != nil {
					fmt.Printf("[Account Balance] %v\n", err)
					os.Exit(1)
				}
			}

			fmt.Printf("Balance: %s\n", balance)
		case "init":
			if err := initialize(c); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println("Initialization complete")
			os.Exit(0)
		case "node":
			ensureInit(c.HomeDir)

			nodeFlags.Parse(os.Args[2:])
			c.WithPort(*port)

			if err := contracts.LoadContracts(c); err != nil {
				fmt.Printf("[Load Contracts] %v\n", err)
				os.Exit(1)
			}

			account, err := LoadAccount(c)
			if err != nil {
				fmt.Printf("[Load Accounts] %v\n", err)
			}

			if ok := account.VerifyStake(); !ok {
				fmt.Println("Your staking account is too low on funds. Please register again to top up your account.")
				os.Exit(1)
			}

			fmt.Printf("Account Loaded: %s\n", account.Address)

			fmt.Println("Starting the CRCLS Node")
			server.Start(c)
		case "register":
			ensureInit(c.HomeDir)

			if err := contracts.LoadContracts(c); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			account, err := LoadAccount(c)
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Account Loaded: %s\n", account.Address)

			stake, err := contracts.CalcStake(account.Address)
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Stake Required: %s\n", stake)
			balance, err := account.Balance()
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Account Balance: %s\n", balance)

			if contracts.Register(account.Sender(), account.IP, stake) {
				fmt.Println("Registration complete.")
				os.Exit(0)
			}

			os.Exit(1)
		case "unregister":
			ensureInit(c.HomeDir)

			if err := contracts.LoadContracts(c); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			account, err := LoadAccount(c)
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Account Loaded: %s\n", account.Address)

			if contracts.Unregister(account.Sender()) {
				fmt.Println("Successfully unregistered.")

				balance, err := account.Balance()
				if err != nil {
					fmt.Printf("%v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Your balance is now: %s\n", balance)
				os.Exit(0)
			}

			os.Exit(1)
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
