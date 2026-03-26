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

const (
	MsgTypeBets   byte = 1
	MsgTypeFinish byte = 2
	MsgTypeQuery  byte = 3
	AckMessage    byte = 4
)

func SendBets(conn net.Conn, br *BettingReader, batchMaxAmount int, agenciaID string) (int, error) {
	sentCount := 0
	var userDataList []UserData

	for i := 0; i < batchMaxAmount; i++ {
		userData, err := br.ReadNext(agenciaID)

		if err != nil {
			if err.Error() == "EOF" {
				if len(userDataList) > 0 {
					log.Infof("action: send_batch_message | result: success | batch_size: %v", len(userDataList))
					batchMessage, err2 := NewBatchMessage(userDataList)
					if err2 != nil {
						return sentCount, err2
					}
					if err2 := SendBatchMessage(conn, batchMessage); err2 != nil {
						return sentCount, err2
					}
					sentCount += len(userDataList)
				}
				return sentCount, err
			}
			return sentCount, err
		}
		userDataList = append(userDataList, userData)
	}

	batchMessage, err := NewBatchMessage(userDataList)
	if err != nil {
		return sentCount, err
	}

	if err := SendBatchMessage(conn, batchMessage); err != nil {
		return sentCount, err
	}
	log.Infof("action: send_batch_message | result: success | batch_size: %v", len(userDataList))
	sentCount += len(userDataList)
	return sentCount, nil
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

	if err := binary.Write(buf, binary.BigEndian, MsgTypeBets); err != nil {
		return err
	}

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

func SendFinishMessage(conn net.Conn, agenciaID string) error {
	agencia, err := strconv.ParseUint(agenciaID, 10, 16)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, MsgTypeFinish)
	binary.Write(buf, binary.BigEndian, uint16(agencia))
	_, err = conn.Write(buf.Bytes())
	return err
}

func SendQueryMessage(conn net.Conn, agenciaID string) (uint32, []uint32, error) {
	agencia, err := strconv.ParseUint(agenciaID, 10, 16)
	if err != nil {
		return 0, nil, err
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, MsgTypeQuery)
	binary.Write(buf, binary.BigEndian, uint16(agencia))

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return 0, nil, err
	}

	// var count uint32
	// if err := binary.Read(conn, binary.BigEndian, &count); err != nil {
	// 	return 0, nil, err
	// }

	// winners := make([]uint32, count)
	// for i := uint32(0); i < count; i++ {
	// 	if err := binary.Read(conn, binary.BigEndian, &winners[i]); err != nil {
	// 		return 0, nil, err
	// 	}
	// }

	return 0, nil, nil
}
