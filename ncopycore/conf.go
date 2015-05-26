package ncopycore

type Conf struct {
	Src struct {
		Path string
	}
	Ignore struct {
		Files []string
	}
}
