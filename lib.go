package seqno

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"time"
)

type SeqNo struct {
	conn *gorm.DB
	logicID string
	prefixFormat string
	startWith int64
	tries int
	locker Locker
}

func NewSeqNoGenerator(db *gorm.DB,logicID string) *SeqNo {
	return &SeqNo{
		conn:db,
		logicID:logicID,
		tries:3,
		startWith:0,
	}
}

func (s *SeqNo) PrefixFormat(format string) *SeqNo {
	s.prefixFormat = format
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
	return time.Now().Format(s.prefixFormat)
}

func (s *SeqNo) genID() string{
	prefix := s.getPrefix()
	return fmt.Sprintf("s!e@q#n￥o%-%s-%s",s.logicID,prefix)
}

func (s *SeqNo) next() (int64,error) {
	if err := s.locker.LockWithTimeout(s.genID(),5 * time.Second,3 *time.Second); err != nil {
		return 0,err
	}
	defer s.locker.Unlock(s.genID())
	var lastID int64
	now := time.Now()
	for i := 0; i<s.tries;i++ {
	getSQL := fmt.Sprint("SELECT last_id FROM `seqno_generator` WHERE `logic_id` = ? AND `prefix` = ?")
	if err := s.conn.Raw(getSQL,s.logicID,s.getPrefix()).Pluck("last_id",&lastID).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return 0,errors.New("[SeqNo] can't get last id:" + s.logicID)
		}
		lastID = s.startWith
	}
		lastID = lastID + 1
		updateSQL := fmt.Sprintf("INSERT INTO `seqno_generator` (`logic_id`,`last_id`,`created_at`,`updated_at`) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE `last_id` = ?,`updated_at` = ?")
		ret := s.conn.Exec(updateSQL, s.logicID, lastID, now, now, lastID, now)
		if ret.Error != nil {
			return 0, errors.New("[SeqNo] can't update last id:" + s.logicID)
		}
	}
	return lastID,nil
}

func (s *SeqNo) Next() (int64,error) {
	return s.next()
}







