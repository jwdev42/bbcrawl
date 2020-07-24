package download

func ADNameFromHeader(dl *Download) {
	name, err := dl.NameFromHeader()
	if err != nil {
		dl.Err = err
		return
	}
	if err := dl.Rename(name); err != nil {
		dl.Err = err
		return
	}
}
