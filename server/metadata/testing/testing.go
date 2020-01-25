/**

    Plik upload server

The MIT License (MIT)

Copyright (c) <2015> Copyright holders list can be found in AUTHORS file
	- Mathieu Bodjikian <mathieu@bodjikian.fr>
	- Charles-Antoine Mathieu <skatkatt@root.gg>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
**/

package testing

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/root-gg/juliet"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/metadata"
)

// Ensure Testing Metadata Backend implements metadata.Backend interface
var _ metadata.Backend = (*Backend)(nil)

// Backend backed in-memory for testing purpose
type Backend struct {
	uploads map[string]*common.Upload
	users   map[string]*common.User

	err error
	mu  sync.Mutex
}

// NewBackend create a new Testing Backend
func NewBackend() (b *Backend) {
	b = new(Backend)
	b.uploads = make(map[string]*common.Upload)
	b.users = make(map[string]*common.User)
	return b
}

// Upsert create or update upload metadata
func (b *Backend) CreateUpload(ctx *juliet.Context, upload *common.Upload) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	u, err := defCopyUpload(upload)
	if err != nil {
		return err
	}

	if _, ok := b.uploads[upload.ID] ; ok {
		return errors.New("Upload already exists")
	}

	b.uploads[upload.ID] = u

	return nil
}

func (b *Backend) get(ctx *juliet.Context, id string) (upload *common.Upload, err error) {
	if upload, ok := b.uploads[id]; ok {
		return upload, nil
	}

	return nil, errors.New("upload does not exists")
}

// Get upload metadata
func (b *Backend) GetUpload(ctx *juliet.Context, id string) (upload *common.Upload, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	upload, err = b.get(ctx, id)

	if err == nil {
		upload, err = defCopyUpload(upload)
		if err != nil {
			return nil, err
		}
	}

	return upload, err
}

// Remove upload metadata
func (b *Backend) RemoveUpload(ctx *juliet.Context, upload *common.Upload) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	upload, err = b.get(ctx, upload.ID)
	if err != nil {
		return err
	}

	delete(b.uploads, upload.ID)

	return nil
}

func (b *Backend) AddFile(ctx *juliet.Context, upload *common.Upload, newFile *common.File) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	upload, err = b.get(ctx, upload.ID)
	if err != nil {
		return err
	}

	if _, ok := upload.Files[newFile.ID]; ok {
		return errors.New("File exists")
	}

	newFile, err = defCopyFile(newFile)
	if err != nil {
		return err
	}

	upload.Files[newFile.ID] = newFile

	return nil
}

func (b *Backend) UpdateFile(ctx *juliet.Context, upload *common.Upload, newFile *common.File) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	upload, err = b.get(ctx, upload.ID)
	if err != nil {
		return err
	}

	if oldFile, ok := upload.Files[newFile.ID]; ok {
		if oldFile.Status != common.FILE_UPLOADING {
			return errors.New("Cannot update file whose status is not uploading")
		}

		newFile, err = defCopyFile(newFile)
		if err != nil {
			return err
		}

		upload.Files[newFile.ID] = newFile
	} else {
		return errors.New("File does not exists")
	}

	upload.Files[newFile.ID] = newFile

	return nil
}

func (b *Backend) SetFileStatus(ctx *juliet.Context, upload *common.Upload, file *common.File, oldStatus string, newStatus string) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	upload, err = b.get(ctx, upload.ID)
	if err != nil {
		return err
	}

	if f, ok := upload.Files[file.ID]; ok {
		if file.Status == oldStatus {
			f.Status = newStatus
		} else {
			return fmt.Errorf("invalid file status %s, expecting %s", f.Status, oldStatus)
		}
	} else {
		return errors.New("file not found")
	}

	return nil
}

// SaveUser create or update user
func (b *Backend) CreateUser(ctx *juliet.Context, user *common.User) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	user, err = defCopyUser(user)
	if err != nil {
		return err
	}

	b.users[user.ID] = user

	return nil
}

