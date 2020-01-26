package testing

import (
	"github.com/root-gg/plik/server/common"
)

// GetUsers return all user ids
func (b *Backend) GetUsers() (ids []string, err error) {
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
func (b *Backend) GetServerStatistics() (stats *common.ServerStats, err error) {
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
func (b *Backend) GetUploadsToRemove() (ids []string, err error) {
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
