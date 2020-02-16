package cmd

import (
	"fmt"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/utils"
	"github.com/spf13/cobra"
	"os"
)

type userFlagParams struct {
	provider string
	login string
	name string
	password string
	email string
	admin bool
}

var userParams = userFlagParams{}

// userCmd represents all users command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manipulate users",
}

// addUserCmd represents the "user add" command
var addUserCmd = &cobra.Command{
	Use:   "add",
	Short: "Add local user",
	Run: addUser,
}

// listUsersCmd represents the "user list" command
var listUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Run: listUsers,
}

// showUserCmd represents the "user show" command
var showUserCmd = &cobra.Command{
	Use:   "show",
	Short: "Show user info",
	Run: showUser,
}

// delUserCmd represents the "user delete" command
var delUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete user",
	Run: delUser,
}

func init() {
	rootCmd.AddCommand(userCmd)

	// Here you will define your flags and configuration settings.
	userCmd.PersistentFlags().StringVar(&userParams.provider, "provider", common.ProviderLocal, "user provider")
	userCmd.PersistentFlags().StringVar(&userParams.login, "login", "", "user login")

	userCmd.AddCommand(addUserCmd)
	addUserCmd.Flags().StringVar(&userParams.name, "name", "", "user name")
	addUserCmd.Flags().StringVar(&userParams.name, "email", "", "user email")
	addUserCmd.Flags().StringVar(&userParams.password, "password", "", "user password")
	addUserCmd.Flags().BoolVar(&userParams.admin, "admin", false, "user admin")

	userCmd.AddCommand(listUsersCmd)
	userCmd.AddCommand(showUserCmd)
	userCmd.AddCommand(delUserCmd)
}

func addUser(cmd *cobra.Command, args []string) {
	if ! config.Authentication {
		fmt.Println("Authentication is disabled !")
		os.Exit(1)
	}

	initializeMetadataBackend()

	if userParams.login == "" {
		fmt.Println("missing user login")
	}

	user, err := metadataBackend.GetUser("local")
	if err != nil {
		fmt.Printf("Unable to get admin user : %s\n", err)
		os.Exit(1)
	}

	if user != nil {
		fmt.Println("Admin user already exists, do you want to reset password ?")
		os.Exit(1)
	}

	// Create admin user
	user = common.NewUser(common.ProviderLocal, userParams.login)
	user.Login = userParams.login
	user.Name = userParams.name
	user.Email = userParams.email
	user.IsAdmin = userParams.admin

	if userParams.password == "" {
		userParams.password = common.GenerateRandomID(32)
		fmt.Printf("Generated user password is %s\n", userParams.password)
	}

	hash, err := common.HashPassword(userParams.password)
	if err != nil {
		fmt.Printf("Unable to hash password : %s\n", err)
		os.Exit(1)
	}
	user.Password = hash

	err = metadataBackend.CreateUser(user)
	if err != nil {
		fmt.Printf("Unable to create user : %s\n", err)
		os.Exit(1)
	}
}

func showUser(cmd *cobra.Command, args []string) {
	if ! config.Authentication {
		fmt.Println("Authentication is disabled !")
		os.Exit(1)
	}

	initializeMetadataBackend()

	if userParams.login == "" {
		fmt.Println("missing user login")
	}

	id := common.GetUserId(userParams.provider, userParams.login)
	user, err := metadataBackend.GetUser(id)
	if err != nil  {
		fmt.Printf("Unable to get user : %s\n", err)
		os.Exit(1)
	}
	if user == nil {
		fmt.Printf("User %s not found\n", id)
		os.Exit(1)
	}

	utils.Dump(user)
}

func listUsers(cmd *cobra.Command, args []string) {
	if ! config.Authentication {
		fmt.Println("Authentication is disabled !")
		os.Exit(1)
	}

	initializeMetadataBackend()

	limit := 99999
	users, _, err := metadataBackend.GetUsers(userParams.provider, &common.PagingQuery{Limit: &limit})
	if err != nil  {
		fmt.Printf("Unable to get users : %s\n", err)
		os.Exit(1)
	}

	if len(users) > 0 {
		for _, user := range users {
			fmt.Println(user.String())
		}
	} else {
		fmt.Println("no users")
	}
}

func delUser(cmd *cobra.Command, args []string) {
	if ! config.Authentication {
		fmt.Println("Authentication is disabled !")
		os.Exit(1)
	}

	initializeMetadataBackend()

	if userParams.login == "" {
		fmt.Println("missing user login")
	}

	id := common.GetUserId(userParams.provider, userParams.login)
	err := metadataBackend.DeleteUser(id)
	if err != nil  {
		fmt.Printf("Unable to delete user : %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("user %s has been deleted\n", id)
}