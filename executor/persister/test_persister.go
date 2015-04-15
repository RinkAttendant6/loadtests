package persister

import "fmt"

//import "time"

// TestPersister is a persister that will save the output to a file
type TestPersister struct {
	Content []string
}

// Persist TestPersister the data to a file with public permissions
func (f *TestPersister) Persist(testName string, ip string, code []byte) error {
	//fmt.Printf("%v\n", time.Now().Local())
	if len(f.Content) == 0 {
		f.Content = make([]string, 1)
		f.Content[0] = fmt.Sprintf("%s: %s", ip, code)
	} else {
		f.Content = append(f.Content, fmt.Sprintf("%s: %s", ip, code))
	}
	return nil
}
