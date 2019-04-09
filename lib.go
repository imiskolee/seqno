package seqno

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"time"
)

type SeqNo struct {
	conn         *gorm.DB
	logicID      string
	prefixFormat string
	startWith    int64
	tries        int
	step int
	locker       Locker
}

func NewSeqNoGenerator(db *gorm.DB, logicID string) *SeqNo {
	return &SeqNo{
		conn:      db,
		logicID:   logicID,
		tries:     3,
		startWith: 1,
		step : 1,
	}
}

func (s *SeqNo) PrefixFormat(format string) *SeqNo {
	s.prefixFormat = format
	return s
}


func (s *SeqNo) Step(step int) *SeqNo {
	s.step = step
	return s
}

func (s *SeqNo) StartWith(start int64) *SeqNo {
	s.startWith = start
	return s
}

func (s *SeqNo) Locker(locker Locker) *SeqNo {
	s.locker = locker
	return s
}

func (s *SeqNo) getPrefix() string {
	if s.prefixFormat != "" {
		return time.Now().Format(s.prefixFormat)
	}
	return ""
}

func (s *SeqNo) genID() string {
	prefix := s.getPrefix()
	return fmt.Sprintf("s!e@q#n￥o%s-%s", s.logicID, prefix)
}

func (s *SeqNo) next() (int64, error) {
	objectID := s.genID()
	if err := s.locker.LockWithTimeout(objectID, 5*time.Second, 3*time.Second); err != nil {
		return 0, err
	}
	defer s.locker.Unlock(objectID)
	now := time.Now()
	prefix := s.getPrefix()
	updateSQL := fmt.Sprintf("INSERT INTO `seqno_generator` (`logic_id`,`prefix`,`last_id`,`created_at`,`updated_at`) VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE `last_id` = `last_id` + ?,`updated_at` = ?")
	ret := s.conn.Exec(updateSQL, s.logicID, prefix, s.startWith, now, now,  s.step,now)
	if ret.Error != nil {
		return 0, errors.New("[SeqNo] can't update last id:" + s.logicID + "  " + ret.Error.Error())
	}
	var lastIDs []int64
	var lastID = s.startWith
	getSQL := fmt.Sprint("SELECT last_id FROM `seqno_generator` WHERE `logic_id` = ? AND `prefix` = ? LIMIT 1")
	if err := s.conn.Raw(getSQL, s.logicID, prefix).Pluck("last_id", &lastIDs).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return 0, errors.New("[SeqNo] can't get last id:" + s.logicID + " " + err.Error())
		}
	}
	if len(lastIDs) > 0 {
		lastID = lastIDs[0]
	}
	return lastID, nil
}

func (s *SeqNo) Next() (int64, error) {
	return s.next()
}
