package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/root-gg/utils"
	"github.com/spf13/cobra"

	"github.com/root-gg/plik/server/common"
)

type inviteFlagParams struct {
	id       string
	validity time.Duration
	admin    bool
}

var inviteParams = inviteFlagParams{}

// inviteCmd represents all invites command
var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Manipulate invites",
}

// createInviteCmd represents the "invite create" command
var createInviteCmd = &cobra.Command{
	Use:   "create",
	Short: "Create invite",
	Run:   createInvite,
}

// listInvitesCmd represents the "invite list" command
var listInvitesCmd = &cobra.Command{
	Use:   "list",
	Short: "List invites",
	Run:   listInvites,
}

// showInviteCmd represents the "invite show" command
var showInviteCmd = &cobra.Command{
	Use:   "show",
	Short: "Show invite info",
	Run:   showInvite,
}

// deleteInviteCmd represents the "invite delete" command
var deleteInviteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete invite",
	Run:   deleteInvite,
}

func init() {
	rootCmd.AddCommand(inviteCmd)

	// Here you will define your flags and configuration settings.
	inviteCmd.AddCommand(createInviteCmd)
	createInviteCmd.Flags().DurationVar(&inviteParams.validity, "validity", 30*24*time.Hour, "invite validity duration [30 days]")
	createInviteCmd.Flags().BoolVar(&inviteParams.admin, "admin", false, "invite admin")

	inviteCmd.AddCommand(listInvitesCmd)

	inviteCmd.AddCommand(showInviteCmd)
	showInviteCmd.Flags().StringVar(&inviteParams.id, "id", "", "invite id")

	inviteCmd.AddCommand(deleteInviteCmd)
	deleteInviteCmd.Flags().StringVar(&inviteParams.id, "id", "", "invite id")
}

func createInvite(cmd *cobra.Command, args []string) {
	if !config.Authentication {
		fmt.Println("Authentication is disabled !")
		os.Exit(1)
	}

	initializeMetadataBackend()

	// Create invite
	invite, err := common.NewInvite(nil, inviteParams.validity)
	if err != nil {
		fmt.Printf("unable to create invite : %s\n", err)
		os.Exit(1)
	}

	invite.Admin = inviteParams.admin

	err = metadataBackend.CreateInvite(invite)
	if err != nil {
		fmt.Printf("Unable to create invite : %s\n", err)
		os.Exit(1)
	}
}

func showInvite(cmd *cobra.Command, args []string) {
	if !config.Authentication {
		fmt.Println("Authentication is disabled !")
		os.Exit(1)
	}

	if inviteParams.id == "" {
		fmt.Println("missing invite id")
		os.Exit(1)
	}

	initializeMetadataBackend()

	invite, err := metadataBackend.GetInvite(inviteParams.id)
	if err != nil {
		fmt.Printf("Unable to get invite : %s\n", err)
		os.Exit(1)
	}
	if invite == nil {
		fmt.Printf("Invite %s not found\n", inviteParams.id)
		os.Exit(1)
	}

	utils.Dump(invite)
}

func listInvites(cmd *cobra.Command, args []string) {
	if !config.Authentication {
		fmt.Println("Authentication is disabled !")
		os.Exit(1)
	}

	initializeMetadataBackend()

	f := func(invite *common.Invite) error {
		fmt.Println(invite.String())
		return nil
	}

	err := metadataBackend.ForEachInvites(f)
	if err != nil {
		fmt.Printf("Unable to get invites : %s\n", err)
		os.Exit(1)
	}
}

func deleteInvite(cmd *cobra.Command, args []string) {
	if !config.Authentication {
		fmt.Println("Authentication is disabled !")
		os.Exit(1)
	}

	if inviteParams.id == "" {
		fmt.Println("missing invite id")
		os.Exit(1)
	}

	initializeMetadataBackend()

	// Ask confirmation
	fmt.Printf("Do you really want to delete this invite %s and all its uploads ? [y/N]\n", inviteParams.id)
	ok, err := common.AskConfirmation(false)
	if err != nil {
		fmt.Printf("Unable to ask for confirmation : %s", err)
		os.Exit(1)
	}
	if !ok {
		os.Exit(0)
	}

	deleted, err := metadataBackend.DeleteInvite(inviteParams.id)
	if err != nil {
		fmt.Printf("Unable to delete invite : %s\n", err)
		os.Exit(1)
	}

	if !deleted {
		fmt.Printf("invite %s not found\n", inviteParams.id)
		os.Exit(1)
	}

	fmt.Printf("invite %s has been deleted\n", inviteParams.id)
}
