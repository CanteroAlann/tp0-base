package common

import (
	"bytes"
	"encoding/binary"
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

type UserData struct {
	Nombre     string
	Apellido   string
	Nacimiento time.Time
	Documento  uint32
	Numero     uint32
}

func NewUserDataFromStrings(nombre, apellido, documentoRaw, nacimientoRaw, numeroRaw string) (UserData, error) {
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
