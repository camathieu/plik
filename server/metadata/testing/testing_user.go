package testing

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/root-gg/plik/server/common"
)

// CreateUser create user metadata
func (b *Backend) CreateUser(user *common.User) (err error) {
	if user == nil {
		return fmt.Errorf("missing user")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	if _, ok := b.users[user.ID]; ok {
		return errors.New("user already exists")
	}

	user, err = defCopyUser(user)
	if err != nil {
		return err
	}

	b.users[user.ID] = user

	return nil
}

// GetUser retrieve user metadata
func (b *Backend) GetUser(userID string) (user *common.User, err error) {
	if userID == "" {
		return nil, fmt.Errorf("missing user id")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	user, ok := b.users[userID]
	if !ok {
		return nil, nil
	}

	user, err = defCopyUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieve user metadata
func (b *Backend) GetUserFromToken(token string) (user *common.User, err error) {
	if token == "" {
		return nil, fmt.Errorf("missing user token")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

LOOP:
	for _, u := range b.users {
		for _, t := range u.Tokens {
			if token == t.Token {
				user = u
				break LOOP
			}
		}
	}

	if user == nil {
		return nil, nil
	}

	user, err = defCopyUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser update user metadata
func (b *Backend) UpdateUser(user *common.User, tx common.UserTx) (u *common.User, err error) {
	if user == nil {
		return nil, fmt.Errorf("missing user")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	user, ok := b.users[user.ID]
	if !ok {
		return nil, errors.New("user does not exists")
	}

	u, err = defCopyUser(user)
	if err != nil {
		return nil, err
	}

	err = tx(u)
	if err != nil {
		return nil, err
	}

	u, err = defCopyUser(u)
	if err != nil {
		return nil, err
	}

	// Avoid the possibility to override an other upload by changing the upload.ID in the tx
	b.users[user.ID] = u

	u, err = defCopyUser(u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// Remove user metadata
func (b *Backend) RemoveUser(user *common.User) (err error) {
	if user == nil {
		return fmt.Errorf("missing user")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	if _, ok := b.users[user.ID]; !ok {
		return errors.New("user does not exists")
	}

	delete(b.users, user.ID)

	return nil
}

// GetUserUploads return a user uploads
func (b *Backend) GetUserUploads(user *common.User, token *common.Token) (ids []string, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	return b.getUserUploads(user, token)
}

func (b *Backend) getUserUploads(user *common.User, token *common.Token) (ids []string, err error) {
	if user == nil {
		return nil, errors.New("Missing user")
	}
	for _, upload := range b.uploads {
		if upload.User != user.ID {
			continue
		}
		if token != nil && upload.Token != token.Token {
			continue
		}

		ids = append(ids, upload.ID)
	}

	return ids, nil
}

// GetUserStatistics return a user statistics
func (b *Backend) GetUserStatistics(user *common.User, token *common.Token) (stats *common.UserStats, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	stats = &common.UserStats{}

	ids, err := b.getUserUploads(user, token)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		upload, ok := b.uploads[id]
		if !ok {
			continue
		}

		stats.Uploads++

		for _, file := range upload.Files {
			stats.Files++
			stats.TotalSize += file.CurrentSize
		}
	}

	return stats, nil
}

// Create a defensive copy of the user object
func defCopyUser(user *common.User) (u *common.User, err error) {
	u = &common.User{}
	j, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, u)
	if err != nil {
		return nil, err
	}
	return u, err
}
