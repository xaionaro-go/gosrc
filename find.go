package gosrc

func (files Files) FindByGoGenerateTag(goGenerateTag string) Files {
	var filteredFiles Files
	for _, file := range files {
		for _, tag := range file.GoGenerateTags() {
			if tag == goGenerateTag {
				filteredFiles = append(filteredFiles, file)
				break
			}
		}
	}
	return filteredFiles
}
