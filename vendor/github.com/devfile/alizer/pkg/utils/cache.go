package utils

import "context"

type key string

func GetCachedFilePathsFromRoot(root string, ctx *context.Context) ([]string, error) {
	filePathsFromRoot := getMapFromContext(*ctx)
	if files, hasRoot := filePathsFromRoot[root]; hasRoot {
		return files, nil
	}

	filePaths, err := GetFilePathsFromRoot(root)
	if err != nil {
		return []string{}, err
	}
	filePathsFromRoot[root] = filePaths

	*ctx = context.WithValue(*ctx, key("mapFilePathsFromRoot"), filePathsFromRoot)
	return filePaths, nil
}

func getMapFromContext(ctx context.Context) map[string][]string {
	filePathsFromRoot := ctx.Value(key("mapFilePathsFromRoot"))
	if filePathsFromRoot != nil {
		return filePathsFromRoot.(map[string][]string)
	}
	return make(map[string][]string)
}
