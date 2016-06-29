package finder

// This is to sort search results by file name to prevent random failure.
type ByFileName []*SearchResult

func (a ByFileName) Len() int           { return len(a) }
func (a ByFileName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFileName) Less(i, j int) bool { return a[i].File < a[j].File }
