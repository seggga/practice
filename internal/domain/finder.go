package domain

type FinderSettings struct {
	Depth      int
	Ignore     map[string]string
	Processors int
}

type Finder interface {
	Find(dir string, depth int, porc int) ([]File, error)
}
