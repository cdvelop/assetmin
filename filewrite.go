package assetmin

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// pathFile ej: "theme/index.html"
// data ej: *bytes.Buffer
// NOTA: la data del buf sera eliminada después de escribir el archivo
func FileWrite(pathFile string, data bytes.Buffer) error {
	const e = "FileWrite "

	// Asegurar que el directorio existe antes de crear el archivo
	dir := filepath.Dir(pathFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.New(e + "al crear directorio " + err.Error())
	}

	dst, err := os.Create(pathFile)
	if err != nil {
		return errors.New(e + "al crear archivo " + err.Error())
	}
	defer dst.Close()

	// fmt.Println("data antes de escribir:", data.String())
	// Copy the uploaded assetFile to the filesystem at the specified destination
	// _, e = io.Copy(dst, bytes.NewReader(data.Bytes()))
	_, err = io.Copy(dst, &data)
	if err != nil {
		return errors.New(e + "no se logro escribir el archivo " + pathFile + " en el destino " + err.Error())
	}
	// fmt.Println("data después de copy:", data.String(), "bytes:", data.Bytes())

	return nil
}
