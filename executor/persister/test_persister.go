package persister

// TestPersister is a persister that will save the output to a file
type TestPersister struct {
	TestName string
	Content  []string
}

// Persist TestPersister the data to a file with public permissions
func (f *TestPersister) Persist(data string) error {
	if len(f.Content) == 0 {
		f.Content = make([]string, 1)
		f.Content[0] = data
	} else {
		f.Content = append(f.Content, data)
	}
	return nil
}

// SetScriptName sets what name the output file has
func (f *TestPersister) SetScriptName(name string) error {
	f.TestName = name
	return nil
}