func (b *Backend) getUser(ctx *juliet.Context, id string, token string) (user *common.User, err error) {
	if id != "" {
		if user, ok := b.users[id]; ok {
			return user, nil
		}
	} else if token != "" {
		for _, u := range b.users {
			for _, t := range u.Tokens {
				if t.Token == token {
					user = u
					return u, nil
				}
			}
		}
	}

	return nil, nil
}

// GetUser get a user
func (b *Backend) GetUser(ctx *juliet.Context, id string, token string) (user *common.User, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	user, err = b.getUser(ctx, id, token)

	if err == nil {
		user, err = defCopyUser(user)
	}

	return user, err
}

// RemoveUser remove a user
func (b *Backend) RemoveUser(ctx *juliet.Context, user *common.User) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	_, err = b.getUser(ctx, user.ID, "")
	if err != nil {
		return err
	}

	delete(b.users, user.ID)

	return nil
}

func (b *Backend) CreateUserToken(ctx *juliet.Context, user *common.User, token *common.Token) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	user, err = b.getUser(ctx, user.ID, "")
	if err != nil {
		return err
	}

	for _, t := range user.Tokens {
		if t.Token == token.Token {
			return errors.New("token already exists")
		}
	}

	user.Tokens = append(user.Tokens, token)

	return nil
}

func (b *Backend) RevokeUserToken(ctx *juliet.Context, user *common.User, token *common.Token) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return b.err
	}

	user, err = b.getUser(ctx, user.ID, "")
	if err != nil {
		return err
	}

	validTokens := user.Tokens[:0]
	for _, t := range user.Tokens {
		if t.Token != token.Token {
			validTokens = append(validTokens, t)
		}
	}

	user.Tokens = validTokens

	return nil
}

// GetUserUploads return a user uploads
func (b *Backend) GetUserUploads(ctx *juliet.Context, user *common.User, token *common.Token) (ids []string, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	return b.getUserUploads(ctx, user, token)
}

func (b *Backend) getUserUploads(ctx *juliet.Context, user *common.User, token *common.Token) (ids []string, err error) {
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
func (b *Backend) GetUserStatistics(ctx *juliet.Context, user *common.User, token *common.Token) (stats *common.UserStats, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	stats = &common.UserStats{}

	ids, err := b.getUserUploads(ctx, user, token)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		upload, err := b.get(ctx, id)
		if err != nil {
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

// GetUsers return all user ids
func (b *Backend) GetUsers(ctx *juliet.Context) (ids []string, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	for id := range b.users {
		ids = append(ids, id)
	}

	return ids, nil
}

// GetServerStatistics return server statistics
func (b *Backend) GetServerStatistics(ctx *juliet.Context) (stats *common.ServerStats, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	stats = new(common.ServerStats)

	byTypeAggregator := common.NewByTypeAggregator()

	for _, upload := range b.uploads {
		stats.AddUpload(upload)

		for _, file := range upload.Files {
			byTypeAggregator.AddFile(file)
		}
	}

	stats.FileTypeByCount = byTypeAggregator.GetFileTypeByCount(10)
	stats.FileTypeBySize = byTypeAggregator.GetFileTypeBySize(10)

	stats.Users = len(b.users)

	return
}

// GetUploadsToRemove return expired upload ids
func (b *Backend) GetUploadsToRemove(ctx *juliet.Context) (ids []string, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return nil, b.err
	}

	for id, upload := range b.uploads {
		if upload.IsExpired() {
			ids = append(ids, id)
		}
	}

	return ids, nil
}

// SetError sets the error any subsequent method other call will return
func (b *Backend) SetError(err error) {
	b.err = err
}

func defCopyUpload(upload *common.Upload) (u *common.Upload, err error) {
	u = &common.Upload{}
	j, err := json.Marshal(upload)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, u)
	if err != nil {
		return nil, err
	}
	return u, err
}

func defCopyFile(file *common.File) (f *common.File, err error) {
	f = &common.File{}
	j, err := json.Marshal(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, f)
	if err != nil {
		return nil, err
	}
	return f, err
}


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

func defCopyToken(token *common.Token) (t *common.Token, err error) {
	t = &common.Token{}
	j, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, t)
	if err != nil {
		return nil, err
	}
	return t, err
}