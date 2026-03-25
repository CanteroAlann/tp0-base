package common

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
	"time"
)

type Header struct {
	NombreLength   uint16
	ApellidoLength uint16
}

type Message struct {
	PayloadSize uint32
	Payload     []byte
}

type BatchMessage struct {
	Amount   uint32
	Messages []Message
}

type UserData struct {
	Agencia    uint16
	Nombre     string
	Apellido   string
	Nacimiento time.Time
	Documento  uint32
	Numero     uint32
}

func SendData(conn net.Conn, batchMaxAmount int, dataPath string, agenciaID string) (int, error) {
	br, err := NewBettingReader(dataPath)
	if err != nil {
		log.Infof("action: create_betting_reader | result: fail | client_id: %s | error: %v", agenciaID, err)
		return 0, err
	}
	defer br.Close()

	sentCount := 0
	for {
		var userDataList []UserData

		for i := 0; i < batchMaxAmount; i++ {
			userData, err := br.ReadNext(agenciaID)

			if err != nil {
				if err.Error() == "EOF" {
					return sentCount, nil
				}
				return sentCount, err
			}
			log.Infof("action: read_user_data | result: success | client_id: %s | user_data: %v", agenciaID, userData)
			userDataList = append(userDataList, userData)
		}

		batchMessage, err := NewBatchMessage(userDataList)
		if err != nil {
			return sentCount, err
		}

		if err := SendBatchMessage(conn, batchMessage); err != nil {
			return sentCount, err
		}

		sentCount += len(userDataList)
	}
}

func NewUserDataFromStrings(agenciaRaw, nombre, apellido, documentoRaw, nacimientoRaw, numeroRaw string) (UserData, error) {
	agencia, err := strconv.ParseUint(agenciaRaw, 10, 16)
	if err != nil {
		return UserData{}, err
	}

	nacimiento, err := time.Parse("2006-01-02", nacimientoRaw)
	if err != nil {
		return UserData{}, err
	}

	documento, err := strconv.ParseUint(documentoRaw, 10, 32)
	if err != nil {
		return UserData{}, err
	}

	numero, err := strconv.ParseUint(numeroRaw, 10, 32)
	if err != nil {
		return UserData{}, err
	}

	return UserData{
		Agencia:    uint16(agencia),
		Nombre:     nombre,
		Apellido:   apellido,
		Nacimiento: nacimiento,
		Documento:  uint32(documento),
		Numero:     uint32(numero),
	}, nil
}

func NewMessage(u UserData) (Message, error) {
	nombreBytes := []byte(u.Nombre)
	apellidoBytes := []byte(u.Apellido)

	header := Header{
		NombreLength:   uint16(len(nombreBytes)),
		ApellidoLength: uint16(len(apellidoBytes)),
	}

	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, header); err != nil {
		return Message{}, err
	}

	if err := binary.Write(buf, binary.BigEndian, u.Agencia); err != nil {
		return Message{}, err
	}

	if _, err := buf.Write(nombreBytes); err != nil {
		return Message{}, err
	}

	if _, err := buf.Write(apellidoBytes); err != nil {
		return Message{}, err
	}

	if err := binary.Write(buf, binary.BigEndian, uint16(u.Nacimiento.Year())); err != nil {
		return Message{}, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint8(u.Nacimiento.Month())); err != nil {
		return Message{}, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint8(u.Nacimiento.Day())); err != nil {
		return Message{}, err
	}

	if err := binary.Write(buf, binary.BigEndian, u.Documento); err != nil {
		return Message{}, err
	}
	if err := binary.Write(buf, binary.BigEndian, u.Numero); err != nil {
		return Message{}, err
	}

	return Message{
		PayloadSize: uint32(buf.Len()),
		Payload:     buf.Bytes(),
	}, nil
}

func MessageSize(m Message) uint32 {
	return m.PayloadSize + 4
}

func NewBatchMessage(userDataList []UserData) (BatchMessage, error) {
	messages := make([]Message, len(userDataList))

	for i, userData := range userDataList {
		msg, err := NewMessage(userData)
		if err != nil {
			return BatchMessage{}, err
		}
		messages[i] = msg
	}

	return BatchMessage{
		Amount:   uint32(len(messages)),
		Messages: messages,
	}, nil
}

func SendBatchMessage(conn net.Conn, batch BatchMessage) error {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, batch.Amount); err != nil {
		return err
	}

	for _, msg := range batch.Messages {
		if err := binary.Write(buf, binary.BigEndian, msg.PayloadSize); err != nil {
			return err
		}
		if _, err := buf.Write(msg.Payload); err != nil {
			return err
		}
	}

	_, err := conn.Write(buf.Bytes())
	return err
}
