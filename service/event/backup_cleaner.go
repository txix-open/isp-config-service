package event

import (
	"context"
	"fmt"
	"os"
	"strings"

	"isp-config-service/service/rqlite"

	"github.com/txix-open/isp-kit/log"
)

type BackupCleaner struct {
	leaderChecker LeaderChecker
	cfg           *rqlite.Backup
	logger        log.Logger
}

func NewBackupCleaner(
	leaderChecker LeaderChecker,
	cfg *rqlite.Backup,
	logger log.Logger,
) BackupCleaner {
	return BackupCleaner{
		leaderChecker: leaderChecker,
		cfg:           cfg,
		logger:        logger,
	}
}

func (c BackupCleaner) Do(ctx context.Context) {
	ctx = log.ToContext(ctx, log.String("worker", "backupCleaner"))
	if !c.leaderChecker.IsLeader() {
		c.logger.Debug(ctx, "is not a leader, skip work")
		return
	}

	files, err := os.ReadDir(c.cfg.Sub.Dir)
	if err != nil {
		c.logger.Error(ctx, fmt.Sprintf("failed to read files from %s: %s", c.cfg.Sub.Dir, err.Error()))
		return
	}

	backups := make([]os.DirEntry, 0)
	for _, file := range files {
		if !strings.Contains(file.Name(), "backup") {
			continue
		}

		backups = append(backups, file)
	}

	if len(backups) <= c.cfg.Amount {
		return
	}

	deleteAmount := len(backups) - c.cfg.Amount
	deleted := 0

	for i := range deleteAmount {
		err = os.Remove(c.cfg.Sub.Dir + "/" + backups[i].Name())
		if err != nil {
			c.logger.Error(ctx, fmt.Sprintf("failed to remove %s: %s", backups[i].Name(), err.Error()))
			continue
		}

		deleted++
		c.logger.Debug(ctx, fmt.Sprintf("delete backup %s", backups[i].Name()))
	}

	c.logger.Debug(ctx, fmt.Sprintf("delete %d old backups", deleted))
}
