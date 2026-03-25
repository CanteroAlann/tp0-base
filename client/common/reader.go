package common

import (
	"encoding/csv"
	"os"
)

type BettingReader struct {
	file   *os.File
	reader *csv.Reader
}

func NewBettingReader(path string) (*BettingReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(f)
	//log.Infof("action: open_data_file | result: success | path: %s", path)

	return &BettingReader{
		file:   f,
		reader: reader,
	}, nil
}

func (br *BettingReader) ReadNext(agenciaID string) (UserData, error) {
	record, err := br.reader.Read()
	if err != nil {
		log.Infof("action: read_line | result: fail | client_id: %s | error: %v", agenciaID, err)
		return UserData{}, err
	}
	//log.Infof("action: read_csv_record | result: success | client_id: %s | record: %v", agenciaID, record)

	return NewUserDataFromStrings(
		agenciaID,
		record[0], // nombre
		record[1], // apellido
		record[2], // documento
		record[3], // nacimiento
		record[4], // numero
	)
}

func (br *BettingReader) Close() {
	br.file.Close()
}
