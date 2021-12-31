package configwatchd

func Flush(opts FlushOpts) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	return c.flush(opts)
}

func List() ([]string, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}
	return c.list()
}
