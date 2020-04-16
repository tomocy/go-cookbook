package batch

func batch(length, size int, fn func(int, int) error) error {
	var begin int
	for begin < length {
		end := begin + size - 1
		if end >= length {
			end = length - 1
		}

		if err := fn(begin, end); err != nil {
			return err
		}

		begin = end + 1
	}

	return nil
}
