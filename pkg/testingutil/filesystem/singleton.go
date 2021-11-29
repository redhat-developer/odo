package filesystem

var singleFs Filesystem

func Get() Filesystem {
	if singleFs == nil {
		singleFs = &DefaultFs{}
	}
	return singleFs
}

func Set(fs Filesystem) {
	singleFs = fs
}
