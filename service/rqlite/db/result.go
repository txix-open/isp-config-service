package db

type Result struct {
	LastInsertIdValue int64             `json:"last_insert_id"`
	RowsAffectedValue int64             `json:"rows_affected"`
	Types             map[string]string `json:"types"`
	Time              float64           `json:"time"`
	Rows              any               `json:"rows"`
	Error             string            `json:"error"`
}

func (r Result) LastInsertId() (int64, error) {
	return r.LastInsertIdValue, nil
}

func (r Result) RowsAffected() (int64, error) {
	return r.RowsAffectedValue, nil
}

type Response struct {
	Results []*Result `json:"results"`
	Time    float64   `json:"time"`
}
