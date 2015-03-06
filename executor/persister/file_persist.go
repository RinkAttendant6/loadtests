package persister

import (
	"fmt"
	"io/ioutil"
)

// FilePersister is a persister that will save the output to a file
type FilePersister struct {
	fileName string
}

// Persist saves the data to a file with public permissions
func (f *FilePersister) Persist(data string) error {
	if f.fileName == "" {
		return fmt.Errorf("The output file name has not been set")
	}

	byteData := []byte(data)
	err := ioutil.WriteFile(f.fileName, byteData, 0666)
	return err
}

// SetScriptName sets what name the output file has
func (f *FilePersister) SetScriptName(name string) error {
	f.fileName = fmt.Sprintf("%s.out", name)
	return nil
}
