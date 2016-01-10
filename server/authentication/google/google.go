package google

import (
	"fmt"
	"github.com/root-gg/plik/server/Godeps/_workspace/src/github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/Godeps/_workspace/src/golang.org/x/oauth2"
	"github.com/root-gg/plik/server/Godeps/_workspace/src/golang.org/x/oauth2/google"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/metadataBackend"
	api_oauth2 "google.golang.org/api/oauth2/v2"
)

func GetUserConsentUrl(ctx *juliet.Context, origin string) (url string, err error) {
	if common.Config.GoogleApiClientID == "" || common.Config.GoogleApiSecret == "" {
		err = fmt.Errorf("Missing google api credentials")
		return
	}

	conf := &oauth2.Config{
		ClientID:     common.Config.GoogleApiClientID,
		ClientSecret: common.Config.GoogleApiSecret,
		RedirectURL:  origin + "/auth/google/callback",
		Scopes: []string{
			api_oauth2.UserinfoEmailScope,
			api_oauth2.UserinfoProfileScope,
		},
		Endpoint: google.Endpoint,
	}

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	url = conf.AuthCodeURL("state")
	return
}

func Callback(ctx *juliet.Context, state string, code string) (user *common.User, err error) {
	if common.Config.GoogleApiClientID == "" || common.Config.GoogleApiSecret == "" {
		err = fmt.Errorf("Missing google api credentials")
		return
	}

	conf := &oauth2.Config{
		ClientID:     common.Config.GoogleApiClientID,
		ClientSecret: common.Config.GoogleApiSecret,
		RedirectURL:  origin + "/auth/google/callback",
		Scopes: []string{
			api_oauth2.UserinfoEmailScope,
			api_oauth2.UserinfoProfileScope,
		},
		Endpoint: google.Endpoint,
	}

	if code == "" {
		err = fmt.Errorf("Missing authorization code")
		return
	}

	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		err = fmt.Errorf("Unable to get google api token : %s", err)
		return
	}

	client, err := api_oauth2.New(conf.Client(oauth2.NoContext, token))
	if err != nil {
		err = fmt.Errorf("Unable to create api client : %s", err)
		return
	}

	userInfo, err := client.Userinfo.Get().Do()
	if err != nil {
		err = fmt.Errorf("Unable to get userinfo : %s", err)
		return
	}

	// Get user from metadata backend
	user, err = metadataBackend.GetMetaDataBackend().GetUser(ctx, userInfo.Email, "")
	if err != nil {
		err = fmt.Errorf("Unable to get user from metadata backend : %s", err)
		return
	}

	if user == nil {
		if common.IsWhitelisted(ctx) {
			// Create new user
			user = common.NewUser()
			user.ID = "google:" + userInfo.Id
			user.Email = userInfo.Email
			user.Name = userInfo.Name
		} else {
			err = fmt.Errorf("Unable to create user from untrusted source IP address : %s", err)
			return
		}
	}

	return
}
