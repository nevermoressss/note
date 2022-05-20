package orm

import (
	"errors"
	"fmt"
	"strings"
)

//Delete no args
func (q *Query) Delete() (int64, error) {
	if len(q.errs) != 0 {
		return 0, errors.New(strings.Join(q.errs, ""))
	}
	var where string
	if len(q.wheres) > 0 {
		where = fmt.Sprintf(`where %s`, strings.Join(q.wheres, " and "))
	}
	st, err := q.db.Prepare(fmt.Sprintf(`delete from %s %s`, q.table, where))
	if err != nil {
		return 0, err
	}
	result, err := st.Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
